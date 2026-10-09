package conformance

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

var ruleFilter = flag.String("rule", "", "only run cases for this rule ID")

// engine is the analysis engine under test; nil until the engine exists.
var engine Engine

const (
	eaIndexPath = "../../../.cache/ea/index.json"
	eaResults   = "../../../.cache/ea/conformance.json"
	// divergencesPath lists EA cases where custos intentionally differs.
	divergencesPath = "../../../testdata/ea-divergences.json"
	ownRoot         = "../../../testdata/rules"
)

func wanted(rules []string) bool {
	if *ruleFilter == "" {
		return true
	}
	for _, r := range rules {
		if r == *ruleFilter {
			return true
		}
	}
	return false
}

// TestEA checks every EA fixture case from the local checkout. Rules the
// engine does not implement yet are reported as pending, not failed.
func TestEA(t *testing.T) {
	idx, err := LoadEAIndex(eaIndexPath)
	if err != nil {
		t.Skipf("no EA index (run `make extract`): %v", err)
	}
	if _, err := os.Stat(idx.EAPath); err != nil {
		t.Skipf("EA checkout not found at %s", idx.EAPath)
	}

	divergences := map[string]string{}
	if b, err := os.ReadFile(divergencesPath); err == nil {
		_ = json.Unmarshal(b, &divergences)
	}
	status := map[string]string{} // rule -> pass | fail | pending | diverged
	mark := func(rules []string, s string) {
		for _, r := range rules {
			if status[r] == "fail" || (status[r] == "pass" && s == "pending") {
				continue
			}
			status[r] = s
		}
	}
	for _, c := range idx.Cases {
		if !wanted(c.Rules) {
			continue
		}
		t.Run(c.Test+"/"+filepath.Base(c.Fixture), func(t *testing.T) {
			marked, err := ReadEA(idx, c.Fixture)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := ParseMarkup(marked); err != nil {
				t.Fatalf("markup: %v", err)
			}
			if engine == nil || !engine.Supports(c.Rules) {
				mark(c.Rules, "pending")
				t.Skip("pending: rule not implemented")
			}
			var fixed []byte
			if c.Fixed != "" {
				if fixed, err = ReadEA(idx, c.Fixed); err != nil {
					t.Fatal(err)
				}
			}
			req := Request{Path: c.Fixture, Rules: c.Rules, PHP: c.PHP, ComparisonStyle: c.ComparisonStyle, Options: c.Options, Calls: c.Calls}
			for _, comp := range c.Companions {
				raw, err := ReadEA(idx, comp)
				if err != nil {
					t.Fatal(err)
				}
				clean, _, err := ParseMarkup(raw)
				if err != nil {
					t.Fatal(err)
				}
				if req.Companions == nil {
					req.Companions = map[string][]byte{}
				}
				req.Companions[comp] = clean
			}
			res, err := Check(engine, req, marked, fixed, CompareOptions{OnlyPrefixed: "[EA]", HideExpected: true})
			if err != nil {
				t.Fatal(err)
			}
			name := c.Test + "/" + filepath.Base(c.Fixture)
			if reason, ok := divergences[name]; ok {
				if res.OK() {
					t.Errorf("listed as divergence but passes now; remove it from %s", divergencesPath)
					return
				}
				mark(c.Rules, "pass")
				t.Skipf("documented divergence: %s", reason)
			}
			if !res.OK() {
				mark(c.Rules, "fail")
				reportResultHidden(t, res)
				return
			}
			mark(c.Rules, "pass")
		})
	}
	if *ruleFilter == "" {
		writeStatus(t, status)
	}
}

// fixtureConfig is the optional sidecar `<fixture>.json` of an own fixture.
type fixtureConfig struct {
	PHP             string         `json:"php"`
	ComparisonStyle string         `json:"comparisonStyle"`
	Options         map[string]any `json:"options"` // OPTION -> value (rule prefix implied); bool/number/string
	Calls           []string       `json:"calls"`   // list-option calls, e.g. `registerX("v")` (rule prefix implied)
	// Companions are extra project files (paths relative to the rule
	// directory, plain PHP, conventionally *.inc so they are not fixtures
	// themselves) whose symbols are indexed next to the fixture.
	Companions []string `json:"companions"`
}

