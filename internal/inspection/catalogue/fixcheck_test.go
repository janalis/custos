package catalogue

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/project"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/infer"
)

// TestFixesKeepCodeParsable applies every quick-fix offered on a real-world
// corpus, one finding at a time and then all together per file (the
// `custos fix` loop), and requires the result to parse with no more syntax
// errors than the original. Each fix applied alone must also remove its
// finding: its rule, run again on the result, must not report it there. Skipped with -short or without a
// corpus (CUSTOS_CORPUS, a list of directories).
func TestFixesKeepCodeParsable(t *testing.T) {
	if os.Getenv("CUSTOS_FIXCHECK") == "" {
		t.Skip("set CUSTOS_FIXCHECK=1 (or run `make fixcheck`); CUSTOS_FIXCHECK=php also lints samples with php -l")
	}
	phpLint := os.Getenv("CUSTOS_FIXCHECK") == "php"
	linted := map[string]int{}
	roots := filepath.SplitList(os.Getenv("CUSTOS_CORPUS"))
	var present []string
	for _, r := range roots {
		if _, err := os.Stat(r); err == nil {
			present = append(present, r)
		}
	}
	if len(present) == 0 {
		t.Skip("no corpus found")
	}
	files, err := project.Discover(present, nil)
	if err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpversion.PHP85}
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	ix := project.BuildIndex(files, opt, infer.Options{})
	snapshot := project.BuildFlowSnapshot(context.Background(), files, nil, opt, ix, flow.Options{})
	e = e.WithIndex(ix).WithFlow(snapshot)
	// Each fix's result is analysed again with its rule alone.
	alone := map[string]*analysis.Engine{}
	for _, r := range All() {
		one, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
		if err != nil {
			t.Fatal(err)
		}
		alone[r.ID()] = one.WithIndex(ix).WithFlow(snapshot)
	}
	lingering := map[string][]string{} // rule -> examples

	broken := map[string][]string{} // rule -> examples
	applied := 0
	combined := 0
	for _, r := range project.Run(e, files, opt) {
		base := len(r.Errors)
		fixable := false
		for _, fd := range r.Findings {
			fixable = fixable || len(fd.Fixes) > 0
			for _, fx := range fd.Fixes {
				edits := fx.Edits()
				out, n := fixing.Apply(r.Src, edits)
				if n == 0 {
					continue
				}
				applied++
				line := 1 + countNewlines(r.Src[:fd.Span.Start])
				if got := len(syntax.ParseBest(r.Path, out, opt).Errors); got > base {
					broken[fd.Rule] = append(broken[fd.Rule], fmt.Sprintf("%s:%d (%s)", r.Path, line, fx.Title))
					continue
				}
				if stillReported(project.AnalyzeBuffer(alone[fd.Rule], r.Path, out, opt).Findings, fd, edits, r.Src, out) {
					lingering[fd.Rule] = append(lingering[fd.Rule], fmt.Sprintf("%s:%d (%s)", r.Path, line, fx.Title))
				}
				// Optional ground truth: lint a sample per rule with the real PHP.
				if phpLint && base == 0 && linted[fd.Rule] < 20 {
					linted[fd.Rule]++
					if msg := phpLintSource(t, out); msg != "" {
						broken[fd.Rule] = append(broken[fd.Rule], fmt.Sprintf("%s:%d (%s) php -l: %s", r.Path, line, fx.Title, msg))
					}
				}
			}
		}
		// All fixes together, as `custos fix` applies them: edits of
		// different rules must not combine into invalid code.
		if fixable {
			res := fixing.FixSource(e, r.Path, r.Src, fixing.Options{Parse: opt})
			combined++
			if got := len(syntax.ParseBest(r.Path, res.Source, opt).Errors); got > base {
				broken["(all fixes combined)"] = append(broken["(all fixes combined)"], r.Path)
			} else if phpLint && base == 0 && linted["(all fixes combined)"] < 50 && res.Applied > 0 {
				linted["(all fixes combined)"]++
				if msg := phpLintSource(t, res.Source); msg != "" {
					broken["(all fixes combined)"] = append(broken["(all fixes combined)"], r.Path+" php -l: "+msg)
				}
			}
		}
	}
	rules := make([]string, 0, len(broken))
	for k := range broken {
		rules = append(rules, k)
	}
	sort.Strings(rules)
	for _, k := range rules {
		ex := broken[k]
		t.Errorf("%s: %d fix(es) produce unparsable code, e.g. %s", k, len(ex), ex[0])
	}
	rules = rules[:0]
	for k := range lingering {
		rules = append(rules, k)
	}
	sort.Strings(rules)
	for _, k := range rules {
		ex := lingering[k]
		t.Errorf("%s: %d fix(es) leave their finding, e.g. %s", k, len(ex), ex[0])
	}
	t.Logf("%d fixes applied on %d files (%d fixed with all fixes combined), %d rules with broken fixes, %d with fixes that leave their finding", applied, len(files), combined, len(broken), len(lingering))
}

// lineAt returns the line of src holding offset pos, without its
// indentation: a finding on a keyword is told from another by its line.
func lineAt(src []byte, pos int) string {
	start := bytes.LastIndexByte(src[:pos], '\n') + 1
	end := len(src)
	if i := bytes.IndexByte(src[pos:], '\n'); i >= 0 {
		end = pos + i
	}
	return strings.TrimSpace(string(src[start:end]))
}

// stillReported reports whether findings, from the source out that fd's
// fix edits made of src, hold fd again: the same rule and message on the
// same text and line, starting where fd's start moved to (anywhere in the
// replacement when an edit covered it). A finding the fix only reveals,
// such as the next level of a nesting it flattened, is on other text.
func stillReported(findings []diagnostic.Finding, fd diagnostic.Finding, edits []diagnostic.TextEdit, src, out []byte) bool {
	moved := func(pos int) int {
		for _, ed := range edits {
			if int(ed.Span.End) <= pos {
				pos += len(ed.NewText) - int(ed.Span.End-ed.Span.Start)
			}
		}
		return pos
	}
	start := int(fd.Span.Start)
	from, to := moved(start), moved(start)
	for _, ed := range edits {
		if int(ed.Span.Start) <= start && start < int(ed.Span.End) {
			from = moved(int(ed.Span.Start))
			to = from + len(ed.NewText)
		}
	}
	for _, got := range findings {
		if got.Rule == fd.Rule && got.Message == fd.Message && from <= int(got.Span.Start) && int(got.Span.Start) <= to &&
			string(out[got.Span.Start:got.Span.End]) == string(src[fd.Span.Start:fd.Span.End]) &&
			lineAt(out, int(got.Span.Start)) == lineAt(src, int(fd.Span.Start)) {
			return true
		}
	}
	return false
}

// phpLintSource runs `php -l` on src and returns the error output ("" = OK).
func phpLintSource(t *testing.T, src []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fix-*.php")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write(src)
	_ = f.Close()
	out, err := exec.Command("php", "-l", f.Name()).CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func countNewlines(b []byte) int {
	n := 0
	for _, c := range b {
		if c == '\n' {
			n++
		}
	}
	return n
}
