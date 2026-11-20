// Command architecture checks package ownership and inspection completeness.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/tools/internal/specmd"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("architecture", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	problems, err := check(*root)
	if err != nil {
		fmt.Fprintln(stderr, "architecture:", err)
		return 2
	}
	for _, problem := range problems {
		fmt.Fprintln(stdout, problem)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Fprintln(stdout, "architecture: package boundaries and inspection catalogue verified")
	return 0
}

func check(root string) ([]string, error) {
	expected := map[string]bool{}
	native := map[string]bool{}
	for _, name := range []string{"rules.json", "native-rules.json"} {
		raw, err := os.ReadFile(filepath.Join(root, "internal/inspection/meta", name))
		if err != nil {
			return nil, err
		}
		var facts []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &facts); err != nil {
			return nil, err
		}
		for _, fact := range facts {
			if expected[fact.ID] {
				return nil, fmt.Errorf("duplicate catalogue ID %s", fact.ID)
			}
			expected[fact.ID] = true
			native[fact.ID] = name == "native-rules.json"
		}
	}
	var problems []string
	report := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }
	ids := map[string]string{}
	constructors := map[string]bool{}
	ruleDirs := map[string]bool{}
	listed := map[string]int{}
	packageDocs := map[string][]string{}
	// Marker methods (Semantic, Flow) by "dir.receiver", and the receiver
	// that declares each rule directory's ID.
	markers := map[string]map[string]bool{}
	idRecv := map[string]string{}
	// Exported semanticquery declarations and the packages that use them.
	shared := map[string]string{}
	users := map[string]map[string]bool{}
	use := func(name, user string) {
		if users[name] == nil {
			users[name] = map[string]bool{}
		}
		users[name][user] = true
	}
	for _, base := range []string{"internal", "cmd", "pkg"} {
		if _, err := os.Stat(filepath.Join(root, base)); base == "pkg" && errors.Is(err, fs.ErrNotExist) {
			continue
		}
		err := filepath.WalkDir(filepath.Join(root, base), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			from := filepath.ToSlash(filepath.Dir(rel))
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			imports := map[string]string{}
			for _, im := range file.Imports {
				target, _ := strconv.Unquote(im.Path.Value)
				alias := filepath.Base(target)
				if im.Name != nil {
					alias = im.Name.Name
				}
				imports[alias] = target
				if analysisLayer(from) && systemPackage(target) {
					report("%s: analysis code cannot use %s; adapters own system access", rel, target)
				}
				target = strings.TrimPrefix(target, "github.com/janalis/custos/")
				if (strings.HasPrefix(target, "internal/") || strings.HasPrefix(target, "pkg/")) && !allowed(from, target) {
					report("%s: forbidden dependency on %s", rel, target)
				}
			}
			if sharedPackages[from] {
				declared := map[*ast.Ident]bool{}
				for _, decl := range file.Decls {
					switch decl := decl.(type) {
					case *ast.FuncDecl:
						if decl.Recv == nil && decl.Name.IsExported() {
							shared[from+"."+decl.Name.Name] = rel
							declared[decl.Name] = true
						}
					case *ast.GenDecl:
						for _, spec := range decl.Specs {
							var names []*ast.Ident
							switch spec := spec.(type) {
							case *ast.TypeSpec:
								names = []*ast.Ident{spec.Name}
							case *ast.ValueSpec:
								names = spec.Names
							}
							for _, name := range names {
								if name.IsExported() {
									shared[from+"."+name.Name] = rel
									declared[name] = true
								}
							}
						}
					}
				}
				ast.Inspect(file, func(node ast.Node) bool {
					if id, ok := node.(*ast.Ident); ok && !declared[id] && id.IsExported() {
						use(from+"."+id.Name, from)
					}
					return true
				})
			}
			ast.Inspect(file, func(node ast.Node) bool {
				if sel, ok := node.(*ast.SelectorExpr); ok {
					if x, ok := sel.X.(*ast.Ident); ok && sharedPackages[strings.TrimPrefix(imports[x.Name], "github.com/janalis/custos/")] {
						use(strings.TrimPrefix(imports[x.Name], "github.com/janalis/custos/")+"."+sel.Sel.Name, from)
					}
				}
				if strings.HasPrefix(from, "internal/inspection/") && versionOrdering(node, imports) {
					report("%s:%d: compare PHP versions with AtLeast or Below", rel, fset.Position(node.Pos()).Line)
				}
				return true
			})
			isRule := strings.HasPrefix(from, "internal/inspection/rules/")
			if isRule {
				ruleDirs[from] = true
				if file.Doc != nil {
					packageDocs[from] = append(packageDocs[from], strings.Join(strings.Fields(file.Doc.Text()), " "))
				}
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if isRule {
					if fn.Name.Name == "init" {
						report("%s: inspections cannot register through init", rel)
					}
					if constructor(fn, imports) {
						constructors[from] = true
					}
					if recv := receiver(fn); recv != "" && (fn.Name.Name == "Semantic" || fn.Name.Name == "Flow") {
						if markers[from+"."+recv] == nil {
							markers[from+"."+recv] = map[string]bool{}
						}
						markers[from+"."+recv][fn.Name.Name] = true
					}
					if fn.Name.Name == "ID" && fn.Recv != nil && fn.Body != nil {
						ast.Inspect(fn.Body, func(node ast.Node) bool {
							ret, ok := node.(*ast.ReturnStmt)
							if !ok || len(ret.Results) != 1 {
								return true
							}
							lit, ok := ret.Results[0].(*ast.BasicLit)
							if !ok || lit.Kind != token.STRING {
								return true
							}
							id, _ := strconv.Unquote(lit.Value)
							if previous := ids[id]; previous != "" {
								report("%s: duplicate inspection ID %s (also %s)", rel, id, previous)
							}
							ids[id] = from
							idRecv[from] = receiver(fn)
							if filepath.Base(from) != strings.ToLower(id) {
								report("%s: directory must be lowercase ID %s", rel, id)
							}
							if !expected[id] {
								report("%s: unknown inspection ID %s", rel, id)
							}
							return true
						})
					}
				}
				if from == "internal/inspection/catalogue" && fn.Name.Name == "All" {
					ast.Inspect(fn.Body, func(node ast.Node) bool {
						call, ok := node.(*ast.CallExpr)
						if !ok {
							return true
						}
						sel, ok := call.Fun.(*ast.SelectorExpr)
						if !ok || sel.Sel.Name != "New" {
							return true
						}
						name, ok := sel.X.(*ast.Ident)
						if !ok {
							return true
						}
						target := strings.TrimPrefix(imports[name.Name], "github.com/janalis/custos/")
						if strings.HasPrefix(target, "internal/inspection/rules/") {
							listed[target]++
						}
						return true
					})
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for id := range expected {
		owner := ids[id]
		if owner == "" {
			report("%s: inspection implementation missing", id)
			continue
		}
		if !constructors[owner] {
			report("%s: New constructor missing", id)
		}
		if listed[owner] != 1 {
			report("%s: catalogue must construct inspection exactly once (got %d)", id, listed[owner])
		}
		spec, err := os.ReadFile(filepath.Join(root, "specs", id+".md"))
		if err != nil {
			report("%s: specification missing", id)
		} else {
			for _, problem := range specProblems(string(spec), ruleKind(markers[owner+"."+idRecv[owner]])) {
				report("%s: %s", id, problem)
			}
		}
		if m := specMechanics.Find(spec); m != nil {
			report("%s: specification describes behavior only; move %q to the rule's NOTES.md", id, m)
		}
		checkPackageDoc(report, id, owner, native[id], packageDocs[owner])
		fixtures, err := os.ReadDir(filepath.Join(root, owner, "testdata"))
		hasPHP := false
		for _, fixture := range fixtures {
			name := fixture.Name()
			if fixture.IsDir() {
				name += "/composer.config.json"
			} else if strings.HasSuffix(name, ".php") {
				hasPHP = true
			}
			checkFixtureName(report, owner, name)
			if strings.HasSuffix(name, ".json") {
				checkSidecar(report, filepath.Join(root, owner, "testdata"), owner, name)
			}
		}
		if err != nil || !hasPHP {
			report("%s: own PHP fixtures missing", id)
		}
	}
	for owner := range ruleDirs {
		found := false
		for _, dir := range ids {
			if dir == owner {
				found = true
			}
		}
		if !found {
			report("%s: rule directory has no inspection ID", owner)
		}
	}
	for owner := range constructors {
		if listed[owner] != 1 {
			report("%s: constructor is absent or repeated in catalogue", owner)
		}
	}
	for owner := range listed {
		if !constructors[owner] {
			report("%s: catalogue references missing constructor", owner)
		}
	}
	for name, rel := range shared {
		if len(users[name]) == 1 {
			for user := range users[name] {
				if strings.HasPrefix(user, "internal/inspection/rules/") {
					report("%s: %s is used only by %s; keep it in that rule", rel, name[strings.LastIndexByte(name, '.')+1:], user)
				}
			}
		}
	}
	sort.Strings(problems)
	return problems, nil
}

// checkPackageDoc requires every package comment of a rule package to
// open with the sentence naming its inspection, and at least one to exist.
func checkPackageDoc(report func(string, ...any), id, owner string, native bool, docs []string) {
	kind := ""
	if native {
		kind = "native "
	}
	want := fmt.Sprintf("Package %s implements the %s%s inspection.", filepath.Base(owner), kind, id)
	if len(docs) == 0 {
		report("%s: package comment missing; start it with %q", owner, want)
	}
	for _, doc := range docs {
		if doc != want && !strings.HasPrefix(doc, want+" ") {
			report("%s: package comment must start with %q", owner, want)
		}
	}
}

// retiredFixtureNames are fixture names replaced by the documented scheme:
// findings in basic.php, code that must not be reported in
// false-positives*.php.
var retiredFixtureNames = map[string]string{
	"positive": "basic",
	"negative": "false-positives",
	"valid":    "false-positives",
}

func checkFixtureName(report func(string, ...any), owner, name string) {
	base := name
	for _, suffix := range []string{".fixed.php", ".php", ".json"} {
		if strings.HasSuffix(base, suffix) {
			base = strings.TrimSuffix(base, suffix)
			break
		}
	}
	if use, ok := retiredFixtureNames[base]; ok {
		report("%s/testdata/%s: rename the fixture to %s", owner, name, use)
	}
}

// checkSidecar rejects fixture settings that only restate the defaults: an
// empty sidecar, empty options or the default PHP version.
func checkSidecar(report func(string, ...any), dir, owner, name string) {
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return // a manifest case without settings
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(raw, &cfg); err != nil {
		report("%s/testdata/%s: invalid sidecar: %v", owner, name, err)
		return
	}
	if len(cfg) == 0 {
		report("%s/testdata/%s: empty sidecar; delete it", owner, name)
	}
	if options, ok := cfg["options"]; ok {
		var set map[string]any
		if json.Unmarshal(options, &set) == nil && len(set) == 0 {
			report("%s/testdata/%s: empty options; remove the key", owner, name)
		}
	}
	if php, ok := cfg["php"]; ok {
		var version string
		if json.Unmarshal(php, &version) == nil {
			if v, err := phpversion.Parse(version); err == nil && v == phpversion.Default {
				report("%s/testdata/%s: php %q is the default; remove the key", owner, name, version)
			}
		}
	}
}

// ruleKind is the spec front matter kind a rule's marker methods call for:
// flow when it needs project-local callable summaries, semantic when it
// needs the project symbol index, syntax otherwise.
func ruleKind(marker map[string]bool) string {
	switch {
	case marker["Flow"]:
		return "flow"
	case marker["Semantic"]:
		return "semantic"
	}
	return "syntax"
}

var (
	specKind = regexp.MustCompile(`(?m)^kind:[ \t]*(\S*)`)
	// optionLabel is a Detection, Exceptions or Fix item label: rule pages
	// and `custos explain` show options without them.
	optionLabel = regexp.MustCompile(`\b[DEF]\d+[a-z]?\b`)
	tableRule   = regexp.MustCompile(`^\|[\s:|-]*\|$`)
)

// specProblems checks a spec against its rule's code and the authoring
// guide (docs/contributing/adding-a-rule.md): its kind matches the rule's
// marker methods, and its Options section, which users read, neither
// refers to D/E/F labels nor leaves a table without rows.
func specProblems(spec, kind string) []string {
	var problems []string
	got := ""
	if m := specKind.FindStringSubmatch(spec); m != nil {
		got = m[1]
	}
	if got != kind {
		problems = append(problems, fmt.Sprintf("specification kind %q, code calls for %q", got, kind))
	}
	options := specmd.Section(spec, "Options")
	if label := optionLabel.FindString(options); label != "" {
		problems = append(problems, fmt.Sprintf("specification Options refers to %s; describe the behavior in words", label))
	}
	table, rows := false, 0
	for _, line := range strings.Split(options, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case tableRule.MatchString(line):
			table = true
		case table && strings.HasPrefix(line, "|"):
			rows++
		}
	}
	if table && rows == 0 {
		problems = append(problems, "specification Options has an empty table; write None.")
	}
	return problems
}

// versionOrdering reports an ordering comparison (<, <=, >, >=) with a
// constant of the PHP version package: inspections write
// v.AtLeast(phpversion.PHP80) or v.Below(phpversion.PHP80) instead, one
// spelling for every version gate.
func versionOrdering(node ast.Node, imports map[string]string) bool {
	bin, ok := node.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	switch bin.Op {
	case token.LSS, token.LEQ, token.GTR, token.GEQ:
	default:
		return false
	}
	for _, side := range []ast.Expr{bin.X, bin.Y} {
		if sel, ok := side.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && imports[x.Name] == "github.com/janalis/custos/internal/php/version" {
				return true
			}
		}
	}
	return false
}

// receiver returns the type name of a method's receiver, or "".
func receiver(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return ""
	}
	typ := fn.Recv.List[0].Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	if id, ok := typ.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// sharedPackages hold the queries several rules share (syntax, flow and
// semantic); a query only one rule uses belongs in that rule.
var sharedPackages = map[string]bool{
	"internal/inspection/astquery":      true,
	"internal/inspection/flowquery":     true,
	"internal/inspection/semanticquery": true,
}

// specMechanics matches implementation detail a spec must not carry
// (docs/contributing/adding-a-rule.md): analysis budgets, repository paths
// (a glob such as **/testdata/** in an option default is not one) and Go
// source files. A rule about a budget of its own (a configured limit)
// names it otherwise.
var specMechanics = regexp.MustCompile(`(?i)\banalysis[- ]budgets?\b|\bbudget exhaustion\b|\b(?:internal|cmd|tools)/[a-z]|(?:^|[\s\x60(])testdata/|\b[a-z_]+\.go\b`)

func allowed(from, to string) bool {
	within := func(prefix string) bool { return to == prefix || strings.HasPrefix(to, prefix+"/") }
	if within("pkg") && !strings.HasPrefix(from, "pkg/") {
		return false // the public API wraps the internals, never the reverse
	}
	switch {
	case from == "pkg/custos/ast":
		return within("internal/php/syntax")
	case strings.HasPrefix(from, "pkg/"):
		return within("internal/cli") || within("internal/inspection/analysis") || within("internal/inspection/meta") || within("internal/inspection/pattern") || within("internal/diagnostic") || within("internal/php") || within("pkg/custos/ast")
	case strings.HasPrefix(from, "internal/php/"):
		return within("internal/php")
	case from == "internal/format":
		return within("internal/php")
	case from == "internal/format/editorconfig":
		return false // standard library only
	case strings.HasPrefix(from, "internal/semantic/"):
		return within("internal/php") || within("internal/semantic")
	case from == "internal/diagnostic":
		return within("internal/php")
	case from == "internal/diagnostic/markup":
		return within("internal/diagnostic") && to != "internal/diagnostic/markup"
	case from == "internal/fixing":
		return within("internal/php") || within("internal/diagnostic")
	case from == "internal/inspection/meta":
		return within("internal/diagnostic")
	case from == "internal/inspection/analysis":
		return within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/meta")
	case from == "internal/inspection/astquery":
		return within("internal/php") || within("internal/diagnostic")
	case from == "internal/inspection/flowquery":
		return within("internal/php") || within("internal/inspection/astquery")
	case strings.HasPrefix(from, "internal/inspection/rules/"):
		return within("internal/php") || within("internal/format") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/analysis") || within("internal/inspection/meta") || within("internal/inspection/astquery") || within("internal/inspection/flowquery") || within("internal/inspection/semanticquery") || within("internal/inspection/phpunit")
	case from == "internal/inspection/pattern":
		// Custom rules resolve names and types as the rules do.
		return within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/analysis") || within("internal/inspection/meta") || within("internal/inspection/astquery") || within("internal/inspection/semanticquery")
	case from == "internal/inspection/catalogue":
		return within("internal/inspection/analysis") || within("internal/inspection/rules")
	case strings.HasPrefix(from, "internal/inspection/"):
		return within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/inspection/analysis") || within("internal/inspection/meta") || within("internal/inspection/astquery") || within("internal/inspection/flowquery")
	case from == "internal/editor/mcp":
		return within("internal/platform/safeio") || within("internal/format") || within("internal/php") || within("internal/semantic") || within("internal/diagnostic") || within("internal/fixing") || within("internal/project") || within("internal/output/diff") || within("internal/inspection/analysis") || within("internal/inspection/meta")
	case strings.HasPrefix(from, "internal/project"):
		return !within("internal/cli") && !within("internal/editor") && !within("internal/output")
	case strings.HasPrefix(from, "internal/output/"):
		return within("internal/diagnostic") || within("internal/php") || within("internal/inspection/meta")
	case strings.HasPrefix(from, "internal/platform/"):
		return within("internal/platform")
	case from == "internal/editor/lsp":
		return within("internal/php") || within("internal/format") || within("internal/semantic") || within("internal/diagnostic") || within("internal/fixing") || within("internal/project") || within("internal/inspection/analysis") || within("internal/inspection/catalogue") || within("internal/inspection/meta")
	case from == "internal/cli":
		return !within("internal/testing")
	case strings.HasPrefix(from, "internal/testing/"):
		return !within("internal/cli") && !within("internal/editor")
	case strings.HasPrefix(from, "cmd/"):
		return within("internal/cli")
	default:
		// A new package states what it may use.
		return false
	}
}

// analysisLayer reports packages that turn source text into findings and
// edits: they compute, and adapters supply whatever comes from the system.
func analysisLayer(from string) bool {
	for _, prefix := range []string{"internal/php", "internal/semantic", "internal/inspection", "internal/diagnostic", "internal/fixing", "internal/format"} {
		if from == prefix || strings.HasPrefix(from, prefix+"/") {
			return true
		}
	}
	return false
}

// systemPackage reports standard packages reaching the filesystem, the
// network, processes or the environment.
func systemPackage(path string) bool {
	switch path {
	case "os", "syscall", "plugin", "net", "net/http", "net/rpc", "net/smtp":
		return true
	}
	return strings.HasPrefix(path, "os/") || strings.HasPrefix(path, "net/http/") || strings.HasPrefix(path, "net/rpc/")
}

func constructor(fn *ast.FuncDecl, imports map[string]string) bool {
	if fn.Name.Name != "New" || fn.Recv != nil {
		return false
	}
	if len(fn.Type.Params.List) != 0 || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	result, ok := fn.Type.Results.List[0].Type.(*ast.SelectorExpr)
	if !ok || result.Sel.Name != "Rule" {
		return false
	}
	qualifier, ok := result.X.(*ast.Ident)
	return ok && imports[qualifier.Name] == "github.com/janalis/custos/internal/inspection/analysis"
}