// TestOwnFixtures checks testdata/rules/<RuleID>/*.php (with optional
// *.fixed.php and *.json sidecars). Messages are compared exactly.
func TestOwnFixtures(t *testing.T) {
	dirs, err := os.ReadDir(ownRoot)
	if err != nil {
		t.Skip("no own fixtures yet")
	}
	for _, d := range dirs {
		if !d.IsDir() || !wanted([]string{d.Name()}) {
			continue
		}
		rule := d.Name()
		files, _ := filepath.Glob(filepath.Join(ownRoot, rule, "*.php"))
		// Manifest fixtures for non-PHP rules: <case>/composer.json with an
		// optional composer.fixed.json and composer.config.json sidecar.
		manifests, _ := filepath.Glob(filepath.Join(ownRoot, rule, "*", "composer.json"))
		files = append(files, manifests...)
		for _, f := range files {
			if strings.HasSuffix(f, ".fixed.php") {
				continue
			}
			ext := filepath.Ext(f)
			name := filepath.Base(f)
			cfgSuffix := ".json"
			if ext != ".php" {
				name = filepath.Base(filepath.Dir(f)) + "/" + name
				cfgSuffix = ".config.json"
			}
			t.Run(rule+"/"+name, func(t *testing.T) {
				if engine == nil || !engine.Supports([]string{rule}) {
					t.Fatalf("fixtures exist for %s but the rule is not registered", rule)
				}
				marked, err := os.ReadFile(f)
				if err != nil {
					t.Fatal(err)
				}
				base := strings.TrimSuffix(f, ext)
				var cfg fixtureConfig
				if b, err := os.ReadFile(base + cfgSuffix); err == nil {
					if err := json.Unmarshal(b, &cfg); err != nil {
						t.Fatal(err)
					}
				}
				opts := map[string]string{}
				for k, v := range cfg.Options {
					if list, ok := v.([]any); ok { // list option: passed as JSON
						b, _ := json.Marshal(list)
						opts[rule+"."+k] = string(b)
						continue
					}
					opts[rule+"."+k] = fmt.Sprint(v)
				}
				var calls []string
				for _, c := range cfg.Calls {
					calls = append(calls, rule+"."+c)
				}
				fixed, _ := os.ReadFile(base + ".fixed" + ext)
				req := Request{Path: f, Rules: []string{rule}, PHP: cfg.PHP, ComparisonStyle: cfg.ComparisonStyle, Options: opts, Calls: calls}
				for _, comp := range cfg.Companions {
					p := filepath.Join(ownRoot, rule, comp)
					src, err := os.ReadFile(p)
					if err != nil {
						t.Fatal(err)
					}
					if req.Companions == nil {
						req.Companions = map[string][]byte{}
					}
					req.Companions[p] = src
				}
				res, err := Check(engine, req, marked, fixed, CompareOptions{Messages: true})
				if err != nil {
					t.Fatal(err)
				}
				reportResult(t, res)
			})
		}
	}
}

func reportResult(t *testing.T, res Result) {
	t.Helper()
	for _, m := range res.Missing {
		t.Errorf("missing: %s", m)
	}
	for _, u := range res.Unexpected {
		t.Errorf("unexpected: %s", u)
	}
	if res.FixDiff != "" {
		t.Errorf("fix output differs:\n%s", res.FixDiff)
	}
}

func writeStatus(t *testing.T, status map[string]string) {
	t.Helper()
	keys := make([]string, 0, len(status))
	for k := range status {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]string, len(keys))
	for _, k := range keys {
		ordered[k] = status[k]
	}
	b, _ := json.MarshalIndent(ordered, "", "  ")
	if err := os.WriteFile(eaResults, append(b, '\n'), 0o644); err != nil {
		t.Logf("write %s: %v", eaResults, err)
	}
}

// TestEAFixturesParse parses every EA fixture (markup stripped) with the
// most permissive version: PhpStorm accepts all syntax regardless of level.
func TestEAFixturesParse(t *testing.T) {
	idx, err := LoadEAIndex(eaIndexPath)
	if err != nil {
		t.Skipf("no EA index: %v", err)
	}
	seen := map[string]bool{}
	bad := 0
	for _, c := range idx.Cases {
		for _, rel := range []string{c.Fixture, c.Fixed} {
			if rel == "" || seen[rel] || !strings.HasSuffix(rel, ".php") {
				continue
			}
			seen[rel] = true
			marked, err := ReadEA(idx, rel)
			if err != nil {
				t.Fatal(err)
			}
			clean, _, err := ParseMarkup(marked)
			if err != nil {
				t.Fatalf("%s: %v", rel, err)
			}
			f := syntax.Parse(rel, clean, syntax.Options{Version: phpversion.Max, Permissive: true})
			if len(f.Errors) > 0 {
				bad++
				e := f.Errors[0]
				s := int(e.Span.Start)
				t.Errorf("%s: %s near %q", rel, e.Msg, clean[max(0, s-40):min(len(clean), s+40)])
			}
		}
	}
	t.Logf("%d fixtures parsed, %d with errors", len(seen), bad)
}

// reportResultHidden reports EA mismatches without upstream message text
// (clean room: implementers must not see EA wording).
func reportResultHidden(t *testing.T, res Result) {
	t.Helper()
	for _, m := range res.Missing {
		t.Errorf("missing: %s[%d:%d]", m.Severity, m.Start, m.End)
	}
	for _, u := range res.Unexpected {
		t.Errorf("unexpected: %s", u) // custos' own finding
	}
	if res.FixDiff != "" {
		t.Errorf("fix output differs:\n%s", res.FixDiff)
	}
}
