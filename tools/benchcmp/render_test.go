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
				{ID: "quick", Label: "Quick", Kind: "linter", Version: "1.0.0", Status: "ok", Median: 0.02, PeakMB: 20, Findings: 9},
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
	golden(t, "page.md", render(sampleResults(), sampleFeatures()))
	r := sampleResults()
	r.Runner.CPU = "unknown"
	r.Corpora[0].Rows = append(r.Corpora[0].Rows, Row{ID: "custos-fix", Label: "custos fix", Kind: "fixer", Status: "timeout"})
	if got := render(r, nil); strings.Contains(got, "## Features") || !strings.Contains(got, "on linux/amd64, 4 cores") || !strings.Contains(got, "| Fixer | fixer | 1.0.0 | 2.00 s | — |") {
		t.Errorf("unknown CPU and failed custos fix:\n%s", got)
	}
	if got := render(&Results{}, nil); !strings.Contains(got, "No results yet") {
		t.Errorf("empty results:\n%s", got)
	}
}

func TestRenderPage(t *testing.T) {
	dir := t.TempDir()
	results, page := filepath.Join(dir, "r.json"), filepath.Join(dir, "p.md")
	features, readme := filepath.Join(dir, "f.json"), filepath.Join(dir, "README.md")
	if err := writeJSON(results, sampleResults()); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(features, sampleFeatures()); err != nil {
		t.Fatal(err)
	}
	if err := renderPage(results, page, "", ""); err != nil || read(t, page) != render(sampleResults(), nil) {
		t.Errorf("renderPage without features: %v", err)
	}
	if err := renderPage(results, page, features, ""); err != nil || read(t, page) != render(sampleResults(), sampleFeatures()) {
		t.Errorf("renderPage: %v", err)
	}
	write(t, readme, "# x\n\n"+readmeStart+"\nold\n"+readmeEnd+"\n\nrest\n")
	if err := renderPage(results, page, features, readme); err != nil || read(t, readme) != "# x\n\n"+readmeStart+"\n\n"+readmeSection(sampleResults(), sampleFeatures())+"\n"+readmeEnd+"\n\nrest\n" {
		t.Errorf("README section: %v\n%s", err, read(t, readme))
	}
	for name, args := range map[string][4]string{
		"missing results":     {filepath.Join(dir, "missing.json"), page, "", ""},
		"missing features":    {results, page, filepath.Join(dir, "missing.json"), ""},
		"readme, no features": {results, page, "", filepath.Join(dir, "missing.md")},
		"missing readme":      {results, page, features, filepath.Join(dir, "missing.md")},
	} {
		if err := renderPage(args[0], args[1], args[2], args[3]); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	write(t, readme, "no markers\n")
	if err := renderPage(results, page, features, readme); err == nil || !strings.Contains(err.Error(), "section") {
		t.Errorf("README without markers: %v", err)
	}
	write(t, results, "{")
	if err := renderPage(results, page, "", ""); err == nil {
		t.Error("bad results: no error")
	}
	_ = os.Remove(results)
}
