package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func project(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("composer.json", `{"require": {"php": ">=8.2"}}`)
	write("a.php", "<?php\n$x = !!$y;;\n")
	return dir
}

func runCLI(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestAnalyseJSON(t *testing.T) {
	dir := project(t)
	code, out, stderr := runCLI(t, "analyse", "--format=json", "--rule", "NestedNotOperators,UnnecessarySemicolon", dir)
	if code != 0 { // both rules report info, below the default fail-on=warning
		t.Fatalf("exit %d, stderr %s", code, stderr)
	}
	var res struct {
		Files    int `json:"files"`
		Findings []struct {
			Rule string `json:"rule"`
			Line int    `json:"line"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if res.Files != 1 || len(res.Findings) != 2 {
		t.Fatalf("got %+v", res)
	}
	if code, _, _ := runCLI(t, "analyse", "--fail-on=info", "--rule", "UnnecessarySemicolon", dir); code != 1 {
		t.Fatalf("fail-on=info: exit %d", code)
	}
}

func TestFixDryRunAndApply(t *testing.T) {
	dir := project(t)
	file := filepath.Join(dir, "a.php")
	code, out, _ := runCLI(t, "fix", "--dry-run", "--diff", "--rule", "NestedNotOperators", dir)
	if code != 0 || !strings.Contains(out, "+$x = (bool)$y;;") {
		t.Fatalf("dry-run diff: exit %d\n%s", code, out)
	}
	if b, _ := os.ReadFile(file); string(b) != "<?php\n$x = !!$y;;\n" {
		t.Fatalf("dry-run modified the file: %q", b)
	}
	if code, _, stderr := runCLI(t, "fix", "--rule", "NestedNotOperators,UnnecessarySemicolon", dir); code != 0 {
		t.Fatalf("fix: exit %d %s", code, stderr)
	}
	if b, _ := os.ReadFile(file); string(b) != "<?php\n$x = (bool)$y;\n" {
		t.Fatalf("fixed content: %q", b)
	}
}

func TestMiscCommands(t *testing.T) {
	if code, out, _ := runCLI(t, "explain", "UnnecessarySemicolonInspection"); code != 0 || !strings.Contains(out, "UnnecessarySemicolon") {
		t.Fatalf("explain: %d %s", code, out)
	}
	if code, out, _ := runCLI(t, "rules"); code != 0 || strings.Count(out, "\n") != 178 {
		t.Fatalf("rules: %d lines", strings.Count(out, "\n"))
	}
	if code, _, _ := runCLI(t, "analyse", "--rule", "Nope", "."); code != 2 {
		t.Fatalf("unknown rule: exit %d", code)
	}
	if code, _, _ := runCLI(t, "bogus"); code != 2 {
		t.Fatalf("unknown command: exit %d", code)
	}
}

func TestBaselineFlags(t *testing.T) {
	dir := project(t)
	bl := filepath.Join(dir, "custos-baseline.json")
	if code, _, stderr := runCLI(t, "analyse", "--rule", "UnnecessarySemicolon", "--generate-baseline", bl, dir); code != 0 || !strings.Contains(stderr, "1 entries") {
		t.Fatalf("generate: %d %s", code, stderr)
	}
	code, out, _ := runCLI(t, "analyse", "--fail-on=info", "--format=json", "--rule", "UnnecessarySemicolon", "--baseline", bl, dir)
	if code != 0 || !strings.Contains(out, `"findings": []`) {
		t.Fatalf("baseline filter: %d %s", code, out)
	}
}

func TestConfiguredBaselineMayBeMissing(t *testing.T) {
	dir := project(t)
	if err := os.WriteFile(filepath.Join(dir, "custos.json"), []byte(`{"baseline": "custos-baseline.json"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runCLI(t, "analyse", "--rule", "UnnecessarySemicolon", dir); code != 0 {
		t.Fatalf("missing configured baseline: exit %d %s", code, stderr)
	}
	if code, _, _ := runCLI(t, "analyse", "--baseline", filepath.Join(dir, "nope.json"), dir); code != 2 {
		t.Fatalf("missing explicit baseline must fail, got %d", code)
	}
}
