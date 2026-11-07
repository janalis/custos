//go:build unix

package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func sampleResults() *Results {
	return &Results{
		Date:   "2026-11-09",
		Runner: Machine{OS: "linux/amd64", CPU: "Test CPU", Cores: 4, MemoryGB: 16},
		PHP:    "8.4.1",
		Runs:   3,
		Corpora: []CorpusResult{{
			Name: "app", Repo: "https://github.com/v/app.git", Tag: "v1.0.0", Files: 120,
			Rows: []Row{
				{ID: "custos", Label: "custos analyse", Kind: "analyzer", Version: "1.0.0", Status: "ok", Median: 0.5, Min: 0.4, Max: 0.6, PeakMB: 80, Findings: 12},
				{ID: "slow", Label: "Slow", Kind: "analyzer", Version: "v2.0.0", Status: "ok", Median: 150, PeakMB: 2048, Findings: 300},
				{ID: "broken", Label: "Broken", Kind: "analyzer", Version: "1.0.0", Status: "error: <bad> | {{x}}\nmore"},
				{ID: "mid", Label: "Mid", Kind: "analyzer", Version: "1.0.0", Status: "ok", Median: 12.34, PeakMB: 300, Findings: 40},
				{ID: "lint", Label: "Lint", Kind: "linter", Version: "1.0.0", Status: "ok", Median: 1, PeakMB: 30, Findings: 7},
				{ID: "fixer", Label: "Fixer", Kind: "fixer", Version: "1.0.0", Status: "ok", Median: 2, PeakMB: 50, Findings: 3},
			},
		}},
	}
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		write(t, path, got)
	}
	if want := read(t, path); got != want {
		t.Errorf("%s differs (go test -update to accept):\n%s", name, got)
	}
}

func TestRender(t *testing.T) {
	golden(t, "page.md", render(sampleResults()))
	r := sampleResults()
	r.Runner.CPU = "unknown"
	r.Corpora[0].Rows = append(r.Corpora[0].Rows, Row{ID: "custos-fix", Label: "custos fix", Kind: "fixer", Status: "timeout"})
	if got := render(r); !strings.Contains(got, "on linux/amd64, 4 cores") || !strings.Contains(got, "| Fixer | fixer | 1.0.0 | 2.00 s | — |") {
		t.Errorf("unknown CPU and failed custos fix:\n%s", got)
	}
	if got := render(&Results{}); !strings.Contains(got, "No results yet") {
		t.Errorf("empty results:\n%s", got)
	}
}

func TestRenderPage(t *testing.T) {
	dir := t.TempDir()
	results, page := filepath.Join(dir, "r.json"), filepath.Join(dir, "p.md")
	if err := writeJSON(results, sampleResults()); err != nil {
		t.Fatal(err)
	}
	if err := renderPage(results, page); err != nil || read(t, page) != render(sampleResults()) {
		t.Errorf("renderPage: %v", err)
	}
	if err := renderPage(filepath.Join(dir, "missing.json"), page); err == nil {
		t.Error("missing results: no error")
	}
	write(t, results, "{")
	if err := renderPage(results, page); err == nil {
		t.Error("bad results: no error")
	}
	_ = os.Remove(results)
}
