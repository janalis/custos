//go:build unix

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func call(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"run", "-nope"}, {"run", "extra"}, {"dance"}} {
		if code, _, stderr := call(args...); code != 2 || !strings.Contains(stderr, "usage") {
			t.Errorf("%q: %d %s", args, code, stderr)
		}
	}
}

// project lays out a configuration with one installed corpus and one fake
// tool, and returns the common flags.
func project(t *testing.T) (string, []string) {
	t.Helper()
	dir := t.TempDir()
	exe, _ := os.Executable()
	repo := gitRepo(t, false, "1.0.0", "1.1.0")
	cfg := `{"runs": 1, "warmup": 0, "timeout": "1m",
  "corpora": [{"name": "c", "repo": "` + repo + `", "tag": "1.0.0", "paths": ["src"]}],
  "packages": {},
  "tools": [{"id": "t", "label": "T", "kind": "linter", "cmd": ["` + exe + `"], "env": {"BENCHCMP_FAKE": "1", "FAKE_STDOUT": "[{}]"}, "findings": "psalm"}]}`
	write(t, filepath.Join(dir, "competitors.json"), cfg)
	write(t, filepath.Join(dir, "configs", "x.conf"), "@CORPUS@\n")
	return dir, []string{
		"-config", filepath.Join(dir, "competitors.json"),
		"-work", filepath.Join(dir, "work"),
		"-configs", filepath.Join(dir, "configs"),
		"-custos", filepath.Join(fakeBin(t, "custos"), "custos"),
		"-results", filepath.Join(dir, "results.json"),
		"-page", filepath.Join(dir, "page.md"),
		"-packagist", "http://127.0.0.1:0",
	}
}

func TestSubcommands(t *testing.T) {
	withPath(t, fakeBin(t, "composer", "php"), false)
	dir, flags := project(t)
	sub := func(name string, extra ...string) (int, string, string) {
		return call(append(append([]string{name}, flags...), extra...)...)
	}
	if code, _, stderr := sub("install"); code != 0 {
		t.Fatalf("install: %d %s", code, stderr)
	}
	if code, _, stderr := sub("run", "-tool", "t", "-corpus", "c"); code != 0 {
		t.Fatalf("run: %d %s", code, stderr)
	}
	if code, _, stderr := sub("render"); code != 0 || !strings.Contains(read(t, filepath.Join(dir, "page.md")), "| T | linter |") {
		t.Fatalf("render: %d %s", code, stderr)
	}
	// No packages to look up, so Packagist is never contacted.
	if code, stdout, stderr := sub("bump"); code != 0 || stdout != "c 1.0.0 → 1.1.0\n" || !strings.Contains(read(t, filepath.Join(dir, "competitors.json")), `"tag": "1.1.0"`) {
		t.Fatalf("bump: %d %q %s", code, stdout, stderr)
	}

	for name, args := range map[string][]string{
		"unknown tool":   {"run", "-tool", "nope"},
		"unknown corpus": {"run", "-corpus", "nope"},
		"bad config":     {"run", "-config", filepath.Join(dir, "missing.json")},
		"no corpus":      {"run", "-work", filepath.Join(dir, "empty")},
		"bad work":       {"run", "-work", filepath.Join(dir, "competitors.json", "w")},
		"bad results":    {"render", "-results", filepath.Join(dir, "missing.json")},
		"bad packagist":  {"bump", "-config", withPackage(t, dir)},
	} {
		if code, _, stderr := call(append(append(args[:1:1], flags...), args[1:]...)...); code != 1 || !strings.HasPrefix(stderr, "benchcmp:") {
			t.Errorf("%s: %d %s", name, code, stderr)
		}
	}
}

// withPackage writes a copy of the configuration that pins a package.
func withPackage(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "pkg.json")
	write(t, path, strings.Replace(read(t, filepath.Join(dir, "competitors.json")), `"packages": {}`, `"packages": {"a/a": "1.0.0"}`, 1))
	return path
}
