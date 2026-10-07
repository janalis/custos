package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// layout writes files (module-relative path -> content) under a temp root.
func layout(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for p, c := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const wrapper = "package main\n\nimport \"os\"\n\nfunc main() { os.Exit(run()) }\n\nfunc run() int {\n\treturn 0\n}\n"

func runCheck(t *testing.T, root, profile string, extra ...string) (int, string) {
	t.Helper()
	p := filepath.Join(root, "cover.out")
	if err := os.WriteFile(p, []byte(profile), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := run(append(append([]string{"-what", "test", "-module", "m", "-root", root}, extra...), p), &out, &errb)
	return code, out.String() + errb.String()
}

func TestCovered(t *testing.T) {
	root := layout(t, map[string]string{"cmd/a/main.go": wrapper})
	// main's body unexecuted (exempt), run's body executed in one binary
	// and not in another (union counts).
	code, out := runCheck(t, root, "mode: set\nm/cmd/a/main.go:5.13,5.31 1 0\nm/cmd/a/main.go:7.16,9.2 1 0\nm/cmd/a/main.go:7.16,9.2 1 1\n")
	if code != 0 || !strings.Contains(out, "coverage: test 100%") {
		t.Fatalf("code %d: %s", code, out)
	}
}

func TestUncovered(t *testing.T) {
	twoStmts := "package main\n\nfunc main() {\n\tprintln(1)\n\tprintln(2)\n}\n"
	notMain := "package lib\n\nfunc main() { println(1) }\n"
	root := layout(t, map[string]string{"cmd/a/main.go": wrapper, "cmd/b/main.go": twoStmts, "lib/lib.go": notMain})
	code, out := runCheck(t, root, "mode: set\n"+
		"m/cmd/a/main.go:7.16,9.2 1 0\n"+ // run's body
		"m/cmd/b/main.go:3.13,6.2 2 0\n"+ // a main doing real work is not exempt
		"m/lib/lib.go:3.13,3.27 1 0\n"+ // nor a `main` outside package main
		"m/lib/lib.go:1.1,1.2 0 0\n") // zero-statement blocks never count
	want := []string{"uncovered: m/cmd/a/main.go:7.16,9.2", "uncovered: m/cmd/b/main.go:3.13,6.2", "uncovered: m/lib/lib.go:3.13,3.27", "3 uncovered block(s) in test"}
	if code != 1 {
		t.Fatalf("code %d: %s", code, out)
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in:\n%s", w, out)
		}
	}
}

func TestWithin(t *testing.T) {
	root := layout(t, map[string]string{"cmd/a/main.go": wrapper})
	body, err := mainBody(filepath.Join(root, "cmd/a/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		b    block
		want bool
	}{
		{block{line0: 5, col0: 13, line1: 5, col1: 31}, true},
		{block{line0: 4, col0: 1, line1: 5, col1: 20}, false}, // starts before the body
		{block{line0: 5, col0: 13, line1: 7, col1: 2}, false}, // ends after it
		{block{line0: 5, col0: 1, line1: 5, col1: 20}, false}, // same line, before the brace
	} {
		if got := within(c.b, body); got != c.want {
			t.Errorf("%+v: got %v", c.b, got)
		}
	}
}

func TestErrors(t *testing.T) {
	root := layout(t, map[string]string{"cmd/a/main.go": wrapper, "bad/bad.go": "package main\nfunc (\n"})
	for _, c := range []struct{ name, profile, want string }{
		{"fields", "mode: set\nonly two\n", "malformed profile line"},
		{"numbers", "m/cmd/a/main.go:7.16,9.2 x 0\n", "malformed profile line"},
		{"no colon", "nocolon 1 0\n", "malformed block"},
		{"range", "m/cmd/a/main.go:seven 1 0\n", "malformed block"},
		{"missing source", "m/cmd/zz/main.go:1.1,1.2 1 0\n", "no such file"},
		{"unparsable source", "m/bad/bad.go:1.1,1.2 1 0\n", "covercheck:"},
	} {
		code, out := runCheck(t, root, c.profile)
		if code != 2 || !strings.Contains(out, c.want) {
			t.Errorf("%s: code %d, output %q", c.name, code, out)
		}
	}
	// missing profile, usage errors
	var out, errb bytes.Buffer
	if code := run([]string{filepath.Join(root, "absent.out")}, &out, &errb); code != 2 || !strings.Contains(errb.String(), "no such file") {
		t.Errorf("absent profile: %d %q", code, errb.String())
	}
	errb.Reset()
	if code := run(nil, &out, &errb); code != 2 || !strings.Contains(errb.String(), "usage") {
		t.Errorf("no args: %d %q", code, errb.String())
	}
	errb.Reset()
	if code := run([]string{"-nope", "x"}, &out, &errb); code != 2 {
		t.Errorf("bad flag: %d", code)
	}
}

// A profile line longer than bufio.Scanner's limit is a read error.
func TestScannerError(t *testing.T) {
	root := t.TempDir()
	code, out := runCheck(t, root, strings.Repeat("x", 70*1024)+"\n")
	if code != 2 || !strings.Contains(out, "too long") {
		t.Errorf("code %d: %q", code, out)
	}
}
