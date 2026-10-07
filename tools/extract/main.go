// Command extract reads a local Php Inspections (EA Extended) checkout and
// produces:
//
//   - internal/meta/rules.json — rule FACTS only (ids, groups, severities,
//     option names/defaults, fix availability). Committed.
//   - .cache/ea/index.json — a local-only index pointing into the EA checkout
//     (inspection sources, test cases, fixture paths) used by the spec writers
//     and the conformance runner. Never committed: it only references files,
//     it never copies them.
package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"custos/internal/meta"
)

const (
	pluginXML   = "src/main/resources/META-INF/plugin.xml"
	rulesMD     = "RULES.md"
	javaMain    = "src/main/java"
	javaTest    = "src/test/java"
	descrDir    = "src/main/resources/inspectionDescriptions"
	metaOut     = "internal/meta/rules.json"
	indexOut    = ".cache/ea/index.json"
	shortSuffix = "Inspection"
)

// RuleIndex points at the EA material for one rule (paths relative to the EA root).
type RuleIndex struct {
	LegacyID    string   `json:"legacyId"`
	Class       string   `json:"class"`
	Source      string   `json:"source"`
	Description string   `json:"description,omitempty"`
	Tests       []string `json:"tests,omitempty"`
}

// Case is one fixture-driven test extracted from an EA Java test method.
type Case struct {
	Test            string            `json:"test"` // file#method
	Rules           []string          `json:"rules"`
	Fixture         string            `json:"fixture"`
	Fixed           string            `json:"fixed,omitempty"`
	PHP             string            `json:"php,omitempty"`
	ComparisonStyle string            `json:"comparisonStyle,omitempty"`
	Options         map[string]string `json:"options,omitempty"` // "Rule.OPTION" -> raw Java value
	Calls           []string          `json:"calls,omitempty"`   // raw option-mutating calls, e.g. registerCustomDebugMethod("x")
	// Companions are files configured earlier in the same test method; they
	// stay in the IDE project, so their symbols are visible to this case.
	Companions []string `json:"companions,omitempty"`
}

// Index is the local-only EA index.
type Index struct {
	EAPath     string               `json:"eaPath"`
	Rules      map[string]RuleIndex `json:"rules"`
	Cases      []Case               `json:"cases"`
	Unresolved []string             `json:"unresolved,omitempty"`
}

func main() {
	ea := flag.String("ea", os.ExpandEnv("$HOME/Sites/phpinspectionsea"), "path to the EA checkout")
	flag.Parse()
	if err := run(*ea); err != nil {
		fmt.Fprintln(os.Stderr, "extract:", err)
		os.Exit(1)
	}
}

