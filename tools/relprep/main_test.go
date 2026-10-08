package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testCL = "# Changelog\n\nIntro.\n\n## [Unreleased]\n\n### Added\n- Thing.\n\n## [0.1.0] - 2026-01-01\n\n- Old.\n"

const testLauncher = "<?php\nconst VERSION = '0.0.0';\necho VERSION;\n"

// repo makes a git repository with the given tags, a changelog and a launcher.
func repo(t *testing.T, tags ...string) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("commit", "-q", "--allow-empty", "-m", "init")
	for _, tag := range tags {
		git("tag", tag)
	}
	write(t, filepath.Join(dir, "CHANGELOG.md"), testCL)
	write(t, filepath.Join(dir, "custos"), testLauncher)
	return dir
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func runIn(dir string, extra ...string) (int, string, string) {
	var out, errb bytes.Buffer
	args := append([]string{"-git", dir, "-changelog", filepath.Join(dir, "CHANGELOG.md"), "-launcher", filepath.Join(dir, "custos"), "-date", "2026-10-08"}, extra...)
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRelease(t *testing.T) {
	dir := repo(t, "v0.1.0", "v0.2.3", "v0.10.0-rc1", "vnext")
	notes := filepath.Join(dir, "notes.md")
	code, out, errs := runIn(dir, "-bump", "minor", "-notes", notes)
	if code != 0 {
		t.Fatalf("code %d: %s", code, errs)
	}
	if out != "version=0.3.0\ntag=v0.3.0\n" {
		t.Fatalf("output %q", out)
	}
	wantCL := "# Changelog\n\nIntro.\n\n## [Unreleased]\n\n## [0.3.0] - 2026-10-08\n\n### Added\n- Thing.\n\n## [0.1.0] - 2026-01-01\n\n- Old.\n"
	if got := read(t, filepath.Join(dir, "CHANGELOG.md")); got != wantCL {
		t.Fatalf("changelog:\n%s", got)
	}
	if got := read(t, notes); got != "### Added\n- Thing.\n" {
		t.Fatalf("notes %q", got)
	}
	if got := read(t, filepath.Join(dir, "custos")); got != "<?php\nconst VERSION = '0.3.0';\necho VERSION;\n" {
		t.Fatalf("launcher %q", got)
	}
}

func TestNextVersion(t *testing.T) {
	tags := []string{"v1.2.3", "v1.10.0", "v0.9.9"}
	for bump, want := range map[string]string{"patch": "1.10.1", "minor": "1.11.0", "major": "2.0.0", "1.10.1": "1.10.1", "v3.0.0": "3.0.0"} {
		if got, err := nextVersion(bump, tags); err != nil || got != want {
			t.Errorf("%s: got %q, %v; want %s", bump, got, err, want)
		}
	}
	if got, _ := nextVersion("patch", nil); got != "0.0.1" {
		t.Errorf("no tags: %s", got)
	}
	for _, bad := range []string{"1.10.0", "1.9.0", "1.2", "01.2.3", "1.2.3-rc1", "99999999999999999999.0.0"} {
		if _, err := nextVersion(bad, tags); err == nil {
			t.Errorf("%s: accepted", bad)
		}
	}
}

func TestErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		setup func(dir string)
		args  []string
		want  string
	}{
		"bad date":      {args: []string{"-bump", "patch", "-date", "today"}, want: "bad date"},
		"bad version":   {args: []string{"-bump", "next"}, want: "bad version"},
		"no changelog":  {setup: func(d string) { os.Remove(filepath.Join(d, "CHANGELOG.md")) }, args: []string{"-bump", "patch"}, want: "CHANGELOG.md"},
		"no unreleased": {setup: func(d string) { write(t, filepath.Join(d, "CHANGELOG.md"), "# Changelog\n") }, args: []string{"-bump", "patch"}, want: "no \"## [Unreleased]\" heading"},
		"empty unreleased": {
			setup: func(d string) { write(t, filepath.Join(d, "CHANGELOG.md"), "# C\n\n## [Unreleased]\n\n") },
			args:  []string{"-bump", "patch"}, want: "section is empty",
		},
		"no launcher":     {setup: func(d string) { os.Remove(filepath.Join(d, "custos")) }, args: []string{"-bump", "patch"}, want: "custos"},
		"launcher no pin": {setup: func(d string) { write(t, filepath.Join(d, "custos"), "<?php\n") }, args: []string{"-bump", "patch"}, want: "found 0"},
		"notes unwritable": {
			args: []string{"-bump", "patch", "-notes", filepath.Join("missing", "dir", "notes.md")}, want: "notes.md",
		},
		"changelog unwritable": {
			setup: func(d string) { _ = os.Chmod(filepath.Join(d, "CHANGELOG.md"), 0o444) }, args: []string{"-bump", "patch"}, want: "CHANGELOG.md",
		},
		"launcher unwritable": {
			setup: func(d string) { _ = os.Chmod(filepath.Join(d, "custos"), 0o444) }, args: []string{"-bump", "patch"}, want: "custos",
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := repo(t)
			if tc.setup != nil {
				tc.setup(dir)
			}
			code, _, errs := runIn(dir, tc.args...)
			if code != 1 || !strings.Contains(errs, tc.want) {
				t.Fatalf("code %d: %s", code, errs)
			}
		})
	}
}

func TestUnreleasedLast(t *testing.T) {
	cl, body, err := cutChangelog("# C\n\n## [Unreleased]\n- x\n", "1.0.0", "2026-10-08")
	if err != nil || body != "- x" || cl != "# C\n\n## [Unreleased]\n\n## [1.0.0] - 2026-10-08\n- x\n" {
		t.Fatalf("%q %q %v", cl, body, err)
	}
}

func TestGitErrors(t *testing.T) {
	code, _, errs := runIn(t.TempDir(), "-bump", "patch")
	if code != 1 || !strings.Contains(errs, "git tag:") {
		t.Fatalf("not a repository: code %d: %s", code, errs)
	}
	t.Setenv("PATH", "")
	code, _, errs = runIn(t.TempDir(), "-bump", "patch")
	if code != 1 || !strings.Contains(errs, "relprep:") {
		t.Fatalf("no git: code %d: %s", code, errs)
	}
}

func TestUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"-bump", "patch", "extra"}, {"-nope"}} {
		var out, errb bytes.Buffer
		if code := run(args, &out, &errb); code != 2 || !strings.Contains(errb.String(), "usage:") {
			t.Fatalf("%v: code %d: %s", args, code, errb.String())
		}
	}
}
