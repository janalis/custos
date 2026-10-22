package project

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/securityadvisories"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// TestIndexScope: the project paths are indexed besides the analysed files,
// each file once, and vendor follows the project excludes.
func TestIndexScope(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "src", "app.php")
	other := filepath.Join(root, "src", "other.php")
	lib := filepath.Join(root, "vendor", "lib", "Lib.php")
	for _, f := range []string{app, other, lib, filepath.Join(root, "src", "gen", "G.php"), filepath.Join(root, "vendor", "lib", "gen", "G.php"), filepath.Join(root, "vendor", "janalis", "custos", "C.php"), filepath.Join(root, "vendor", "lib", "node_modules", "N.php")} {
		writeFile(t, f, "<?php\n")
	}
	t.Chdir(root)
	// The analysed file is named relative to the working directory and the
	// project path holds vendor: neither is listed twice.
	got := IndexSources(root, []string{filepath.Join("src", "app.php")}, []string{root}, []string{"gen"})
	want := []string{filepath.Join("src", "app.php"), other, lib}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sources:\n got %v\nwant %v", got, want)
	}
	// An unreadable project path still indexes the rest.
	if got := IndexSources(root, []string{app}, []string{filepath.Join(root, "missing")}, []string{"gen"}); len(got) != 2 || got[1] != lib {
		t.Fatalf("missing path: %v", got)
	}
}

// TestDiscoverSymlinks: a link is listed only when it leads to a regular
// file inside the scanned path.
func TestDiscoverSymlinks(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	src := filepath.Join(root, "src")
	real := filepath.Join(src, "real.php")
	writeFile(t, real, "<?php\n")
	writeFile(t, filepath.Join(outside, "out.php"), "<?php\n")
	if err := os.Mkdir(filepath.Join(src, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	for link, target := range map[string]string{"in.php": real, "out.php": filepath.Join(outside, "out.php"), "dangling.php": filepath.Join(src, "gone"), "dir.php": filepath.Join(src, "dir")} {
		if err := os.Symlink(target, filepath.Join(src, link)); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Discover([]string{src}, nil)
	if err != nil || !reflect.DeepEqual(got, []string{filepath.Join(src, "in.php"), real}) {
		t.Fatalf("got %v, %v", got, err)
	}
	if linkInside(filepath.Join(root, "missing"), filepath.Join(src, "in.php")) {
		t.Fatal("unresolvable scanned path")
	}
}

// TestRequestedFilesNotIndexed: a file a rule requests by name is analysed
// but not indexed as PHP.
func TestRequestedFilesNotIndexed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "composer.json"), `{}`)
	writeFile(t, filepath.Join(root, "src", "a.php"), "<?php\nfunction a() {}\n")
	old := extract
	t.Cleanup(func() { extract = old })
	var extracted []string
	var mu sync.Mutex
	extract = func(path string, src []byte, opt syntax.Options, options infer.Options) *index.FileSymbols {
		mu.Lock()
		extracted = append(extracted, filepath.Base(path))
		mu.Unlock()
		return old(path, src, opt, options)
	}
	opt := syntax.Options{Version: phpversion.Default}
	p, err := Open([]analysis.Rule{securityadvisories.New(), untrustedshellcommand.New()}, Options{Root: root, Paths: []string{root}, Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: opt})
	if err != nil {
		t.Fatal(err)
	}
	_, sources := p.prepare()
	if len(p.Files) != 2 || sources[0] != nil || sources[1] == nil || !reflect.DeepEqual(extracted, []string{"a.php"}) {
		t.Fatalf("files %v sources %q extracted %v", p.Files, sources, extracted)
	}
	if results := p.AnalyzeReport(); len(results) != 2 {
		t.Fatalf("results %d", len(results))
	}
}
