package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture installs corpus "c" at version 1.0 under a temporary directory,
// with a stand-in custos that prints the corpus's out.txt and exits with
// the status in its rc file (0 when absent). It returns the directory and
// the arguments that point corpusdiff at it.
func fixture(t *testing.T) (string, []string) {
	t.Helper()
	root := t.TempDir()
	bin := filepath.Join(root, "custos")
	script := "#!/bin/sh\ncat out.txt 2>/dev/null\necho failure >&2\nexit $(cat rc 2>/dev/null || echo 0)\n"
	write(t, bin, script)
	if err := os.Chmod(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "corpora", "c.version"), "1.0\n")
	report(t, root, two, one)
	return root, []string{"-corpora", "c", "-bin", bin, "-dir", filepath.Join(root, "corpora"), "-out", filepath.Join(root, "records")}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// findings sets what the stand-in custos prints on corpus "c".
func findings(t *testing.T, root, out string) {
	t.Helper()
	write(t, filepath.Join(root, "corpora", "c", "out.txt"), out)
}

// report has the stand-in custos report fs, as its JSON output does
// (lower-case keys, an analysis and files section corpusdiff ignores).
func report(t *testing.T, root string, fs ...finding) {
	t.Helper()
	list := make([]map[string]any, len(fs))
	for i, f := range fs {
		list[i] = map[string]any{
			"path": f.Path, "line": f.Line, "column": f.Column, "endLine": f.EndLine, "endColumn": f.EndColumn,
			"rule": f.Rule, "severity": f.Severity, "message": f.Message, "fixable": f.Fixable,
		}
	}
	b, err := json.Marshal(map[string]any{"analysis": map[string]any{"complete": true}, "files": []string{}, "findings": list})
	if err != nil {
		t.Fatal(err)
	}
	findings(t, root, string(b))
}

// one and two are findings the stand-in reports; two's message spans two
// lines and its fix is offered.
var (
	one    = finding{Path: "a.php", Line: 1, Column: 1, EndLine: 1, EndColumn: 4, Rule: "RuleA", Severity: "info", Message: "One."}
	two    = finding{Path: "b.php", Line: 1, Column: 1, EndLine: 2, EndColumn: 3, Rule: "RuleB", Severity: "warning", Message: "Two\nlines.", Fixable: true}
	oneAt2 = finding{Path: "a.php", Line: 2, Column: 1, EndLine: 2, EndColumn: 4, Rule: "RuleA", Severity: "info", Message: "One."}
)

const (
	oneLine    = "a.php:1:1-1:4: info: One. [RuleA]"
	twoLine    = `b.php:1:1-2:3: warning: Two\nlines. [RuleB] (fixable)`
	oneAt2Line = "a.php:2:1-2:4: info: One. [RuleA]"
)

func runWith(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRecordThenCompare(t *testing.T) {
	root, args := fixture(t)
	if code, out, _ := runWith(append(args, "-record")...); code != 0 || out != "c 1.0: 2 findings recorded\n" {
		t.Fatalf("record: %d %q", code, out)
	}
	record, _ := os.ReadFile(filepath.Join(root, "records", "c.txt"))
	if string(record) != "# c 1.0\n"+oneLine+"\n"+twoLine+"\n" {
		t.Fatalf("record: %q", record)
	}
	if code, out, _ := runWith(args...); code != 0 || out != "c 1.0: 2 findings, unchanged\n" {
		t.Fatalf("unchanged: %d %q", code, out)
	}
	// One finding moves and one appears twice.
	report(t, root, oneAt2, two, two)
	code, out, _ := runWith(args...)
	diff := filepath.Join(root, "records", "c.diff")
	want := "c 1.0: 1 findings removed, 2 added (" + diff + ")\n  RuleA: -1 +1\n  RuleB: -0 +1\n"
	if code != 1 || out != want {
		t.Fatalf("changed: %d %q", code, out)
	}
	text, _ := os.ReadFile(diff)
	if string(text) != "- "+oneLine+"\n+ "+oneAt2Line+"\n+ "+twoLine+"\n" {
		t.Fatalf("diff: %q", text)
	}
	// Back to the record: the difference file goes.
	report(t, root, one, two)
	if code, _, _ := runWith(args...); code != 0 {
		t.Fatal("not unchanged again")
	}
	if _, err := os.Stat(diff); !os.IsNotExist(err) {
		t.Fatal("difference file kept")
	}
}

func TestNoFindings(t *testing.T) {
	root, args := fixture(t)
	report(t, root)
	if code, out, _ := runWith(append(args, "-record")...); code != 0 || out != "c 1.0: 0 findings recorded\n" {
		t.Fatalf("record: %d %q", code, out)
	}
	if code, out, _ := runWith(args...); code != 0 || out != "c 1.0: 0 findings, unchanged\n" {
		t.Fatalf("compare: %d %q", code, out)
	}
}

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  func(root string)
		record bool
		want   string
	}{
		{"not installed", func(root string) { os.Remove(filepath.Join(root, "corpora", "c.version")) }, true, "not installed"},
		{"not installed", func(root string) { os.Remove(filepath.Join(root, "corpora", "c.version")) }, false, "not installed"},
		{"custos fails", func(root string) { write(t, filepath.Join(root, "corpora", "c", "rc"), "3") }, true, "custos: exit status 3: failure"},
		{"no record", func(string) {}, false, "no record"},
		{"output not JSON", func(root string) { findings(t, root, "a.php:1:1: info: One. [RuleA]") }, true, "custos output:"},
		{"other version", func(root string) { write(t, filepath.Join(root, "records", "c.txt"), "# c 0.9\n") }, false, "recorded on another version (c 0.9), installed 1.0"},
		{"custos fails", func(root string) {
			write(t, filepath.Join(root, "records", "c.txt"), "# c 1.0\n")
			write(t, filepath.Join(root, "corpora", "c", "rc"), "3")
		}, false, "exit status 3"},
		{"records is a file", func(root string) { write(t, filepath.Join(root, "records"), "") }, true, "not a directory"},
		{"record is a directory", func(root string) {
			if err := os.MkdirAll(filepath.Join(root, "records", "c.txt", "x"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, true, "is a directory"},
		{"difference is a directory", func(root string) {
			write(t, filepath.Join(root, "records", "c.txt"), "# c 1.0\n")
			write(t, filepath.Join(root, "records", "c.diff", "x"), "")
		}, false, "is a directory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, args := fixture(t)
			tc.setup(root)
			if tc.record {
				args = append(args, "-record")
			}
			if code, _, stderr := runWith(args...); code != 2 || !strings.Contains(stderr, tc.want) {
				t.Fatalf("%d %q", code, stderr)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	if code, _, stderr := runWith("extra"); code != 2 || !strings.HasPrefix(stderr, "usage:") {
		t.Fatalf("%d %q", code, stderr)
	}
}
