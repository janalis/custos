package catalogue

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/project"
)

// TestFixesKeepCodeParsable applies every quick-fix offered on a real-world
// corpus, one finding at a time and then all together per file (the
// `custos fix` loop), and requires the result to parse with no more syntax
// errors than the original. Skipped with -short or without a
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
	e = e.WithIndex(project.BuildIndex(files, opt))

	broken := map[string][]string{} // rule -> examples
	applied := 0
	combined := 0
	for _, r := range project.Run(e, files, opt) {
		base := len(r.Errors)
		fixable := false
		for _, fd := range r.Findings {
			fixable = fixable || len(fd.Fixes) > 0
			for _, fx := range fd.Fixes {
				out, n := fixing.Apply(r.Src, fx.Edits())
				if n == 0 {
					continue
				}
				applied++
				line := 1 + countNewlines(r.Src[:fd.Span.Start])
				if got := len(syntax.ParseBest(r.Path, out, opt).Errors); got > base {
					broken[fd.Rule] = append(broken[fd.Rule], fmt.Sprintf("%s:%d (%s)", r.Path, line, fx.Title))
					continue
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
	t.Logf("%d fixes applied on %d files (%d fixed with all fixes combined), %d rules with broken fixes", applied, len(files), combined, len(broken))
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
