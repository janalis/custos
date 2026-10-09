package main

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Invented stub files.
const bakery = `<?php
/**
 * @param array{flour: int, eggs?: int} $recipe
 * @return array{crumb: string, crust?: bool}
 */
function bake_loaf(array $recipe, $oven) {}

class Oven {
    /** @var array{dial: int} */
    public $settings;

    /**
     * @param array{tray: int} $rack
     * @return array{heat: int}
     */
    public function preheat(array $rack, int $minutes) {}
}

const OVEN_MAX = 300;
`

func stubTree(t *testing.T) string {
	src := filepath.Join(t.TempDir(), "fake-phpstorm-stubs-abc")
	writeFile(t, filepath.Join(src, "bakery/bakery.php"), bakery)
	writeFile(t, filepath.Join(src, "broken/broken.php"), "<?php function half_baked( {\n")
	writeFile(t, filepath.Join(src, "README.txt"), "not php")
	writeFile(t, filepath.Join(src, "PhpStormStubsMap.php"), "<?php class Skipped1 {}\n")
	for _, d := range []string{"tests", "meta", ".github", "vendor"} {
		writeFile(t, filepath.Join(src, d, "skip.php"), "<?php class Skipped2 {}\n")
	}
	return src
}

func load(t *testing.T, path string) []*index.FileSymbols {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var all []*index.FileSymbols
	if err := gob.NewDecoder(zr).Decode(&all); err != nil {
		t.Fatal(err)
	}
	return all
}

func TestRun(t *testing.T) {
	src := stubTree(t)
	outDir := t.TempDir()
	out := filepath.Join(outDir, "stubs.gob.gz")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-src", src, "-out", out, "-rev", "r42"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if got := stdout.String(); !strings.HasPrefix(got, "stubs: 2 files (1 with parse errors), 1 classes, 2 functions, 1 constants -> "+out+" (") {
		t.Errorf("stdout = %q", got)
	}
	if v, _ := os.ReadFile(filepath.Join(outDir, "VERSION")); string(v) != "r42\n" {
		t.Errorf("VERSION = %q", v)
	}
	all := load(t, out)
	if len(all) != 2 || all[0].Path != "stubs/bakery/bakery.php" || all[1].Path != "stubs/broken/broken.php" {
		t.Fatalf("files = %+v", all)
	}
	fs := all[0]
	if len(fs.Functions) != 1 || len(fs.Classes) != 1 || len(fs.Constants) != 1 {
		t.Fatalf("symbols = %+v", fs)
	}
	fn := fs.Functions[0]
	c := fs.Classes[0]
	if fn.File != fs.Path || c.File != fs.Path {
		t.Errorf("file not recorded: %q %q", fn.File, c.File)
	}
	var m *index.Method
	for _, x := range c.Methods {
		m = x
	}
	var p *index.Property
	for _, x := range c.Props {
		p = x
	}
	if m == nil || p == nil {
		t.Fatalf("class = %+v", c)
	}
	// Array shapes are dropped from every doc type; empty ones stay empty.
	for what, s := range map[string]string{
		"function return": fn.DocReturn, "function param": fn.Params[0].DocType,
		"method return": m.DocReturn, "method param": m.Params[0].DocType, "property": p.DocType,
	} {
		if s == "" || strings.Contains(s, "{") {
			t.Errorf("%s doc type = %q", what, s)
		}
	}
	if fn.Params[1].DocType != "" {
		t.Errorf("undocumented param got %q", fn.Params[1].DocType)
	}
}

func TestRunDefaultSource(t *testing.T) {
	src := stubTree(t)
	old := defaultSrc
	defaultSrc = filepath.Join(filepath.Dir(src), "*phpstorm-stubs*")
	t.Cleanup(func() { defaultSrc = old })
	out := filepath.Join(t.TempDir(), "s.gob.gz")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if v, _ := os.ReadFile(filepath.Join(filepath.Dir(out), "VERSION")); string(v) != "fake-phpstorm-stubs-abc\n" {
		t.Errorf("VERSION = %q", v)
	}
	// No checkout found.
	defaultSrc = filepath.Join(t.TempDir(), "*phpstorm-stubs*")
	stderr.Reset()
	if code := run([]string{"-out", out}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "genstubs: no -src given and no ") {
		t.Fatalf("exit %d: %q", code, stderr.String())
	}
}

func TestRunErrors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-h"}, &stdout, &stderr); code != 0 {
		t.Errorf("-h: exit %d", code)
	}
	if code := run([]string{"-bogus"}, &stdout, &stderr); code != 2 {
		t.Errorf("bad flag: exit %d", code)
	}
	tmp := t.TempDir()
	// Missing source tree.
	if code := run([]string{"-src", filepath.Join(tmp, "none"), "-out", filepath.Join(tmp, "o.gz")}, &stdout, &stderr); code != 1 {
		t.Errorf("missing src: exit %d", code)
	}
	// Unreadable stub (dangling symlink).
	src := filepath.Join(tmp, "src")
	writeFile(t, filepath.Join(src, "ok.php"), "<?php\n")
	if err := os.Symlink(filepath.Join(src, "nowhere"), filepath.Join(src, "gone.php")); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"-src", src, "-out", filepath.Join(tmp, "o.gz")}, &stdout, &stderr); code != 1 {
		t.Errorf("unreadable stub: exit %d", code)
	}
	// Unwritable output.
	src2 := filepath.Join(tmp, "src2")
	writeFile(t, filepath.Join(src2, "ok.php"), "<?php\n")
	stderr.Reset()
	if code := run([]string{"-src", src2, "-out", filepath.Join(tmp, "no/dir/o.gz")}, &stdout, &stderr); code != 1 || !strings.HasPrefix(stderr.String(), "genstubs: ") {
		t.Errorf("unwritable out: exit %d, %q", code, stderr.String())
	}
}

func TestNoShapes(t *testing.T) {
	// An unparsable (over-long) doc type is kept as written.
	long := strings.Repeat("x", types.MaxDocTypeLen+1)
	s := long
	noShapes(&s)
	if s != long {
		t.Errorf("unknown doc type rewritten")
	}
	s = "array{a: int}|null"
	noShapes(&s)
	if strings.Contains(s, "{") {
		t.Errorf("shape kept: %q", s)
	}
}
