package catalogue

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/project"
	"github.com/janalis/custos/internal/semantic/flow"
	"github.com/janalis/custos/internal/semantic/infer"
	"github.com/janalis/custos/internal/semantic/stubs"
)

// partialArgs is the argument list of a partial application on a line of
// testdata/partial/partial.php: one holding a `?` or ending with `, ...`.
var partialArgs = regexp.MustCompile(`\(([^()]*\?[^()]*|[^()]*,\s*\.{3})\)`)

// TestPartialApplicationCallsNothing: a partial application (`f(?, 1)`,
// PHP 8.6) only builds a closure, as `f(...)` does. Every rule runs at 8.6
// on the fixture and on the same file with each partial application
// rewritten as a first-class callable; a rule reporting more on a line of
// the first treats the closure as a call (its arity, its arguments, its
// effects or its result). Rules checking partial applications by design
// (analysis.PartialCallRule) are left out.
func TestPartialApplicationCallsNothing(t *testing.T) {
	src, err := os.ReadFile("testdata/partial/partial.php")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(src), "\n")
	marked := 0
	for i, line := range lines {
		if strings.HasSuffix(line, "// partial") {
			marked++
			lines[i] = partialArgs.ReplaceAllString(line, "(...)")
			if lines[i] == line {
				t.Fatalf("line %d: no partial application to rewrite", i+1)
			}
		}
	}
	if marked < 50 {
		t.Fatalf("only %d partial applications", marked)
	}
	partial := findingsAt86(t, src)
	neutral := findingsAt86(t, []byte(strings.Join(lines, "\n")))
	for _, f := range partial {
		if i := slices.Index(neutral, f); i >= 0 {
			neutral = slices.Delete(neutral, i, i+1)
			continue
		}
		t.Errorf("only with partial application: %s", f)
	}
}

// findingsAt86 runs every rule at PHP 8.6 on src alone and returns its
// findings as "line: rule".
func findingsAt86(t *testing.T, src []byte) []string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.php")
	if err := os.WriteFile(path, src, 0o600); err != nil {
		t.Fatal(err)
	}
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpversion.PHP86, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	files := []string{path}
	opt := syntax.Options{Version: phpversion.PHP86}
	ix := project.BuildIndex(files, opt, infer.Options{})
	e = e.WithIndex(ix).WithFlow(project.BuildFlowSnapshot(context.Background(), files, nil, opt, ix, flow.Options{}))
	// Rules that check partial applications report on them by design.
	partial := map[string]bool{}
	for _, r := range All() {
		if _, ok := r.(analysis.PartialCallRule); ok {
			partial[r.ID()] = true
		}
	}
	var out []string
	for _, r := range project.Run(e, files, opt) {
		for _, f := range r.Findings {
			if f.Rule == "syntax" || f.Rule == "internal" {
				t.Errorf("%s: %s", f.Rule, f.Message)
			}
			if partial[f.Rule] {
				continue
			}
			line := 1 + strings.Count(string(src[:f.Span.Start]), "\n")
			out = append(out, fmt.Sprintf("%d: %s", line, f.Rule))
		}
	}
	return out
}

// TestPartialApplicationNeverCrashes: every PHP function a rule names in its
// source, partially applied in several shapes, analyzed by every rule at
// PHP 8.6 without a rule failing. Rules that index a call's arguments by
// position must not take a placeholder for one.
func TestPartialApplicationNeverCrashes(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	sources, err := filepath.Glob("../rules/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	quoted := regexp.MustCompile(`"([a-z_][a-z0-9_]*)"`)
	ix := stubs.Index()
	names := map[string]bool{}
	for _, path := range sources {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range quoted.FindAllSubmatch(src, -1) {
			if f := ix.Function(string(m[1]), phpversion.PHP86); f != nil && f.Builtin {
				names[string(m[1])] = true
			}
		}
	}
	if len(names) < 200 {
		t.Fatalf("only %d functions named by rules", len(names))
	}
	shapes := []string{
		`(?)`, `(?, ?)`, `(?, ?, ?)`, `('x', ?)`, `(?, 'x')`, `('x', ...)`, `('x', 'y', ...)`, `('x', ?, 'y')`, `(?, ...)`,
		`('/a/i', ?)`, `('/a/i', ...)`, `('/a/i', 'b', ...)`, `('/a/i', 'b', ?)`, `(1, ?)`, `([], ?)`,
	}
	dir := t.TempDir()
	var files []string
	for _, name := range slices.Sorted(maps.Keys(names)) {
		var b strings.Builder
		b.WriteString("<?php\nfunction probe($a, $b) {\n")
		for _, shape := range shapes {
			b.WriteString("    $c[] = " + name + shape + ";\n    if (" + name + shape + ") {}\n")
		}
		b.WriteString("    return $c;\n}\n")
		path := filepath.Join(dir, name+".php")
		if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, path)
	}
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpversion.PHP86, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpversion.PHP86}
	ix2 := project.BuildIndex(files, opt, infer.Options{})
	e = e.WithIndex(ix2).WithFlow(project.BuildFlowSnapshot(context.Background(), files, nil, opt, ix2, flow.Options{}))
	for _, r := range project.Run(e, files, opt) {
		for _, f := range r.Findings {
			if f.Rule == "internal" || f.Rule == "syntax" {
				t.Errorf("%s: %s: %s", filepath.Base(r.Path), f.Rule, f.Message)
			}
		}
		for _, issue := range r.Incomplete {
			if issue.Reason == "rule-failure" {
				t.Errorf("%s at %d", issue.Reason, issue.Span.Start)
			}
		}
	}
	t.Logf("%d functions in %d shapes", len(names), len(shapes))
}

