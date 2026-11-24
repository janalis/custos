package conformance

import (
	"cmp"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// mutationsReport is where TestClosureMutations writes its suspects.
const mutationsReport = "../../../.cache/mutations.json"

// mutationSuspect is a finding that survives when every call assigned to a
// variable it reads only builds a closure.
type mutationSuspect struct {
	Rule    string `json:"rule"`
	Fixture string `json:"fixture"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Var     string `json:"var"`
	Writes  int    `json:"writes"`
	Partial bool   `json:"partial"`
	Source  string `json:"source"`
}

// TestClosureMutations is a report, not a gate (CUSTOS_MUTATIONS=1, or
// `make mutations`). For each finding a rule reports on its own fixtures, it
// rewrites every call assigned, in the finding's scope, to a variable the
// finding reads into a closure: a partial application (`f($a, ...)`) from PHP
// 8.6, a first-class callable (`f(...)`) from 8.1. The variable then holds a
// closure, not the call's result, so a finding that survives is a suspect
// for a rule reading the closure as that result. Survivors also include
// findings that do not depend on the value (another write reaches them, or
// the rule only looks at syntax): triage them by hand from the report.
func TestClosureMutations(t *testing.T) {
	if os.Getenv("CUSTOS_MUTATIONS") == "" {
		t.Skip("set CUSTOS_MUTATIONS=1 (or run `make mutations`)")
	}
	rules, err := meta.All()
	if err != nil {
		t.Fatal(err)
	}
	var (
		mu       sync.Mutex
		suspects []mutationSuspect
	)
	for _, r := range rules {
		rule := r.ID
		if !wanted([]string{rule}) {
			continue
		}
		dir := filepath.Join(ownRoot, strings.ToLower(rule), "testdata")
		files, _ := filepath.Glob(filepath.Join(dir, "*.php"))
		for _, f := range files {
			if strings.HasSuffix(f, ".fixed.php") {
				continue
			}
			t.Run(rule+"/"+filepath.Base(f), func(t *testing.T) {
				t.Parallel()
				req, marked, _ := ownCase(t, rule, dir, f, ".json")
				found := mutateFixture(t, rule, req, marked)
				mu.Lock()
				suspects = append(suspects, found...)
				mu.Unlock()
			})
		}
	}
	t.Cleanup(func() {
		slices.SortFunc(suspects, func(a, b mutationSuspect) int {
			return cmp.Or(strings.Compare(a.Rule, b.Rule), strings.Compare(a.Fixture, b.Fixture), a.Line-b.Line, strings.Compare(a.Var, b.Var))
		})
		b, _ := json.MarshalIndent(suspects, "", " ")
		if err := os.WriteFile(mutationsReport, b, 0o644); err != nil {
			t.Error(err)
		}
		rulesSeen := map[string]bool{}
		for _, s := range suspects {
			rulesSeen[s.Rule] = true
		}
		t.Logf("%d surviving findings in %d rules: %s", len(suspects), len(rulesSeen), mutationsReport)
	})
}

// mutateFixture returns the findings of rule on the fixture that survive the
// closure rewrite of the variables they read.
func mutateFixture(t *testing.T, rule string, req Request, marked []byte) []mutationSuspect {
	t.Helper()
	clean, _, err := ParseMarkup(marked)
	if err != nil {
		t.Fatal(err)
	}
	// A fixture without a version runs at the newest one, where both
	// closure forms exist.
	if req.PHP == "" {
		req.PHP = phpversion.Max.String()
	}
	v := target(req)
	forms := []bool{false} // first-class callable
	if v.AtLeast(phpversion.PHP86) {
		forms = append(forms, true) // partial application
	}
	if !v.AtLeast(phpversion.PHP81) {
		return nil
	}
	file := syntax.Parse(req.Path, clean, parseOpts())
	if len(file.Errors) > 0 {
		return nil
	}
	req.Source = clean
	base, err := engine.Analyze(req)
	if err != nil {
		t.Fatal(err)
	}
	var out []mutationSuspect
	for _, finding := range base {
		if finding.Rule != rule {
			continue
		}
		for name, scope := range readVariables(file, finding) {
		forms:
			for _, partial := range forms {
				edits := closureEdits(file, clean, scope, name, partial)
				if len(edits) == 0 {
					continue
				}
				mutated, shift := applyEdits(clean, edits, finding.Start)
				req.Source = mutated
				after, err := engine.Analyze(req)
				if err != nil {
					t.Fatal(err)
				}
				for _, a := range after {
					if a.Rule == rule && a.Message == finding.Message && a.Start == finding.Start+shift {
						out = append(out, mutationSuspect{
							Rule: rule, Fixture: filepath.Base(req.Path), Line: 1 + strings.Count(string(clean[:finding.Start]), "\n"),
							Message: finding.Message, Var: "$" + name, Writes: len(edits), Partial: partial, Source: lineAt(mutated, a.Start),
						})
						// One suspect per variable: the other form adds nothing.
						break forms
					}
				}
			}
		}
	}
	return out
}

// readVariables maps each plain variable read inside the finding to its
// local scope (nil at file level).
func readVariables(file *syntax.File, finding Finding) map[string]syntax.Node {
	vars := map[string]syntax.Node{}
	syntax.InspectFile(file, func(n syntax.Node) bool {
		sp := n.Span()
		if int(sp.End) <= finding.Start || int(sp.Start) >= finding.End {
			return false
		}
		if x, ok := n.(*syntax.Variable); ok && x.Name != "" && x.Name != "this" && int(sp.Start) >= finding.Start {
			if _, seen := vars[x.Name]; !seen {
				vars[x.Name] = syntax.EnclosingVariableScope(x)
			}
		}
		return true
	})
	return vars
}

// edit replaces src[start:end] with text.
type edit struct {
	start, end int
	text       string
}

// closureEdits turns every call assigned to the variable in scope (not in
// nested scopes) into a closure.
func closureEdits(file *syntax.File, src []byte, scope syntax.Node, name string, partial bool) []edit {
	var edits []edit
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		a, ok := n.(*syntax.Assign)
		// A plain `=`: compound assignments keep the previous value.
		if !ok || a.ByRef || strings.TrimSpace(string(src[a.Var.Span().End:a.Value.Span().Start])) != "=" {
			return true
		}
		if v, ok := a.Var.(*syntax.Variable); !ok || v.Name != name {
			return true
		}
		var args *syntax.ArgList
		switch c := syntax.UnwrapParens(a.Value).(type) {
		case *syntax.FuncCall:
			args = c.Args
		case *syntax.MethodCall:
			args = c.Args
		case *syntax.StaticCall:
			args = c.Args
		}
		if args == nil || args.Partial() || slices.ContainsFunc(args.Args, func(x syntax.Expr) bool {
			arg, ok := x.(*syntax.Arg)
			return !ok || arg.ByRef || arg.Unpack
		}) {
			return true
		}
		open, closing := int(args.Span().Start), int(args.Span().End)-1
		switch {
		case !partial:
			edits = append(edits, edit{open + 1, closing, "..."})
		case len(args.Args) == 0:
			edits = append(edits, edit{closing, closing, "..."})
		case strings.Contains(string(src[args.Args[len(args.Args)-1].Span().End:closing]), ","):
			edits = append(edits, edit{closing, closing, " ..."})
		default:
			edits = append(edits, edit{closing, closing, ", ..."})
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(file, visit)
	} else if body := syntax.VariableScopeBody(scope); body != nil {
		syntax.Inspect(body, visit)
	}
	return edits
}

// applyEdits applies non-overlapping edits and returns how far they move
// the offset at.
func applyEdits(src []byte, edits []edit, at int) ([]byte, int) {
	slices.SortFunc(edits, func(a, b edit) int { return b.start - a.start })
	out := slices.Clone(src)
	shift := 0
	for _, e := range edits {
		out = slices.Concat(out[:e.start], []byte(e.text), out[e.end:])
		if e.end <= at {
			shift += len(e.text) - (e.end - e.start)
		}
	}
	return out, shift
}

// lineAt returns the trimmed source line holding offset i.
func lineAt(src []byte, i int) string {
	start := strings.LastIndexByte(string(src[:i]), '\n') + 1
	end := strings.IndexByte(string(src[i:]), '\n')
	if end < 0 {
		end = len(src) - i
	}
	return strings.TrimSpace(string(src[start : i+end]))
}