func run(ea string) error {
	ea, err := filepath.Abs(ea)
	if err != nil {
		return err
	}
	decls, err := parsePluginXML(filepath.Join(ea, pluginXML))
	if err != nil {
		return err
	}
	qf, err := parseRulesMD(filepath.Join(ea, rulesMD))
	if err != nil {
		return err
	}

	var rules []meta.Rule
	idx := Index{EAPath: ea, Rules: map[string]RuleIndex{}}
	classToID := map[string]string{}
	seen := map[string]bool{}
	for _, d := range decls {
		id := strings.TrimSuffix(d.ShortName, shortSuffix)
		if seen[id] {
			return fmt.Errorf("duplicate rule id %s", id)
		}
		seen[id] = true
		src := filepath.ToSlash(filepath.Join(javaMain, strings.ReplaceAll(d.Class, ".", "/")+".java"))
		body, err := os.ReadFile(filepath.Join(ea, src))
		if err != nil {
			return fmt.Errorf("%s: %w", d.ShortName, err)
		}
		// RULES.md is not always accurate: also trust fixes found in the source.
		hasFix := qf[d.ShortName] || fixRe.Match(body)
		sev, err := severity(d.Level)
		if err != nil {
			return fmt.Errorf("%s: %w", d.ShortName, err)
		}
		rules = append(rules, meta.Rule{
			ID:               id,
			LegacyID:         d.ShortName,
			Group:            d.Group,
			Severity:         sev,
			EnabledByDefault: d.Enabled == "true",
			HasFix:           hasFix,
			Experimental:     d.Experimental,
			Options:          parseOptions(string(body)),
		})
		ri := RuleIndex{LegacyID: d.ShortName, Class: d.Class, Source: src}
		descr := filepath.ToSlash(filepath.Join(descrDir, d.ShortName+".html"))
		if _, err := os.Stat(filepath.Join(ea, descr)); err == nil {
			ri.Description = descr
		}
		idx.Rules[id] = ri
		classToID[simpleName(d.Class)] = id
	}

	if err := parseTests(ea, classToID, &idx); err != nil {
		return err
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	if err := writeJSON(metaOut, rules); err != nil {
		return err
	}
	if err := writeJSON(indexOut, idx); err != nil {
		return err
	}

	fixed, withCases := 0, map[string]bool{}
	for _, c := range idx.Cases {
		if c.Fixed != "" {
			fixed++
		}
		for _, r := range c.Rules {
			withCases[r] = true
		}
	}
	fmt.Printf("rules: %d (with fix: %d)\n", len(rules), countFix(rules))
	fmt.Printf("cases: %d (with .fixed: %d), rules covered: %d\n", len(idx.Cases), fixed, len(withCases))
	for _, r := range rules {
		if !withCases[r.ID] {
			fmt.Printf("  no cases: %s\n", r.ID)
		}
	}
	for _, u := range idx.Unresolved {
		fmt.Printf("  unresolved: %s\n", u)
	}
	return nil
}

// ---- plugin.xml -------------------------------------------------------------

type inspectionDecl struct {
	ShortName, Group, Level, Enabled, Class string
	Experimental                            bool
}

func parsePluginXML(path string) ([]inspectionDecl, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	dec.Strict = false
	var out []inspectionDecl
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("plugin.xml: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "localInspection" {
			continue
		}
		var d inspectionDecl
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "shortName":
				d.ShortName = a.Value
			case "groupName":
				d.Group = a.Value
			case "level":
				d.Level = a.Value
			case "enabledByDefault":
				d.Enabled = a.Value
			case "implementationClass":
				d.Class = a.Value
			}
		}
		if d.ShortName == "" || d.Class == "" {
			return nil, fmt.Errorf("plugin.xml: incomplete localInspection %+v", d)
		}
		out = append(out, d)
	}
	// Inspections commented out upstream are still part of the catalogue,
	// flagged experimental and disabled by default.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for _, m := range commentRe.FindAllSubmatch(raw, -1) {
		for _, el := range declRe.FindAll(m[1], -1) {
			var x struct {
				ShortName string `xml:"shortName,attr"`
				Group     string `xml:"groupName,attr"`
				Level     string `xml:"level,attr"`
				Class     string `xml:"implementationClass,attr"`
			}
			if err := xml.Unmarshal(el, &x); err != nil || x.ShortName == "" {
				continue
			}
			out = append(out, inspectionDecl{ShortName: x.ShortName, Group: x.Group, Level: x.Level, Enabled: "false", Class: x.Class, Experimental: true})
		}
	}
	return out, nil
}

var (
	commentRe = regexp.MustCompile(`(?s)<!--(.*?)-->`)
	declRe    = regexp.MustCompile(`(?s)<localInspection\b.*?/>`)
)

func severity(level string) (meta.Severity, error) {
	switch level {
	case "ERROR":
		return meta.SeverityError, nil
	case "WARNING":
		return meta.SeverityWarning, nil
	case "WEAK WARNING", "INFO", "INFORMATION", "TYPO":
		return meta.SeverityInfo, nil
	}
	return "", fmt.Errorf("unknown level %q", level)
}

// ---- RULES.md ---------------------------------------------------------------

// parseRulesMD returns short name -> QF column is "yes".
func parseRulesMD(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		cols := strings.Split(sc.Text(), "|")
		if len(cols) < 6 {
			continue
		}
		short := strings.TrimSpace(cols[2])
		if !strings.HasSuffix(short, shortSuffix) {
			continue
		}
		out[short] = strings.TrimSpace(cols[4]) == "yes"
	}
	return out, sc.Err()
}

// ---- options ----------------------------------------------------------------

var fixRe = regexp.MustCompile(`LocalQuickFix|new \w+(?:Fix|Fixer)\(`)

var optionRe = regexp.MustCompile(`(?m)^\s*public\s+(?:final\s+)?(boolean|int|String|PhpUnitVersion|List<String>)\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*([^;]+);`)

func parseOptions(src string) []meta.Option {
	var out []meta.Option
	for _, m := range optionRe.FindAllStringSubmatch(src, -1) {
		typ, name, raw := m[1], m[2], strings.TrimSpace(m[3])
		o := meta.Option{Name: name}
		switch typ {
		case "boolean":
			o.Type, o.Default = "bool", raw == "true"
		case "int":
			o.Type = "int"
			if n, err := strconv.Atoi(raw); err == nil {
				o.Default = n
			}
		case "String":
			o.Type, o.Default = "string", strings.Trim(raw, `"`)
		case "PhpUnitVersion":
			o.Type, o.Default = "enum", strings.TrimPrefix(raw, "PhpUnitVersion.")
		case "List<String>":
			o.Type = "list"
		}
		out = append(out, o)
	}
	return out
}

