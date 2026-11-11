package project

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/undefinedclass"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/infer"
)

// What discovery and indexing leave out is unseen by the world: the
// project's excluded directories, the vendor walk's nested test
// directories and files too large to index, so the names they may declare
// are not reported.
func TestWorldSkips(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "composer.json"), `{"require": {"acme/lib": "*"}, "autoload": {"psr-4": {"App\\": "src/"}}}`)
	writeFile(t, filepath.Join(root, "vendor", "composer", "installed.json"), `{"packages": [{"name": "acme/lib", "autoload": {"psr-4": {"Acme\\Lib\\": "src/"}}}]}`)
	writeFile(t, filepath.Join(root, "vendor", "acme", "lib", "src", "Client.php"), `<?php namespace Acme\Lib; class Client {}`)
	writeFile(t, filepath.Join(root, "vendor", "acme", "lib", "src", "Tests", "Helper.php"), `<?php namespace Acme\Lib\Tests; class Helper {}`)
	big := filepath.Join(root, "vendor", "acme", "lib", "src", "Big.php")
	writeFile(t, big, `<?php namespace Acme\Lib; class Big {}`)
	if err := os.Truncate(big, syntax.MaxFileSize+1); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "src", "Legacy", "Foo.php"), `<?php namespace App\Legacy; class Foo {}`)
	main := filepath.Join(root, "src", "main.php")
	writeFile(t, main, `<?php
new \App\Legacy\Foo();
new \Acme\Lib\Tests\Helper();
new \Acme\Lib\Big();
new \Acme\Lib\Client();
new \Acme\Lib\Missing();
new \App\Missing();
`)
	opt := syntax.Options{Version: phpversion.PHP84}
	src := filepath.Join(root, "src")
	p, err := Open([]analysis.Rule{undefinedclass.New()}, Options{Root: root, Paths: []string{src}, Exclude: []string{"Legacy"}, Parse: opt, Analysis: analysis.Config{PHP: phpversion.PHP84, EnableAll: true}})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, result := range p.AnalyzeReport() {
		if result.Path == main {
			for _, f := range result.Findings {
				got = append(got, f.Message)
			}
		}
	}
	want := []string{"Fix the name or install the package declaring Acme\\Lib\\Missing.", "Fix the name or install the package declaring App\\Missing."}
	if !slices.Equal(sortedCopy(got), want) {
		t.Fatalf("reported %q, want %q", got, want)
	}
}

// IndexSourcesSkipping records what the walks leave out: excluded, hidden
// and vendor test directories, declaration-only and other files Composer
// may load, and the custos package; the project walk leaves vendor to the
// vendor walk, and a path it cannot walk counts as skipped.
func TestIndexSourcesSkipping(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{
		"src/A.php", "src/Gen/B.php", "src/.cache/C.php", "src/api.stub.php", "src/old.inc", "src/view.hh", "src/README.md",
		"vendor/acme/lib/src/D.php", "vendor/acme/lib/src/test/E.php", "vendor/janalis/custos/bin/F.php",
	} {
		writeFile(t, filepath.Join(root, filepath.FromSlash(f)), "<?php")
	}
	at := func(f string) string { return filepath.Join(root, filepath.FromSlash(f)) }
	files, skipped := IndexSourcesSkipping(root, nil, []string{root}, []string{"Gen", "vendor"})
	if want := []string{at("src/A.php"), at("vendor/acme/lib/src/D.php")}; !slices.Equal(files, want) {
		t.Fatalf("files %v, want %v", files, want)
	}
	if want := []string{at("src/.cache"), at("src/Gen"), at("vendor/acme/lib/src/test"), at("vendor/janalis/custos")}; !slices.Equal(sortedCopy(skipped.Dirs), want) {
		t.Fatalf("dirs %v, want %v", skipped.Dirs, want)
	}
	if want := []string{at("src/api.stub.php"), at("src/old.inc"), at("src/view.hh")}; !slices.Equal(sortedCopy(skipped.Files), want) {
		t.Fatalf("files %v, want %v", skipped.Files, want)
	}
	if _, skipped := IndexSourcesSkipping(root, nil, []string{at("missing")}, nil); !slices.Contains(skipped.Dirs, at("missing")) {
		t.Fatalf("missing path: %v", skipped.Dirs)
	}
}

// A file too large to analyse declares nothing the index can read, kept
// tree or not, cached or not; a vendor directory the walk cannot read is
// skipped whole.
func TestIndexUnread(t *testing.T) {
	root := t.TempDir()
	big := filepath.Join(root, "Big.php")
	writeFile(t, big, `<?php class Big {}`)
	if err := os.Truncate(big, syntax.MaxFileSize+1); err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpversion.PHP84}
	if _, srcs, trees, unread := buildIndex([]string{big}, 1, opt, infer.Options{}, nil, true); !slices.Equal(unread, []string{big}) || srcs[0] == nil || trees[0] == nil {
		t.Fatalf("kept tree: unread %v", unread)
	}
	cache := OpenSymbolCache(t.TempDir(), root, opt, infer.Options{})
	if _, unread := BuildIndexUnread([]string{big}, opt, infer.Options{}, cache); !slices.Equal(unread, []string{big}) {
		t.Fatalf("cached: unread %v", unread)
	}
	locked := filepath.Join(root, "vendor", "acme")
	writeFile(t, filepath.Join(locked, "A.php"), "<?php")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(locked, 0o755) }()
	if _, skipped := IndexSourcesSkipping(root, nil, nil, nil); !slices.Equal(skipped.Dirs, []string{filepath.Join(root, "vendor")}) {
		t.Fatalf("unreadable vendor: %v", skipped.Dirs)
	}
}

// A snapshot records what its walks leave out, for the world.
func TestSnapshotUnseen(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"src/main.php", "src/old.inc", "src/.cache/A.php", "vendor/janalis/custos/B.php", "vendor/acme/lib/src/Tests/C.php", "vendor/acme/lib/src/D.php"} {
		writeFile(t, filepath.Join(dir, filepath.FromSlash(f)), "<?php")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	s, err := CaptureSnapshot(t.Context(), root, []analysis.Rule{undefinedclass.New()}, []string{"src"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"src/.cache", "vendor/acme/lib/src/Tests", "vendor/janalis/custos"}; !slices.Equal(slices.Compact(sortedCopy(s.unseen.Dirs)), want) {
		t.Fatalf("dirs %v, want %v", s.unseen.Dirs, want)
	}
	if want := []string{"src/old.inc"}; !slices.Equal(slices.Compact(sortedCopy(s.unseen.Files)), want) {
		t.Fatalf("files %v, want %v", s.unseen.Files, want)
	}
}

func sortedCopy(list []string) []string {
	out := slices.Clone(list)
	sort.Strings(out)
	return out
}
