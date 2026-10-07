package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestHostileRepository runs analyse and fix on a repository whose files try
// to make custos read devices or write outside the project: a .php symlink
// to /dev/zero, a .php symlink to a file outside the project, and a
// custos.json baseline escaping the root.
func TestHostileRepository(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks")
	}
	dir := project(t)
	outside := filepath.Join(t.TempDir(), "victim.php")
	const victim = "<?php\n$v = !!$w;\n"
	if err := os.WriteFile(outside, []byte(victim), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "link.php")); err != nil {
		t.Skip(err)
	}
	if err := os.Symlink("/dev/zero", filepath.Join(dir, "zero.php")); err != nil {
		t.Skip(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _, stderr := runCLI(t, "fix", "--rule", "NestedNotOperators", dir)
		if !strings.Contains(stderr, "not a regular file") {
			t.Errorf("fix stderr: %s", stderr)
		}
		_, out, _ := runCLI(t, "analyse", "--rule", "NestedNotOperators", dir)
		if !strings.Contains(out, "zero.php") {
			t.Errorf("analyse output does not mention zero.php: %s", out)
		}
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("custos hung on a hostile repository")
	}
	if b, _ := os.ReadFile(outside); string(b) != victim {
		t.Fatalf("file outside the project was modified: %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "a.php")); !strings.Contains(string(b), "(bool)$y") {
		t.Fatalf("regular file not fixed: %q", b)
	}
	if err := os.WriteFile(filepath.Join(dir, "custos.json"), []byte(`{"baseline": "../../../../../../dev/zero"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runCLI(t, "analyse", dir); code != 2 || !strings.Contains(stderr, "leaves the project root") {
		t.Fatalf("escaping baseline: exit %d, %s", code, stderr)
	}
}