// ---- tests ------------------------------------------------------------------

var (
	testMethodRe  = regexp.MustCompile(`public\s+void\s+(test\w*)\s*\(\s*\)`)
	newVarRe      = regexp.MustCompile(`^(?:final\s+)?(\w+)\s+(\w+)\s*=\s*new\s+(\w+)\s*\(\s*\)$`)
	assignRe      = regexp.MustCompile(`^(\w+)\.(\w+)\s*=\s*(.+)$`)
	callRe        = regexp.MustCompile(`^(\w+)\.(\w+(?:\.\w+)*)\s*\((.*)\)$`)
	enableRe      = regexp.MustCompile(`enableInspections\((.*)\)$`)
	newInlineRe   = regexp.MustCompile(`^new\s+(\w+)\s*\(\s*\)$`)
	levelConstRe  = regexp.MustCompile(`PhpLanguageLevel\.PHP(\d)(\d)\d`)
	levelParseRe  = regexp.MustCompile(`^(?:final\s+)?PhpLanguageLevel\s+(\w+)\s*=\s*PhpLanguageLevel\.parse\("([^"]+)"\)$`)
	setLevelVarRe = regexp.MustCompile(`setLanguageLevel\((\w+)\)$`)
	styleRe       = regexp.MustCompile(`ComparisonStyle\.force\(ComparisonStyle\.(\w+)\)`)
	configureRe   = regexp.MustCompile(`configureByFile\(\s*"([^"]+)"\s*\)`)
	// checkResultByFile("expected") or checkResultByFile("actual", "expected", ignoreWS).
	checkResultRe = regexp.MustCompile(`checkResultByFile\(\s*(?:"[^"]+"\s*,\s*)?"([^"]+)"`)
	lineCommentRe = regexp.MustCompile(`(?m)//[^\n]*$`)
	blockCommRe   = regexp.MustCompile(`(?s)/\*.*?\*/`)
)

func parseTests(ea string, classToID map[string]string, idx *Index) error {
	root := filepath.Join(ea, javaTest)
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".java") {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(ea, path)
		rel = filepath.ToSlash(rel)
		src := blockCommRe.ReplaceAllString(string(body), "")
		src = lineCommentRe.ReplaceAllString(src, "")
		helpers := parseHelpers(src, classToID)
		locs := testMethodRe.FindAllStringSubmatchIndex(src, -1)
		touched := map[string]bool{}
		for i, loc := range locs {
			end := len(src)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			name := src[loc[2]:loc[3]]
			for _, c := range parseMethod(rel+"#"+name, src[loc[1]:end], classToID, helpers, idx) {
				idx.Cases = append(idx.Cases, c)
				for _, r := range c.Rules {
					touched[r] = true
				}
			}
		}
		for id := range touched {
			ri := idx.Rules[id]
			ri.Tests = appendUnique(ri.Tests, rel)
			idx.Rules[id] = ri
		}
		return nil
	})
}

