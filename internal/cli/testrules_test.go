package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/pattern"
)

// writeTree writes files under dir, creating directories; a name ending in
// "/" is an empty directory.
func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, text := range files {
		path := filepath.Join(dir, name)
		if strings.HasSuffix(name, "/") {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const testRulesConfig = `{"customRules": [
	{"id": "StrictNull", "match": "$X == null", "message": "Compare $X with ===", "fix": "$X === null"},
	{"id": "NoDump", "match": "\\dump(...)", "paths": ["src/**"], "message": "Remove the dump", "severity": "error", "php": "8.0"}
]}`

func TestTestRules(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"custos.json":                             testRulesConfig,
		"custos-tests/StrictNull/basic.php":       "<?php\nif (<warning descr=\"Compare $a with ===\">$a == null</warning>) {}\n",
		"custos-tests/StrictNull/basic.fixed.php": "<?php\nif ($a === null) {}\n",
		"custos-tests/StrictNull/wrong.php":       "<?php\n$a == null;\n<warning descr=\"nope\">$b</warning>;\n",
		"custos-tests/StrictNull/wrong.fixed.php": "<?php\n$a == null;\n$b;\n",
		"custos-tests/NoDump/basic.php":           "<?php\n<error descr=\"Remove the dump\">dump(1)</error>;\n",
		"custos-tests/NoDump/basic.json":          `{"path": "src/a.php"}`,
		"custos-tests/NoDump/old.php":             "<?php\ndump(1);\n",
		"custos-tests/NoDump/old.json":            `{"path": "src/a.php", "php": "7.4"}`,
		"custos-tests/NoDump/outside.php":         "<?php\ndump(1);\n",
		"custos-tests/README.md":                  "not a rule",
	})
	t.Chdir(dir)
	code, out, stderr := runCLI(t, "test-rules")
	for _, want := range []string{
		"ok    NoDump/basic.php", "ok    NoDump/old.php", "ok    NoDump/outside.php", "ok    StrictNull/basic.php",
		"FAIL  StrictNull/wrong.php",
		`missing    3:1 warning "$b": "nope"`,
		`unexpected 2:1 warning "$a == null": "Compare $a with ==="`,
		"fixed source differs from wrong.fixed.php:\n      <?php\n      $a === null;",
		"5 fixture(s), 1 failed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if code != 1 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	// An explicit directory and a passing run.
	if code, out, _ := runCLI(t, "test-rules", filepath.Join(dir, "custos-tests")); code != 1 || !strings.Contains(out, "5 fixture(s)") {
		t.Fatalf("explicit dir: %d %s", code, out)
	}
	if err := os.RemoveAll(filepath.Join(dir, "custos-tests/StrictNull/wrong.php")); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := runCLI(t, "test-rules"); code != 0 || !strings.Contains(out, "4 fixture(s), 0 failed") {
		t.Fatalf("passing: %d %s", code, out)
	}
}

func TestTestRulesErrors(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		args  []string
		want  string
	}{
		{"flag", nil, []string{"--nope"}, "flag provided but not defined"},
		{"config", map[string]string{"custos.json": "{"}, nil, "custos.json"},
		{"no directory", map[string]string{"custos.json": "{}"}, nil, "rule fixtures:"},
		{"empty", map[string]string{"custos.json": `{"customRuleTests": "t"}`, "t/": ""}, nil, "none in"},
		{"unknown rule", map[string]string{"custos.json": "{}", "custos-tests/Nope/a.php": "<?php"}, nil, "Nope: unknown rule"},
		{"markup", map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php": "<?php <warning>"}, nil, "unclosed <warning>"},
		{"unreadable fixture", map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php/": ""}, nil, "a.php"},
		{"sidecar", map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php": "<?php", "custos-tests/StrictNull/a.json": "{"}, nil, "a.json:"},
		{"unreadable sidecar", map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php": "<?php", "custos-tests/StrictNull/a.json/": ""}, nil, "a.json"},
		{"php", map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php": "<?php", "custos-tests/StrictNull/a.json": `{"php": "9.9"}`}, nil, "phpver"},
		{"unreadable fixed", map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php": "<?php", "custos-tests/StrictNull/a.fixed.php/": ""}, nil, "a.fixed.php"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTree(t, dir, tc.files)
			t.Chdir(dir)
			if code, _, stderr := runCLI(t, append([]string{"test-rules"}, tc.args...)...); code != 2 || !strings.Contains(stderr, tc.want) {
				t.Fatalf("exit %d: %s", code, stderr)
			}
		})
	}
}

// A rule a program adds may not take the ID of a configured one.
func TestTestRulesEngineError(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"custos.json": testRulesConfig, "custos-tests/StrictNull/a.php": "<?php"})
	t.Chdir(dir)
	twin, err := pattern.Compile(pattern.Definition{ID: "StrictNull", Match: pattern.Patterns{"f()"}, Message: "m"})
	if err != nil {
		t.Fatal(err)
	}
	extra = []analysis.Rule{twin}
	defer func() { extra = nil }()
	if code, err := cmdTestRules(nil, &strings.Builder{}, &strings.Builder{}); code != 2 || err == nil || !strings.Contains(err.Error(), "registered twice") {
		t.Fatalf("%d %v", code, err)
	}
}
