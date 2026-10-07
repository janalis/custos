package rules

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/runner"
	"custos/internal/syntax"
)

// TestNoCrashOnCorpus runs every rule over a real-world corpus (CUSTOS_CORPUS,
// a list of directories) and fails on any rule panic ("internal" findings).
// Skipped with -short or without corpus.
func TestNoCrashOnCorpus(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
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
	files, err := runner.Discover(present, nil)
	if err != nil {
		t.Fatal(err)
	}
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpver.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpver.PHP85}
	e.SetIndex(runner.BuildIndex(files, opt))
	crashes := 0
	for _, r := range runner.Run(e, files, opt) {
		for _, f := range r.Findings {
			if f.Rule == "internal" {
				crashes++
				if crashes <= 10 {
					t.Errorf("%s: %s", r.Path, f.Message)
				}
			}
		}
	}
	t.Logf("%d files, %d internal errors", len(files), crashes)
}
