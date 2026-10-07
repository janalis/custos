package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "specs/Teapot.md"), "# Teapot\n\n## Summary\n\nFlags kettles that whistle twice.\n\n## Options\n\n- `LOUD` (bool, default false): also flag hums.\n\n## Examples\n\nnone here\n")
	write(t, filepath.Join(root, "specs/Saucer.md"), "# Saucer\n\n## Summary\nSaucers must be round.\n\n## Options\n\nNone.\n")
	write(t, filepath.Join(root, "specs/Spoon.md"), "# Spoon\n\nNo sections at all.\n")
	write(t, filepath.Join(root, "specs/Fork.md"), "# Fork\n## Summary\nForks.\n")
	write(t, filepath.Join(root, "specs/Cup.md"), "# Cup\n## Summary\nCups.\n")
	write(t, filepath.Join(root, "specs/Mug.md"), "# Mug\n## Summary\nMugs.\n")
	write(t, filepath.Join(root, "specs/_template.md"), "## Summary\nignored\n")
	write(t, filepath.Join(root, "specs/README.md"), "## Summary\nignored\n")
	write(t, filepath.Join(root, "internal/meta/.keep"), "")

	// Teapot: a real fix (content changes) in a subdirectory.
	write(t, filepath.Join(root, "testdata/rules/Teapot/sub/a.php"), "<?php <warning descr=\"x\">brew()</warning>;\n")
	write(t, filepath.Join(root, "testdata/rules/Teapot/sub/a.fixed.php"), "<?php steep();\n")
	// Saucer: only "no change" expectations (markup and whitespace ignored).
	write(t, filepath.Join(root, "testdata/rules/Saucer/b.php"), "<?php <error>spin()</error>;  <caret>\n")
	write(t, filepath.Join(root, "testdata/rules/Saucer/b.fixed.php"), "<?php\n spin();\n")
	// Fork: the expected output has no source file.
	write(t, filepath.Join(root, "testdata/rules/Fork/orphan.fixed.php"), "<?php\n")
	// Cup: the expected output is unreadable (a directory).
	write(t, filepath.Join(root, "testdata/rules/Cup/c.php"), "<?php\n")
	if err := os.MkdirAll(filepath.Join(root, "testdata/rules/Cup/c.fixed.php"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := run(root, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if got := stdout.String(); got != "descriptions: 6 rules\n" {
		t.Errorf("stdout = %q", got)
	}
	raw, err := os.ReadFile(filepath.Join(root, "internal/meta/descriptions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(raw, []byte("}\n")) {
		t.Errorf("missing trailing newline")
	}
	var got map[string]desc
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]desc{
		"Teapot": {Summary: "Flags kettles that whistle twice.", Options: "- `LOUD` (bool, default false): also flag hums.", Fix: true},
		"Saucer": {Summary: "Saucers must be round."},
		"Spoon":  {},
		"Fork":   {Summary: "Forks.", Fix: true},
		"Cup":    {Summary: "Cups.", Fix: true},
		"Mug":    {Summary: "Mugs."},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries: %v", len(got), got)
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s = %+v, want %+v", k, got[k], w)
		}
	}
}

func TestRunErrors(t *testing.T) {
	// Unreadable spec (a directory named like a spec).
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "specs/Odd.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(root, &stdout, &stderr); code != 1 || stderr.Len() == 0 {
		t.Fatalf("unreadable spec: exit %d, %q", code, stderr.String())
	}
	// Missing output directory.
	root = t.TempDir()
	write(t, filepath.Join(root, "specs/Lid.md"), "## Summary\nLids.\n")
	stderr.Reset()
	if code := run(root, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "descriptions.json") {
		t.Fatalf("missing output dir: exit %d, %q", code, stderr.String())
	}
}
