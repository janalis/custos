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
