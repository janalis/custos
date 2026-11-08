//go:build unix

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func adoptConfig() *Config {
	return &Config{
		Runs: 1, Timeout: "1m",
		Corpora:  []Corpus{{Name: "app", Repo: "https://example.com/app.git", Tag: "v1.0.0", Paths: []string{"src"}}},
		Packages: map[string]string{"v/tool": "1.0.0", "v/other": "2.0.0"},
		Tools:    []Tool{{ID: "t", Label: "T", Kind: "linter", Package: "v/tool", Cmd: []string{"vendor/bin/tool"}, Findings: "psalm"}},
	}
}

func TestAdoptTakesVersionsAndTagsOnly(t *testing.T) {
	trusted, measured := adoptConfig(), adoptConfig()
	measured.Corpora[0].Tag = "v1.1.0"
	measured.Packages["v/tool"] = "1.2.0"
	changes, err := adopt(trusted, measured)
	if err != nil || strings.Join(changes, "; ") != "app v1.0.0 → v1.1.0; v/tool 1.0.0 → 1.2.0" {
		t.Fatalf("changes %q: %v", changes, err)
	}
	if trusted.Corpora[0].Tag != "v1.1.0" || trusted.Packages["v/tool"] != "1.2.0" || trusted.Packages["v/other"] != "2.0.0" {
		t.Fatalf("not adopted: %+v", trusted)
	}
	if changes, err := adopt(trusted, measured); err != nil || len(changes) != 0 {
		t.Fatalf("unchanged: %q %v", changes, err)
	}
}

func TestAdoptRefusesOtherChanges(t *testing.T) {
	for name, edit := range map[string]func(*Config){
		"command":       func(c *Config) { c.Tools[0].Cmd = []string{"sh", "-c", "curl evil"} },
		"environment":   func(c *Config) { c.Tools[0].Env = map[string]string{"X": "1"} },
		"new tool":      func(c *Config) { c.Tools = append(c.Tools, c.Tools[0]) },
		"repository":    func(c *Config) { c.Corpora[0].Repo = "https://example.com/evil.git" },
		"new corpus":    func(c *Config) { c.Corpora = append(c.Corpora, c.Corpora[0]) },
		"new package":   func(c *Config) { c.Packages["v/evil"] = "1.0.0" },
		"lost package":  func(c *Config) { delete(c.Packages, "v/other") },
		"runs":          func(c *Config) { c.Runs = 9 },
		"bad version":   func(c *Config) { c.Packages["v/tool"] = "dev-main" },
		"bad tag":       func(c *Config) { c.Corpora[0].Tag = "main; rm -rf /" },
		"version range": func(c *Config) { c.Packages["v/tool"] = "^1.0" },
	} {
		trusted, measured := adoptConfig(), adoptConfig()
		edit(measured)
		if _, err := adopt(trusted, measured); err == nil {
			t.Errorf("%s: adopted", name)
		}
		if want := adoptConfig(); trusted.Corpora[0].Tag != want.Corpora[0].Tag || trusted.Packages["v/tool"] != want.Packages["v/tool"] {
			t.Errorf("%s: trusted configuration changed on refusal", name)
		}
	}
}

func TestAdoptCommand(t *testing.T) {
	dir := t.TempDir()
	trusted, measured := filepath.Join(dir, "trusted.json"), filepath.Join(dir, "measured.json")
	m := adoptConfig()
	m.Packages["v/other"] = "2.1.0"
	if err := writeJSON(trusted, adoptConfig()); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(measured, m); err != nil {
		t.Fatal(err)
	}
	if code, out, stderr := call("adopt", "-config", trusted, "-from", measured); code != 0 || out != "v/other 2.0.0 → 2.1.0\n" || !strings.Contains(read(t, trusted), `"v/other": "2.1.0"`) {
		t.Fatalf("adopt: %d %q %s", code, out, stderr)
	}
	m.Tools[0].Cmd = []string{"evil"}
	if err := writeJSON(measured, m); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := call("adopt", "-config", trusted, "-from", measured); code != 1 || !strings.Contains(stderr, "more than versions and tags") {
		t.Fatalf("tampered: %d %s", code, stderr)
	}
	if code, _, _ := call("adopt", "-config", trusted, "-from", filepath.Join(dir, "missing.json")); code != 1 {
		t.Fatal("missing measured configuration accepted")
	}
}

