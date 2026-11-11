package main

import (
	"bytes"
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
	findings(t, root, "b.php:1:1: info: Two. [RuleB] (fixable)\na.php:1:1: info: One. [RuleA]\n")
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

// findings sets what the stand-in custos reports on corpus "c".
func findings(t *testing.T, root, out string) {
	t.Helper()
	write(t, filepath.Join(root, "corpora", "c", "out.txt"), out)
}

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
	if string(record) != "# c 1.0\na.php:1:1: info: One. [RuleA]\nb.php:1:1: info: Two. [RuleB] (fixable)\n" {
		t.Fatalf("record: %q", record)
	}
	if code, out, _ := runWith(args...); code != 0 || out != "c 1.0: 2 findings, unchanged\n" {
		t.Fatalf("unchanged: %d %q", code, out)
	}
	// One finding moves, one appears twice, one has no rule.
	findings(t, root, "a.php:2:1: info: One. [RuleA]\nb.php:1:1: info: Two. [RuleB] (fixable)\nb.php:1:1: info: Two. [RuleB] (fixable)\nnoise\n")
	code, out, _ := runWith(args...)
	diff := filepath.Join(root, "records", "c.diff")
	want := "c 1.0: 1 findings removed, 3 added (" + diff + ")\n  (unknown rule): -0 +1\n  RuleA: -1 +1\n  RuleB: -0 +1\n"
	if code != 1 || out != want {
		t.Fatalf("changed: %d %q", code, out)
	}
	text, _ := os.ReadFile(diff)
	if string(text) != "- a.php:1:1: info: One. [RuleA]\n+ a.php:2:1: info: One. [RuleA]\n+ b.php:1:1: info: Two. [RuleB] (fixable)\n+ noise\n" {
		t.Fatalf("diff: %q", text)
	}
	// Back to the record: the difference file goes.
	findings(t, root, "a.php:1:1: info: One. [RuleA]\nb.php:1:1: info: Two. [RuleB] (fixable)\n")
	if code, _, _ := runWith(args...); code != 0 {
		t.Fatal("not unchanged again")
	}
	if _, err := os.Stat(diff); !os.IsNotExist(err) {
		t.Fatal("difference file kept")
	}
}

func TestNoFindings(t *testing.T) {
	root, args := fixture(t)
	findings(t, root, "")
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
