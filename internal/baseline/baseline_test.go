package baseline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/report"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	php := filepath.Join(dir, "a.php")
	if err := os.WriteFile(php, []byte("<?php\n$a = !!$b;\n$c = !!$d;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items := []report.Item{
		{Path: php, Line: 2, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 3, Rule: "NestedNotOperators", Message: "m"},
	}
	bl := filepath.Join(dir, "custos-baseline.json")
	if n, err := Write(bl, items); err != nil || n != 2 {
		t.Fatalf("write: %d %v", n, err)
	}
	// Code moves down two lines and a new finding appears.
	if err := os.WriteFile(php, []byte("<?php\n// x\n// y\n$a = !!$b;\n$c = !!$d;\n$e = !!$f;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := []report.Item{
		{Path: php, Line: 4, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 5, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 6, Rule: "NestedNotOperators", Message: "m"},
	}
	b, err := Load(bl)
	if err != nil {
		t.Fatal(err)
	}
	rest, suppressed := b.Filter(moved)
	if suppressed != 2 || len(rest) != 1 || rest[0].Line != 6 {
		t.Fatalf("suppressed=%d rest=%+v", suppressed, rest)
	}
}

func TestLoadErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("missing baseline accepted")
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bad); err == nil || !strings.Contains(err.Error(), "bad.json") {
		t.Fatalf("invalid JSON: %v", err)
	}
}

// TestWriteSortsAndSkips: tool findings (syntax/internal/io) are never
// baselined; entries are sorted by path, rule then line text; a line past
// the end of the file (or an unreadable file) keys on empty text.
func TestWriteSortsAndSkips(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.php")
	b := filepath.Join(dir, "b.php")
	for _, p := range []string{a, b} {
		if err := os.WriteFile(p, []byte("<?php\n$x = 1;\n$y = 2;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	items := []report.Item{
		{Path: b, Line: 2, Rule: "R1", Message: "m"},
		{Path: a, Line: 3, Rule: "R2", Message: "m"},
		{Path: a, Line: 3, Rule: "R1", Message: "m"},
		{Path: a, Line: 2, Rule: "R1", Message: "m"},
		{Path: a, Line: 99, Rule: "R1", Message: "m"},
		{Path: a, Line: 1, Rule: "syntax", Message: "m"},
		{Path: a, Line: 1, Rule: "internal", Message: "m"},
		{Path: a, Line: 1, Rule: "io", Message: "m"},
	}
	bl := filepath.Join(dir, "bl.json")
	if n, err := Write(bl, items); err != nil || n != 5 {
		t.Fatalf("write: %d %v", n, err)
	}
	data, err := os.ReadFile(bl)
	if err != nil {
		t.Fatal(err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range f.Entries {
		got = append(got, e.Path+"|"+e.Rule+"|"+e.Line)
	}
	want := []string{"a.php|R1|", "a.php|R1|$x = 1;", "a.php|R1|$y = 2;", "a.php|R2|$y = 2;", "b.php|R1|$x = 1;"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("entries:\n got %v\nwant %v", got, want)
	}
	if _, err := Write(filepath.Join(dir, "no", "such", "dir.json"), items); err == nil {
		t.Fatal("write into a missing directory succeeded")
	}
}

// TestEmptyFiltersNothing: the empty baseline has no directory, so paths
// stay absolute and nothing is suppressed.
func TestEmptyFiltersNothing(t *testing.T) {
	items := []report.Item{{Path: filepath.Join(t.TempDir(), "a.php"), Line: 1, Rule: "R", Message: "m"}}
	rest, n := Empty().Filter(items)
	if n != 0 || len(rest) != 1 {
		t.Fatalf("suppressed=%d rest=%v", n, rest)
	}
	if got := rel("", "/abs/x.php"); got != "/abs/x.php" {
		t.Fatalf("rel without dir: %q", got)
	}
}

// TestNoWorkingDirectory: when the working directory cannot be resolved
// (unreadable and removed, $PWD unset) relative paths cannot be made
// absolute: Write reports it, and Filter keys on the path as given. getcwd
// fails on macOS for an unreadable directory and on Linux (even as root)
// for a removed one, so the test does both.
func TestNoWorkingDirectory(t *testing.T) {
	sub := filepath.Join(t.TempDir(), "cwd")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	t.Setenv("PWD", "")
	if err := os.Chmod(sub, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(sub); err != nil {
		t.Fatal(err)
	}
	if _, err := Write("bl.json", nil); err == nil {
		t.Fatal("write with no working directory succeeded")
	}
	if got := rel("/x", "a.php"); got != "a.php" {
		t.Fatalf("rel: %q", got)
	}
}
