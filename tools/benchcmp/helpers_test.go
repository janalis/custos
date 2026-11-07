//go:build unix

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// The test binary doubles as the external programs benchcmp runs: a
// measured tool (BENCHCMP_FAKE=1), and composer, php and custos when invoked
// through a link of that name.
func TestMain(m *testing.M) {
	if os.Getenv("BENCHCMP_FAKE") == "1" {
		fakeTool()
	}
	switch filepath.Base(os.Args[0]) {
	case "composer":
		fakeComposer()
	case "php":
		fmt.Print("8.4.1")
		os.Exit(0)
	case "custos":
		fmt.Println("custos v1.2.3")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fakeTool prints FAKE_STDOUT and FAKE_STDERR, optionally starts a child
// in its process group (FAKE_CHILD), sleeps FAKE_SLEEP and exits FAKE_EXIT.
func fakeTool() {
	if os.Getenv("FAKE_CHILD") == "1" {
		child := exec.Command(os.Args[0])
		child.Env = append(os.Environ(), "FAKE_CHILD=0", "FAKE_STDOUT=", "FAKE_STDERR=")
		_ = child.Start()
	}
	fmt.Fprint(os.Stdout, os.Getenv("FAKE_STDOUT"))
	fmt.Fprint(os.Stderr, os.Getenv("FAKE_STDERR"))
	if d, err := time.ParseDuration(os.Getenv("FAKE_SLEEP")); err == nil {
		buf := make([]byte, 32<<20) // some resident memory to measure
		for i := range buf {
			buf[i] = byte(i)
		}
		time.Sleep(d)
		_ = buf[len(buf)-1]
	}
	code, _ := strconv.Atoi(os.Getenv("FAKE_EXIT"))
	os.Exit(code)
}

// fakeComposer creates vendor/ in the working directory, or fails when
// FAKE_COMPOSER_FAIL names its subcommand.
func fakeComposer() {
	if len(os.Args) > 1 && os.Getenv("FAKE_COMPOSER_FAIL") == os.Args[1] {
		fmt.Fprintln(os.Stderr, "composer failed")
		os.Exit(1)
	}
	if err := os.MkdirAll("vendor", 0o755); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// fakeBin returns a directory holding links to the test binary named after
// each program.
func fakeBin(t *testing.T, names ...string) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, n := range names {
		if err := os.Symlink(exe, filepath.Join(dir, n)); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// withPath puts dir first in PATH for the test (or alone when only is set).
func withPath(t *testing.T, dir string, only bool) {
	t.Helper()
	if only {
		t.Setenv("PATH", dir)
		return
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// gitRepo makes a repository with a PHP file in src/, an optional
// composer.json and the given tags, and returns its file:// URL.
func gitRepo(t *testing.T, composer bool, tags ...string) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{
			"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t",
			"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false",
		}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	write(t, filepath.Join(dir, "src", "a.php"), "<?php\n")
	if composer {
		write(t, filepath.Join(dir, "composer.json"), "{}\n")
	}
	git("add", ".")
	git("commit", "-q", "-m", "init")
	for _, tag := range tags {
		git("tag", tag)
	}
	return "file://" + dir
}
