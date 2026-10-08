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

func TestResolveOptions(t *testing.T) {
	yes := true
	c, err := Resolve(t.TempDir(), File{
		ShortOpenTag: &yes,
		Rules:        map[string]RuleSettings{"UnnecessarySemicolon": {Severity: "error"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !c.ShortOpenTag || c.Rules["UnnecessarySemicolon"].Severity != "error" {
		t.Fatalf("got %+v", c)
	}
	for name, f := range map[string]File{
		"comparison style": {ComparisonStyle: "backwards"},
		"severity":         {Rules: map[string]RuleSettings{"UnnecessarySemicolon": {Severity: "fatal"}}},
	} {
		if _, err := Resolve(t.TempDir(), f); err == nil {
			t.Errorf("%s: invalid value accepted", name)
		}
	}
}

// TestComposerWithoutUsableVersion: constraints naming only unsupported
// versions fall back to the default target.
func TestComposerWithoutUsableVersion(t *testing.T) {
	if _, ok := lowestVersion("^4.0 || ^9.1"); ok {
		t.Fatal("unsupported versions accepted")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"require": {"php": "^4.0"}, "config": {"platform": {"php": "dev"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.PHP != phpver.Default || c.PHPSource != "default" {
		t.Fatalf("got php=%s src=%s", c.PHP, c.PHPSource)
	}
}

// TestLoadWithoutWorkingDirectory: a relative start directory cannot be
// resolved when the working directory is unreadable and $PWD is unset.
func TestLoadWithoutWorkingDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	sub := filepath.Join(t.TempDir(), "cwd")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	t.Setenv("PWD", "")
	if err := os.Chmod(sub, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(sub, 0o755)
	if _, err := Load("."); err == nil {
		t.Fatal("load with no working directory succeeded")
	}
}

// TestLoadWithoutMarkers: with no custos.json/composer.json up to the
// filesystem root, the start directory is the root.
func TestLoadWithoutMarkers(t *testing.T) {
	dir := t.TempDir()
	for d := dir; ; d = filepath.Dir(d) {
		for _, m := range []string{FileName, "composer.json"} {
			if _, err := os.Stat(filepath.Join(d, m)); err == nil {
				t.Skipf("marker %s above the temp dir", filepath.Join(d, m))
			}
		}
		if filepath.Dir(d) == d {
			break
		}
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := filepath.Abs(dir); c.Root != want || c.PHPSource != "default" {
		t.Fatalf("got root=%s src=%s", c.Root, c.PHPSource)
	}
}

func TestComposerPHPUnit(t *testing.T) {
	for body, want := range map[string]int{
		`{"require-dev": {"phpunit/phpunit": "^11.5"}}`:                        100,
		`{"require-dev": {"phpunit/phpunit": "^8.5.21 || ^9.3"}}`:              85,
		`{"require": {"phpunit/phpunit": "~9.1"}}`:                             91,
		`{"require-dev": {"phpunit/phpunit": "7"}}`:                            70,
		`{"require-dev": {"phpunit/phpunit": "^9.12"}}`:                        99,
		`{"require-dev": {"phpunit/phpunit": "dev-main"}}`:                     0,
		`{"require-dev": {"phpunit/phpunit": ""}, "require": {"x/y": "^1.0"}}`: 0,
		`not json`: 0,
	} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := composerPHPUnit(dir); got != want {
			t.Errorf("%s: got %d want %d", body, got, want)
		}
		c, err := Resolve(dir, File{})
		if err != nil {
			t.Fatal(err)
		}
		if got, _ := c.Rules["PhpUnitTests"].Options["@composerPHPUnit"].(int); got != want {
			t.Errorf("%s: option %d want %d", body, got, want)
		}
	}
	if composerPHPUnit(t.TempDir()) != 0 {
		t.Error("no composer.json")
	}
}
