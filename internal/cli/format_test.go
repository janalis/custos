package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"

	"github.com/janalis/custos/internal/format"
)

const (
	messyPHP = "<?php\n$a=1;\n"
	cleanPHP = "<?php\n\n$a = 1;\n"
)

// formatProject writes files into a new project directory.
func formatProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	files["composer.json"] = `{"require": {"php": ">=8.2"}}`
	files[".editorconfig"] = "root = true\n"
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func runStdin(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := Run("dev", args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFormatWrite(t *testing.T) {
	dir := formatProject(t, map[string]string{
		"messy.php":          messyPHP,
		"clean.php":          cleanPHP,
		"broken.php":         "<?php\nif (\n",
		"ignored.php":        "<?php\n// @custos-format-ignore\n$a=1;\n",
		"legacy/old.php":     messyPHP,
		"skipdir/x.php":      messyPHP,
		"custos.json":        `{"format": {"exclude": ["legacy/**"]}}`,
		"long.php":           "<?php\n\n$a = 'a very long string literal';\n",
		"sub/.editorconfig":  "[*.php]\nindent_style = tab\n",
		"sub/nested.php":     "<?php\n\nif ($a) {\n    $b = 1;\n}\n",
		"perm/readonly.php":  messyPHP,
		"perm/.editorconfig": "",
	})
	if err := os.Chmod(filepath.Join(dir, "messy.php"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "messy.php"), filepath.Join(dir, "link.php")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(dir, "perm", "readonly.php"), 0o444); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCLI(t, "format", "--exclude", "skipdir", "--set", "report_long_lines=true", "--set", "line_length=20", "--set", "concat_space=one", dir)
	if code != 2 || !strings.Contains(stderr, "1 file(s) could not be formatted") {
		t.Fatalf("exit %d\n%s\n%s", code, out, stderr)
	}
	if !strings.Contains(stderr, "2 file(s) reformatted, 2 unchanged, 4 skipped\n") {
		t.Fatalf("summary: %s", stderr)
	}
	for _, want := range []string{"broken.php: 1 syntax error, not formatted", "link.php: not a regular file, not formatted", "readonly.php: permission denied"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
	// Only what formatting cannot fix is listed after a write.
	if !strings.Contains(out, "long.php\n  3:1  format.line_length") || strings.Contains(out, "format.spacing") {
		t.Errorf("listing: %s", out)
	}
	if got := readFile(t, filepath.Join(dir, "messy.php")); got != cleanPHP {
		t.Fatalf("messy.php: %q", got)
	}
	if st, _ := os.Stat(filepath.Join(dir, "messy.php")); st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", st.Mode())
	}
	if got := readFile(t, filepath.Join(dir, "sub", "nested.php")); !strings.Contains(got, "\n\t$b = 1;") {
		t.Fatalf("nested .editorconfig ignored: %q", got)
	}
	for _, name := range []string{"legacy/old.php", "skipdir/x.php", "ignored.php"} {
		if got := readFile(t, filepath.Join(dir, name)); strings.Contains(got, "$a = 1") {
			t.Errorf("%s formatted", name)
		}
	}
}

func TestFormatCheck(t *testing.T) {
	dir := formatProject(t, map[string]string{"a.php": messyPHP, "b.php": cleanPHP, "c.php": "<?php\nif (\n"})
	code, out, stderr := runCLI(t, "format", "--check", dir)
	if code != 1 || !strings.HasSuffix(stderr, "1 file(s) would be reformatted, 1 unchanged, 1 skipped\n") {
		t.Fatalf("exit %d\n%s", code, stderr)
	}
	if !strings.Contains(out, "/a.php\n  2:1  format.") || !strings.Contains(out, "c.php\n  1:1  format.skipped  Not formatted: 1 syntax error") {
		t.Fatalf("listing: %s", out)
	}
	if got := readFile(t, filepath.Join(dir, "a.php")); got != messyPHP {
		t.Fatal("--check wrote")
	}
	code, out, _ = runCLI(t, "format", "--check", "--report", "json", "--no-cache", dir)
	var res struct {
		Files    int `json:"files"`
		Findings []struct {
			Rule     string `json:"rule"`
			Severity string `json:"severity"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil || code != 1 || res.Files != 3 || len(res.Findings) == 0 || res.Findings[0].Severity != "info" {
		t.Fatalf("json: %d %v %s", code, err, out)
	}
	code, out, _ = runCLI(t, "format", "--dry-run", "--diff", dir)
	if code != 1 || !strings.Contains(out, "+$a = 1;") || strings.Contains(out, "format.") {
		t.Fatalf("diff: %d %s", code, out)
	}
	// A file with syntax errors alone does not fail a check.
	clean := formatProject(t, map[string]string{"b.php": cleanPHP, "c.php": "<?php\nif (\n"})
	for range 2 { // the second run takes b.php from the cache
		if code, _, stderr := runCLI(t, "format", "--check", clean); code != 0 || !strings.HasSuffix(stderr, "0 file(s) would be reformatted, 1 unchanged, 1 skipped\n") {
			t.Fatalf("clean: %d %s", code, stderr)
		}
	}
	// Paths default to the configured ones.
	t.Chdir(dir)
	if code, _, stderr := runCLI(t, "format", "--check"); code != 1 || !strings.Contains(stderr, "1 file(s) would be reformatted") {
		t.Fatalf("default paths: %d %s", code, stderr)
	}
}

func TestFormatStdin(t *testing.T) {
	dir := formatProject(t, map[string]string{"custos.json": `{"format": {"exclude": ["gen/**"]}}`, "tabs/.editorconfig": "[*.php]\nindent_style = tab\n"})
	t.Chdir(dir)
	src := "<?php\nif ($a) {\n$b=1;\n}\n"
	if code, out, stderr := runStdin(t, src, "format", "-"); code != 0 || out != "<?php\n\nif ($a) {\n    $b = 1;\n}\n" || stderr != "" {
		t.Fatalf("stdin: %d %q %q", code, out, stderr)
	}
	if code, out, _ := runStdin(t, src, "format", "--stdin", "--stdin-path", "tabs/x.php"); code != 0 || !strings.Contains(out, "\n\t$b = 1;") {
		t.Fatalf("stdin path: %d %q", code, out)
	}
	if code, out, _ := runStdin(t, src, "format", "--stdin-path", "gen/x.php", "-"); code != 0 || out != src {
		t.Fatalf("excluded: %d %q", code, out)
	}
	code, out, stderr := runStdin(t, src, "format", "--check", "--stdin-path", "src/x.php", "-")
	if code != 1 || !strings.HasPrefix(out, "src/x.php\n  2:1  format.") || stderr != "1 file(s) would be reformatted, 0 unchanged, 0 skipped\n" {
		t.Fatalf("check: %d %q %q", code, out, stderr)
	}
	if code, out, _ := runStdin(t, "<?php\n\n$a = 1;\n", "format", "--check", "-"); code != 0 || out != "" {
		t.Fatalf("check clean: %d %q", code, out)
	}
	if code, out, _ := runStdin(t, src, "format", "--dry-run", "--diff", "-"); code != 1 || !strings.Contains(out, "+    $b = 1;") {
		t.Fatalf("diff: %d %q", code, out)
	}
	if code, out, _ := runStdin(t, cleanPHP, "format", "--diff", "-"); code != 0 || out != "" {
		t.Fatalf("diff clean: %d %q", code, out)
	}
	broken := "<?php\nif (\n"
	if code, out, stderr := runStdin(t, broken, "format", "-"); code != 0 || out != broken || !strings.Contains(stderr, "stdin.php: 1 syntax error, not formatted") {
		t.Fatalf("broken: %d %q %q", code, out, stderr)
	}
}

func TestFormatShowConfig(t *testing.T) {
	dir := formatProject(t, map[string]string{
		"pint.json":   `{"preset": "psr12"}`,
		"custos.json": `{"format": {"settings": {"concat_space": "none"}}}`,
		"a.php":       cleanPHP,
	})
	code, out, _ := runCLI(t, "format", "--show-config", "--set", "line_length=99", dir)
	for _, want := range []string{"preset", "psr12", "pint.json", `"none"`, "custos.json", "99", "command line", "preset psr1"} {
		if !strings.Contains(out, want) {
			t.Errorf("show-config lacks %q", want)
		}
	}
	if code != 0 || strings.Count(out, "\n") != len(format.Registry())+1 {
		t.Fatalf("show-config: %d\n%s", code, out)
	}
	if code, out, _ := runCLI(t, "format", "--show-config", "--preset", "drupal", filepath.Join(dir, "a.php")); code != 0 || !strings.Contains(out, "drupal") || !strings.Contains(out, "command line") {
		t.Fatalf("show-config file: %d %s", code, out)
	}
	t.Chdir(dir)
	if code, out, _ := runCLI(t, "format", "--show-config"); code != 0 || !strings.Contains(out, "pint.json") {
		t.Fatalf("show-config default: %d %s", code, out)
	}
	if code, _, stderr := runCLI(t, "format", "--show-config", "--set", "nope=1"); code != 2 || !strings.Contains(stderr, `command line: unknown format setting "nope"`) {
		t.Fatalf("bad set: %d %s", code, stderr)
	}
}

// failWriter fails every write.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// failAfter fails the writes after the first n.
type failAfter struct{ n int }

func (w *failAfter) Write(b []byte) (int, error) {
	if w.n == 0 {
		return 0, errors.New("write failed")
	}
	w.n--
	return len(b), nil
}

func TestFormatErrors(t *testing.T) {
	dir := formatProject(t, map[string]string{"a.php": messyPHP})
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--report", "xml", dir}, `unknown report format "xml"`},
		{[]string{"--diff", "--report", "json", dir}, "--diff prints text"},
		{[]string{"--stdin", dir}, "cannot be formatted with other paths"},
		{[]string{"--set", "novalue", dir}, "want key=value"},
		{[]string{"--php", "x", dir}, "x"},
		{[]string{"--preset", "nope", dir}, `unknown preset "nope"`},
		{[]string{filepath.Join(dir, "missing")}, "no such file"},
		{[]string{"--bogus"}, "flag provided but not defined"},
	} {
		if code, _, stderr := runCLI(t, append([]string{"format"}, tc.args...)...); code != 2 || !strings.Contains(stderr, tc.want) {
			t.Errorf("%v: %d %s", tc.args, code, stderr)
		}
	}
	if code, _, stderr := runCLI(t, "format", "-h"); code != 0 || !strings.Contains(stderr, "Usage of format") {
		t.Fatalf("-h: %d %s", code, stderr)
	}
	// An unreadable .editorconfig fails the files below it.
	if err := os.Mkdir(filepath.Join(dir, "bad"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "bad", ".editorconfig"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad", "b.php"), []byte(cleanPHP), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runCLI(t, "format", "--check", filepath.Join(dir, "bad")); code != 2 || !strings.Contains(stderr, "1 file(s) could not be formatted") {
		t.Fatalf("editorconfig: %d %s", code, stderr)
	}
	if code, _, _ := runCLI(t, "format", "--show-config", filepath.Join(dir, "bad")); code != 2 {
		t.Fatalf("show-config editorconfig: %d", code)
	}
	if code, _, _ := runStdin(t, cleanPHP, "format", "--stdin-path", filepath.Join(dir, "bad", "x.php"), "-"); code != 2 {
		t.Fatalf("stdin editorconfig: %d", code)
	}
	// Output and cache failures.
	var errb bytes.Buffer
	for _, args := range [][]string{
		{"--check", dir},
		{"--show-config", dir},
		{"--check", "-"},
		{"-"},
	} {
		if code := Run("dev", append([]string{"format"}, args...), strings.NewReader(messyPHP), failWriter{}, &errb); code != 2 {
			t.Errorf("%v: failing stdout: %d", args, code)
		}
	}
	if code := Run("dev", []string{"format", "-"}, iotestErrReader{}, &bytes.Buffer{}, &errb); code != 2 {
		t.Errorf("failing stdin: %d", code)
	}
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUSTOS_CACHE_DIR", blocked)
	clean := formatProject(t, map[string]string{"b.php": cleanPHP})
	if code, _, stderr := runCLI(t, "format", "--check", clean); code != 0 || !strings.Contains(stderr, "format cache not saved") {
		t.Fatalf("cache: %d %s", code, stderr)
	}
	listed := &formatTally{}
	listed.skip(&errb, "x.php", "1 syntax error")
	if listed.list(&failAfter{n: 1}, "text", 1) == nil {
		t.Fatal("listing write error lost")
	}
	// The engine refusing its own output fails the run.
	tally := &formatTally{}
	tally.skip(&errb, "x.php", "formatting would change the code (syntax error); please report it")
	if code, err := tally.exit(true); code != 2 || err == nil {
		t.Fatalf("refused: %d %v", code, err)
	}
}

type iotestErrReader struct{}

func (iotestErrReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

// TestFormatLongLinesOnly counts a file whose only findings are lines too
// long apart from the files to reformat.
func TestFormatLongLinesOnly(t *testing.T) {
	long := "<?php\n\n$a = 'a string literal far longer than twenty columns';\n"
	dir := formatProject(t, map[string]string{"a.php": long})
	args := []string{"--set", "report_long_lines=true", "--set", "line_length=20"}
	code, _, stderr := runCLI(t, append(append([]string{"format", "--check", "--no-cache"}, args...), dir)...)
	if code != 1 || !strings.HasSuffix(stderr, "0 file(s) would be reformatted, 0 unchanged, 0 skipped, 1 with lines too long\n") {
		t.Fatalf("check: %d %q", code, stderr)
	}
	t.Chdir(dir)
	code, _, stderr = runStdin(t, long, append(append([]string{"format", "--check"}, args...), "-")...)
	if code != 1 || !strings.HasSuffix(stderr, ", 1 with lines too long\n") {
		t.Fatalf("stdin: %d %q", code, stderr)
	}
}

func TestFormatStdinTooLarge(t *testing.T) {
	big := "<?php\n" + strings.Repeat("$a = 1;\n", syntax.MaxFileSize/8+1)
	code, out, stderr := runStdin(t, big, "format", "-")
	if code != 2 || out != "" || !strings.Contains(stderr, "standard input exceeds") {
		t.Fatalf("too large: %d %d bytes %q", code, len(out), stderr)
	}
}