func adoptResultsSample() *Results {
	return &Results{Date: "2026-11-08", Corpora: []CorpusResult{{
		Name: "app", Repo: "https://evil.example/x.git", Tag: "{{x}}", Files: 3,
		Rows: []Row{
			{ID: "t", Label: "<img onerror=x>", Kind: "{{k}}", Version: "{{v}}", Status: "ok", Median: 1.5, PeakMB: 10, Findings: 4},
			{ID: "custos", Label: "x", Kind: "x", Version: "0.1.10-1-gf0966217", Status: "ok", Median: 0.5},
		},
	}}}
}

func TestAdoptResultsKeepsOnlyNumbers(t *testing.T) {
	cfg := adoptConfig()
	cfg.Tools = append(cfg.Tools, Tool{ID: "custos", Label: "custos analyse", Kind: "analyzer", Findings: "custos"})
	r := adoptResultsSample()
	if err := adoptResults(cfg, r); err != nil {
		t.Fatal(err)
	}
	c, row := r.Corpora[0], r.Corpora[0].Rows[0]
	if c.Repo != "https://example.com/app.git" || c.Tag != "v1.0.0" || c.Files != 3 {
		t.Errorf("corpus identity not rebuilt: %+v", c)
	}
	if row.Label != "T" || row.Kind != "linter" || row.Version != "1.0.0" || row.Median != 1.5 || row.Findings != 4 {
		t.Errorf("tool identity not rebuilt: %+v", row)
	}
	if custos := r.Corpora[0].Rows[1]; custos.Label != "custos analyse" || custos.Version != "0.1.10-1-gf0966217" {
		t.Errorf("custos row: %+v", custos)
	}
	for name, edit := range map[string]func(*Results){
		"unknown corpus": func(r *Results) { r.Corpora[0].Name = "other" },
		"unknown tool":   func(r *Results) { r.Corpora[0].Rows[0].ID = "evil" },
		"custos version": func(r *Results) { r.Corpora[0].Rows[1].Version = "{{constructor}}" },
	} {
		r := adoptResultsSample()
		edit(r)
		if err := adoptResults(cfg, r); err == nil {
			t.Errorf("%s: adopted", name)
		}
	}
}

func TestAdoptCommandResults(t *testing.T) {
	dir := t.TempDir()
	trusted, measured := filepath.Join(dir, "trusted.json"), filepath.Join(dir, "measured.json")
	rin, rout := filepath.Join(dir, "rin.json"), filepath.Join(dir, "rout.json")
	cfg := adoptConfig()
	cfg.Tools = append(cfg.Tools, Tool{ID: "custos", Label: "custos analyse", Kind: "analyzer", Findings: "custos"})
	for path, v := range map[string]any{trusted: cfg, measured: cfg, rin: adoptResultsSample()} {
		if err := writeJSON(path, v); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"adopt", "-config", trusted, "-from", measured, "-results", rout}
	if code, _, stderr := call(append(args, "-results-from", rin)...); code != 0 || !strings.Contains(read(t, rout), `"label": "T"`) || strings.Contains(read(t, rout), "onerror") {
		t.Fatalf("adopt results: %d %s", code, stderr)
	}
	for name, body := range map[string]string{"bad json": "{", "unknown tool": `{"corpora": [{"name": "app", "rows": [{"id": "evil"}]}]}`} {
		write(t, rin, body)
		if code, _, _ := call(append(args, "-results-from", rin)...); code != 1 {
			t.Errorf("%s: accepted", name)
		}
	}
	if code, _, _ := call(append(args, "-results-from", filepath.Join(dir, "missing.json"))...); code != 1 {
		t.Error("missing results accepted")
	}
	if err := writeJSON(rin, adoptResultsSample()); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := call("adopt", "-config", trusted, "-from", measured, "-results", dir, "-results-from", rin); code != 1 {
		t.Error("unwritable results accepted")
	}
}

func TestRenderEscapesMeasuredText(t *testing.T) {
	r := sampleResults()
	r.Date, r.PHP, r.Runner.OS, r.Runner.CPU = "{{a}}", "{{b}}", "<os>", "{{c}}"
	r.Corpora[0].Name, r.Corpora[0].Rows[0].Kind, r.Corpora[0].Rows[0].Version = "{{d}}", "{{e}}", "{{f}}"
	r.Corpora[0].Rows[1].Label = "{{g}}"
	if got := render(r, nil); strings.Contains(got, "{{") || strings.Contains(got, "<os>") {
		t.Errorf("unescaped text:\n%s", got)
	}
}