func parseMethod(test, body string, classToID map[string]string, helpers map[string]helper, idx *Index) []Case {
	vars := map[string]string{}      // java var -> rule id
	levelVars := map[string]string{} // java var -> php version
	options := map[string]string{}
	var calls, active []string
	// PhpStorm project level (facade) and EA's own override (PhpLanguageLevel.set).
	facade, override, style, pending := "", "", "", ""
	php := func() string {
		if override != "" {
			return override
		}
		return facade
	}
	var cases []Case
	emit := func() {
		if pending == "" {
			return
		}
		var companions []string
		for _, prev := range cases {
			companions = append(companions, prev.Fixture)
		}
		c := Case{
			Test:            test,
			Rules:           append([]string(nil), active...),
			Companions:      companions,
			Fixture:         pending,
			PHP:             php(),
			ComparisonStyle: style,
			Calls:           append([]string(nil), calls...),
		}
		if len(options) > 0 {
			c.Options = map[string]string{}
			for k, v := range options {
				c.Options[k] = v
			}
		}
		cases = append(cases, c)
		pending = ""
	}

	for _, stmt := range strings.Split(body, ";") {
		s := strings.Join(strings.Fields(stmt), " ")
		s = strings.TrimLeft(s, "{} ")
		if s == "" {
			continue
		}
		// Statements may be prefixed by control-flow (`if (...) {`); keep the tail.
		if i := strings.LastIndex(s, "{ "); i >= 0 && !strings.Contains(s[i:], "}") {
			s = strings.TrimSpace(s[i+1:])
		}
		switch {
		case newVarRe.MatchString(s):
			m := newVarRe.FindStringSubmatch(s)
			if id, ok := classToID[m[3]]; ok {
				vars[m[2]] = id
			}
		case levelParseRe.MatchString(s):
			m := levelParseRe.FindStringSubmatch(s)
			levelVars[m[1]] = m[2]
		case enableRe.MatchString(s):
			active = active[:0:0]
			for _, arg := range splitArgs(enableRe.FindStringSubmatch(s)[1]) {
				if m := newInlineRe.FindStringSubmatch(arg); m != nil {
					if id, ok := classToID[m[1]]; ok {
						active = append(active, id)
					} else {
						idx.Unresolved = append(idx.Unresolved, test+": "+arg)
					}
				} else if id, ok := vars[arg]; ok {
					active = append(active, id)
				} else if h, ok := helpers[strings.TrimSuffix(arg, "()")]; ok && strings.HasSuffix(arg, "()") {
					active = append(active, h.id)
					for k, v := range h.options {
						options[k] = v
					}
					calls = append(calls, h.calls...)
				} else {
					idx.Unresolved = append(idx.Unresolved, test+": "+arg)
				}
			}
		case configureRe.MatchString(s):
			// Some tests configure the fixture before enabling the inspection:
			// the case is snapshotted at testHighlighting (or method end).
			emit()
			pending = configureRe.FindStringSubmatch(s)[1]
		case strings.Contains(s, "testHighlighting("):
			emit()
		case checkResultRe.MatchString(s):
			emit()
			if len(cases) > 0 {
				cases[len(cases)-1].Fixed = checkResultRe.FindStringSubmatch(s)[1]
			}
		case strings.Contains(s, "PhpLanguageLevel.set("):
			override = ""
			if m := levelConstRe.FindStringSubmatch(s); m != nil {
				override = m[1] + "." + m[2]
			}
		case strings.Contains(s, "setLanguageLevel("):
			if m := levelConstRe.FindStringSubmatch(s); m != nil {
				facade = m[1] + "." + m[2]
			} else if m := setLevelVarRe.FindStringSubmatch(s); m != nil && levelVars[m[1]] != "" {
				facade = levelVars[m[1]]
			}
		case styleRe.MatchString(s):
			style = strings.ToLower(styleRe.FindStringSubmatch(s)[1])
		case assignRe.MatchString(s):
			m := assignRe.FindStringSubmatch(s)
			if id, ok := vars[m[1]]; ok {
				options[id+"."+m[2]] = strings.TrimSpace(m[3])
			}
		case callRe.MatchString(s):
			m := callRe.FindStringSubmatch(s)
			if id, ok := vars[m[1]]; ok {
				calls = append(calls, id+"."+m[2]+"("+m[3]+")")
			}
		}
	}
	emit()
	return cases
}

// helper is a test-class factory method returning a configured inspection.
type helper struct {
	id      string
	options map[string]string
	calls   []string
}

var helperRe = regexp.MustCompile(`(?s)(?:private|protected|public)?\s*(?:static\s+)?(\w+)\s+(\w+)\s*\(\s*\)\s*\{(.*?)\breturn\s+\w+\s*;`)

func parseHelpers(src string, classToID map[string]string) map[string]helper {
	out := map[string]helper{}
	for _, m := range helperRe.FindAllStringSubmatch(src, -1) {
		id, ok := classToID[m[1]]
		if !ok {
			continue
		}
		h := helper{id: id, options: map[string]string{}}
		vars := map[string]bool{}
		for _, stmt := range strings.Split(m[3], ";") {
			s := strings.Join(strings.Fields(stmt), " ")
			if v := newVarRe.FindStringSubmatch(s); v != nil {
				vars[v[2]] = true
			} else if a := assignRe.FindStringSubmatch(s); a != nil && vars[a[1]] {
				h.options[id+"."+a[2]] = strings.TrimSpace(a[3])
			} else if c := callRe.FindStringSubmatch(s); c != nil && vars[c[1]] {
				h.calls = append(h.calls, id+"."+c[2]+"("+c[3]+")")
			}
		}
		out[m[2]] = h
	}
	return out
}

// splitArgs splits a Java argument list on top-level commas.
func splitArgs(s string) []string {
	var out []string
	depth, start := 0, 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if tail := strings.TrimSpace(s[start:]); tail != "" {
		out = append(out, tail)
	}
	return out
}

// ---- helpers ----------------------------------------------------------------

func simpleName(fqcn string) string {
	return fqcn[strings.LastIndex(fqcn, ".")+1:]
}

func appendUnique(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

func countFix(rules []meta.Rule) int {
	n := 0
	for _, r := range rules {
		if r.HasFix {
			n++
		}
	}
	return n
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
