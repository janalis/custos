package main

import (
	"bytes"
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

func dangling(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(filepath.Dir(path), "nowhere"), path); err != nil {
		t.Fatal(err)
	}
}

// Invented "upstream" text; none of it exists anywhere else.
const (
	sentence  = "Purple giraffes rarely juggle seven lemons"
	prefixed  = "Moonlit otters negotiate with the tide"
	longLine  = "$marmalade = orbitTheTeacup($spoon, $saucer, $sugarCubesOnTheWindowsill, $lighthouseKeeper);"
	shortLine = "$x = 1;"
)

func upstream(t *testing.T) string {
	ea := t.TempDir()
	write(t, filepath.Join(ea, "src/main/java/zoo/Giraffe.java"), strings.Join([]string{
		`String a = "` + sentence + `";`,
		`String b = "[EA] ` + prefixed + ` %s\\.";`,
		`String c = "zoo\\animals\\GiraffeWithAVeryLongName";`, // identifier-like
		`String d = "%s%s%s%s%s%s%s%s%s%s tiny";`,              // too short once cleaned
	}, "\n"))
	write(t, filepath.Join(ea, "src/main/java/zoo/notes.txt"), `"`+strings.Repeat("ignored words ", 3)+`"`)
	dangling(t, filepath.Join(ea, "src/main/java/zoo/Broken.java"))
	write(t, filepath.Join(ea, "testData/fixtures/zoo/case.php"), "<?php\n"+shortLine+"\n    <warning descr=\"x\">"+longLine+"</warning>\n")
	return ea
}

func TestRunHits(t *testing.T) {
	ea := upstream(t)
	repo := t.TempDir()
	write(t, filepath.Join(repo, "specs/Giraffe.md"), "We say: "+sentence+".\n")
	write(t, filepath.Join(repo, "docs/notes.md"), longLine+"\n"+shortLine+"\n")
	write(t, filepath.Join(repo, "docs/otters.md"), prefixed+" .\n")
	write(t, filepath.Join(repo, "docs/internals/migration.md"), "Upstream prefixes messages with [EA].\n")
	write(t, filepath.Join(repo, "internal/testing/conformance/strip.go"), "// drops [EA]\n")
	write(t, filepath.Join(repo, "internal/inspection/meta/rules.json"), sentence)
	write(t, filepath.Join(repo, "cmd/tool/main.go"), "// PsiWobbleNode\n")
	write(t, filepath.Join(repo, "docs/node_modules/pkg/copy.md"), sentence)
	write(t, filepath.Join(repo, "docs/.vitepress/dist/page.html"), sentence)
	dangling(t, filepath.Join(repo, "internal/broken.go"))

	var stdout, stderr bytes.Buffer
	code := run(repo, []string{"-ea", ea}, &stdout, &stderr)
	want := strings.Join([]string{
		`specs/Giraffe.md: verbatim EA text "` + sentence + `"`,
		`cmd/tool/main.go: forbidden token "PsiWobbleNode"`,
		`docs/internals/migration.md: forbidden token "[EA]"`,
		`docs/notes.md: verbatim EA text "` + longLine[:70] + `…"`,
		`docs/otters.md: verbatim EA text "` + prefixed + ` ."`,
		"cleanroom: 5 hit(s)",
		"",
	}, "\n")
	if code != 1 || stdout.String() != want {
		t.Fatalf("exit %d, stdout:\n%s\nwant:\n%s", code, stdout.String(), want)
	}
}

func TestRunClean(t *testing.T) {
	ea := upstream(t)
	repo := t.TempDir()
	write(t, filepath.Join(repo, "specs/Own.md"), "Our own words about teacups.\n"+shortLine+"\n")
	var stdout, stderr bytes.Buffer
	if code := run(repo, []string{"-ea", ea}, &stdout, &stderr); code != 0 || stdout.String() != "cleanroom: ok\n" {
		t.Fatalf("exit %d, %q", code, stdout.String())
	}
}

func TestRunNoCheckout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	missing := filepath.Join(t.TempDir(), "none")
	if code := run(t.TempDir(), []string{"-ea", missing}, &stdout, &stderr); code != 0 || stdout.String() != "cleanroom: EA checkout not found, skipped\n" {
		t.Fatalf("exit %d, %q", code, stdout.String())
	}
}

func TestRunFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(".", []string{"-h"}, &stdout, &stderr); code != 0 || !strings.Contains(stderr.String(), "-ea") {
		t.Fatalf("-h: exit %d, %q", code, stderr.String())
	}
	if code := run(".", []string{"-nope"}, &stdout, &stderr); code != 2 {
		t.Fatalf("bad flag: exit %d", code)
	}
}

func TestTrunc(t *testing.T) {
	if got := trunc("short"); got != "short" {
		t.Errorf("trunc(short) = %q", got)
	}
	// A 3-byte rune straddling byte 70 is dropped whole.
	s := strings.Repeat("a", 69) + "€tail"
	if got := trunc(s); got != strings.Repeat("a", 69)+"…" {
		t.Errorf("trunc = %q", got)
	}
}
