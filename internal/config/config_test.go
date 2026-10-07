package config

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/analysis"
	"custos/internal/phpver"
)

func TestLowestVersion(t *testing.T) {
	cases := map[string]phpver.Version{
		"^7.4 || ^8.0": phpver.PHP74,
		">=8.1":        phpver.PHP81,
		"~5.6|^7.0":    phpver.PHP56,
		"8.2.*":        phpver.PHP82,
	}
	for in, want := range cases {
		if got, ok := lowestVersion(in); !ok || got != want {
			t.Errorf("%q: got %s want %s", in, got, want)
		}
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("composer.json", `{"require": {"php": ">=8.1"}}`)
	sub := filepath.Join(dir, "src", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	c, err := Load(sub)
	if err != nil {
		t.Fatal(err)
	}
	if c.Root != dir || c.PHP != phpver.PHP81 || c.PHPSource != "composer" {
		t.Fatalf("got root=%s php=%s src=%s", c.Root, c.PHP, c.PHPSource)
	}

	write(FileName, `{"php": "7.4", "comparisonStyle": "yoda", "rules": {"UnnecessarySemicolonInspection": {"enabled": false}}}`)
	c, err = Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.PHP != phpver.PHP74 || c.ComparisonStyle != analysis.StyleYoda {
		t.Fatalf("got %+v", c)
	}
	if rc, ok := c.Rules["UnnecessarySemicolon"]; !ok || rc.Enabled == nil || *rc.Enabled {
		t.Fatalf("legacy rule id not resolved: %+v", c.Rules)
	}

	write(FileName, `{"rules": {"Nope": {}}}`)
	if _, err := Load(dir); err == nil {
		t.Fatal("expected unknown rule error")
	}
}