// TestCompileErrorsStillAnalyzed: rules that guard against recovered trees
// run on a file whose only errors are compile errors.
func TestCompileErrorsStillAnalyzed(t *testing.T) {
	src := []byte("<?php\nenum E { case A; public function __toString(): string { return 'a'; } }\n$s = str_repeat('x', -1);\n")
	var rules []analysis.Rule
	for _, r := range All() {
		if r.ID() == "StringRepeatNegativeCount" {
			rules = append(rules, r)
		}
	}
	e, err := analysis.NewEngine(rules, analysis.Config{PHP: phpversion.PHP86, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	r := project.AnalyzeBuffer(e, "e.php", src, syntax.Options{Version: phpversion.PHP86})
	if len(r.Errors) != 1 || len(r.Findings) != 1 || len(r.Incomplete) != 0 {
		t.Fatalf("errors %v, findings %v, incomplete %v", r.Errors, r.Findings, r.Incomplete)
	}
}

// TestPartialMethodApplicationNeverCrashes: every PHP method a rule names
// in its source, partially applied on a receiver of its class and through
// its class, analyzed by every rule at PHP 8.6 without a rule failing.
func TestPartialMethodApplicationNeverCrashes(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	sources, err := filepath.Glob("../rules/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	quoted := regexp.MustCompile(`"([A-Z][A-Za-z_]+)", "([a-z_][A-Za-z_]+)"`)
	ix := stubs.Index()
	pairs := map[[2]string]bool{}
	for _, path := range sources {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range quoted.FindAllSubmatch(src, -1) {
			class, method := string(m[1]), string(m[2])
			if c := ix.Class(class, phpversion.PHP86); c != nil && ix.FindMethod(class, method, phpversion.PHP86) != nil {
				pairs[[2]string{class, method}] = true
			}
		}
	}
	if len(pairs) < 40 {
		t.Fatalf("only %d methods named by rules", len(pairs))
	}
	shapes := []string{`(?)`, `(?, ?)`, `('x', ?)`, `(?, 'x')`, `('x', ...)`, `(...)`, `(1, ?)`, `('/a/i', ?)`}
	dir := t.TempDir()
	var files []string
	for pair := range pairs {
		var b strings.Builder
		b.WriteString("<?php\nfunction probe(\\" + pair[0] + " $o) {\n")
		for _, shape := range shapes {
			b.WriteString("    $c[] = $o->" + pair[1] + shape + ";\n    if ($o->" + pair[1] + shape + ") {}\n")
			b.WriteString("    $c[] = \\" + pair[0] + "::" + pair[1] + shape + ";\n")
		}
		b.WriteString("    return $c;\n}\n")
		path := filepath.Join(dir, pair[0]+"-"+pair[1]+".php")
		if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, path)
	}
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpversion.PHP86, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpversion.PHP86}
	index := project.BuildIndex(files, opt, infer.Options{})
	e = e.WithIndex(index).WithFlow(project.BuildFlowSnapshot(context.Background(), files, nil, opt, index, flow.Options{}))
	for _, r := range project.Run(e, files, opt) {
		for _, f := range r.Findings {
			if f.Rule == "internal" || f.Rule == "syntax" {
				t.Errorf("%s: %s: %s", filepath.Base(r.Path), f.Rule, f.Message)
			}
		}
	}
	t.Logf("%d methods in %d shapes", len(pairs), len(shapes))
}
