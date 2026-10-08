package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"

	"custos/internal/analysis"
	"custos/internal/meta"
	"custos/internal/rules"
	"custos/internal/syntax"
)

func lspFrame(body string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

func TestTopLevel(t *testing.T) {
	if code, _, stderr := runCLI(t); code != 2 || !strings.Contains(stderr, "Usage:") {
		t.Fatalf("no args: %d %s", code, stderr)
	}
	for _, a := range []string{"version", "--version"} {
		if code, out, _ := runCLI(t, a); code != 0 || out != "custos dev\n" {
			t.Fatalf("%s: %d %q", a, code, out)
		}
	}
	for _, a := range []string{"help", "-h", "--help"} {
		if code, out, _ := runCLI(t, a); code != 0 || !strings.Contains(out, "Usage:") {
			t.Fatalf("%s: %d %q", a, code, out)
		}
	}
	// -h on a sub-command prints its flags and succeeds
	for _, c := range []string{"analyse", "fix", "rules", "lsp"} {
		if code, _, stderr := runCLI(t, c, "-h"); code != 0 || !strings.Contains(stderr, "Usage of "+c) {
			t.Fatalf("%s -h: %d %s", c, code, stderr)
		}
		if code, _, stderr := runCLI(t, c, "--bogus"); code != 2 || !strings.Contains(stderr, "flag provided but not defined") {
			t.Fatalf("%s --bogus: %d %s", c, code, stderr)
		}
	}
}

func TestLSPCommand(t *testing.T) {
	in := lspFrame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`) + lspFrame(`{"jsonrpc":"2.0","method":"exit"}`)
	var out, errb bytes.Buffer
	if code := run([]string{"lsp", "--stdio"}, strings.NewReader(in), &out, &errb); code != 0 {
		t.Fatalf("lsp: exit %d %s", code, errb.String())
	}
	if !strings.Contains(out.String(), `"name":"custos"`) {
		t.Fatalf("initialize reply: %s", out.String())
	}
}

func TestRulesCommand(t *testing.T) {
	code, out, _ := runCLI(t, "rules", "--json")
	if code != 0 || !strings.Contains(out, `"implemented": true`) || !strings.Contains(out, `"id": "UnnecessarySemicolon"`) {
		t.Fatalf("rules --json: %d %.300s", code, out)
	}
	defer func() { catalogue = meta.All }()
	catalogue = func() ([]meta.Rule, error) { return nil, errors.New("corrupt catalogue") }
	if code, _, stderr := runCLI(t, "rules"); code != 2 || !strings.Contains(stderr, "corrupt catalogue") {
		t.Fatalf("catalogue error: %d %s", code, stderr)
	}
}

func TestExplainCommand(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{nil, 2, "usage: custos explain"},
		{[]string{"a", "b"}, 2, "usage: custos explain"},
		{[]string{"Nope"}, 2, `unknown rule "Nope"`},
		{[]string{"CompactCanBeUsed"}, 0, "disabled by default"},
		{[]string{"NullPointerException"}, 0, ", experimental"},
		{[]string{"AlterInForeach"}, 0, "\nOptions:\n"},
		{[]string{"NestedNotOperators"}, 0, "has a quick-fix"},
	} {
		code, out, stderr := runCLI(t, append([]string{"explain"}, tc.args...)...)
		if code != tc.code || !strings.Contains(out+stderr, tc.want) {
			t.Errorf("explain %v: %d\n%s%s", tc.args, code, out, stderr)
		}
	}
}

// fakeRule is a rule missing from the catalogue.
type fakeRule struct{}

func (fakeRule) ID() string                           { return "NotInCatalogue" }
func (fakeRule) Kinds() []syntax.NodeKind             { return nil }
func (fakeRule) Check(*analysis.Context, syntax.Node) {}

func TestAnalyseOptions(t *testing.T) {
	dir := project(t)
	file := filepath.Join(dir, "a.php")
	for _, tc := range []struct {
		name string
		args []string
		code int
		want string // substring of stdout+stderr
	}{
		{"file argument", []string{"--rule", "UnnecessarySemicolon", file}, 0, "a.php:2:"},
		{"php", []string{"--php", "7.4", "--stats", "--rule", "UnnecessarySemicolon", dir}, 0, "PHP 7.4 (flag)"},
		{"bad php", []string{"--php", "nine", dir}, 2, "invalid version"},
		{"regular", []string{"--comparison-style", "regular", "--rule", "UnnecessarySemicolon", dir}, 0, "1 file(s) analysed"},
		{"yoda", []string{"--comparison-style", "yoda", "--rule", "UnnecessarySemicolon", dir}, 0, "1 file(s) analysed"},
		{"bad style", []string{"--comparison-style", "sideways", dir}, 2, "--comparison-style must be"},
		{"config dir", []string{"--config", dir, "--rule", "UnnecessarySemicolon", "--format", "github"}, 0, "a.php,line=2"},
		{"exclude", []string{"--exclude", "sub", "--rule", "UnnecessarySemicolon", dir}, 0, "1 file(s) analysed"},
		{"missing path", []string{filepath.Join(dir, "missing")}, 2, "no such file"},
		{"fail-on never", []string{"--fail-on", "never", "--rule", "UnnecessarySemicolon", dir}, 0, "1 info"},
		{"fail-on typo", []string{"--fail-on", "warn", dir}, 2, "--fail-on must be"},
		{"bad format", []string{"--format", "yaml", dir}, 2, `unknown format "yaml"`},
		{"profile in missing dir", []string{"--cpuprofile", filepath.Join(dir, "no", "cpu.out"), dir}, 2, "no such file"},
		{"baseline in missing dir", []string{"--generate-baseline", filepath.Join(dir, "no", "bl.json"), dir}, 2, "no such file"},
	} {
		code, out, stderr := runCLI(t, append([]string{"analyse"}, tc.args...)...)
		if code != tc.code || !strings.Contains(out+stderr, tc.want) {
			t.Errorf("%s: exit %d\n%s%s", tc.name, code, out, stderr)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "b.php"), []byte("<?php\n;;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, out, _ := runCLI(t, "analyse", "--exclude", "sub", "--rule", "UnnecessarySemicolon", dir); !strings.Contains(out, "1 file(s) analysed") {
		t.Fatalf("--exclude did not exclude: %s", out)
	}

	defer func() { registry = rules.All }()
	registry = func() []analysis.Rule { return []analysis.Rule{fakeRule{}} }
	if code, _, stderr := runCLI(t, "analyse", dir); code != 2 || !strings.Contains(stderr, "NotInCatalogue") {
		t.Fatalf("broken registry: %d %s", code, stderr)
	}
}

func TestAnalyseProfileAndStats(t *testing.T) {
	dir := project(t)
	prof := filepath.Join(t.TempDir(), "cpu.out")
	bl := filepath.Join(dir, "bl.json")
	if code, _, _ := runCLI(t, "analyse", "--rule", "UnnecessarySemicolon", "--generate-baseline", bl, dir); code != 0 {
		t.Fatal("baseline generation failed")
	}
	code, _, stderr := runCLI(t, "analyse", "--cpuprofile", prof, "--stats", "--baseline", bl, "--rule", "UnnecessarySemicolon", dir)
	if code != 0 || !strings.Contains(stderr, "1 finding(s) suppressed by baseline") {
		t.Fatalf("stats: %d %s", code, stderr)
	}
	if st, err := os.Stat(prof); err != nil || st.Size() == 0 {
		t.Fatalf("profile not written: %v", err)
	}
	// a CPU profile already running makes --cpuprofile fail
	var sink bytes.Buffer
	if err := pprof.StartCPUProfile(&sink); err != nil {
		t.Fatal(err)
	}
	defer pprof.StopCPUProfile()
	for _, c := range []string{"analyse", "fix"} {
		if code, _, stderr := runCLI(t, c, "--cpuprofile", prof, dir); code != 2 || !strings.Contains(stderr, "profiling already in use") {
			t.Errorf("%s with profiling active: %d %s", c, code, stderr)
		}
	}
}

// brokenWriter fails every write, like stdout closed by the reader.
type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestAnalyseOutputError(t *testing.T) {
	dir := project(t)
	var errb bytes.Buffer
	if code := run([]string{"analyse", dir}, strings.NewReader(""), brokenWriter{}, &errb); code != 2 || !strings.Contains(errb.String(), "closed pipe") {
		t.Fatalf("broken stdout: %d %s", code, errb.String())
	}
}

func TestFixEdgeCases(t *testing.T) {
	dir := project(t)
	clean := filepath.Join(dir, "clean.php")
	if err := os.WriteFile(clean, []byte("<?php\n$a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// semantic rule (needs the project index), nothing to fix
	if code, _, stderr := runCLI(t, "fix", "--rule", "MktimeUsage", dir); code != 0 || !strings.Contains(stderr, "fixed 0 file(s)") {
		t.Fatalf("semantic fix: %d %s", code, stderr)
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permissions")
	}
	// a read-only file cannot be rewritten; an unreadable one is not read;
	// both are reported, the other files are still fixed, exit 2
	ro := filepath.Join(dir, "b_readonly.php")
	locked := filepath.Join(dir, "c_locked.php")
	for _, p := range []string{ro, locked} {
		if err := os.WriteFile(p, []byte("<?php\n;;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(ro, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })
	code, _, stderr := runCLI(t, "fix", "--rule", "UnnecessarySemicolon", dir)
	if code != 2 || !strings.Contains(stderr, "2 file(s) could not be fixed") || !strings.Contains(stderr, "fixed 1 file(s)") ||
		strings.Count(stderr, "permission denied") != 2 {
		t.Fatalf("fix with failures: %d %s", code, stderr)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "a.php")); string(b) != "<?php\n$x = !!$y;\n" {
		t.Fatalf("a.php not fixed: %q", b)
	}
}
