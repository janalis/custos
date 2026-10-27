package project

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverWithPatterns(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.php", "composer.json", "composer.lock", "sub/b.php", "sub/composer.json", "vendor/x/composer.json"} {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rel := func(files []string) []string {
		var out []string
		for _, f := range files {
			r, _ := filepath.Rel(dir, f)
			out = append(out, filepath.ToSlash(r))
		}
		return out
	}
	got, err := Discover([]string{dir}, []string{"vendor"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.php", "sub/b.php"}; !reflect.DeepEqual(rel(got), want) {
		t.Fatalf("Discover: got %v want %v", rel(got), want)
	}
	got, err = DiscoverWith([]string{dir}, []string{"vendor"}, []string{"composer.json"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.php", "composer.json", "sub/b.php", "sub/composer.json"}; !reflect.DeepEqual(rel(got), want) {
		t.Fatalf("DiscoverWith: got %v want %v", rel(got), want)
	}
}

// Hidden directories below a scanned path (tool state, generated code) are
// left out; a scanned path that is itself hidden is walked.
func TestDiscoverSkipsHiddenDirectories(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.php", ".phpunit.cache/container/c.php", "src/.cache/x.php", "src/b.php", ".config/php/c.php"} {
		writeFile(t, filepath.Join(dir, f), "<?php")
	}
	got, err := Discover([]string{dir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var rel []string
	for _, f := range got {
		r, _ := filepath.Rel(dir, f)
		rel = append(rel, filepath.ToSlash(r))
	}
	if want := []string{"a.php", "src/b.php"}; !reflect.DeepEqual(rel, want) {
		t.Fatalf("Discover: got %v want %v", rel, want)
	}
	named, err := Discover([]string{filepath.Join(dir, ".config")}, nil)
	if err != nil || len(named) != 1 {
		t.Fatalf("a hidden scanned path: %v %v", named, err)
	}
}

// A file named on its own and again through its directory is listed once.
func TestDiscoverListsFilesOnce(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "src/a.php"), "<?php")
	got, err := Discover([]string{filepath.Join(dir, "src") + "/./a.php", filepath.Join(dir, "src"), filepath.Join(dir, "src")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("Discover: got %v, want one file", got)
	}
}

// Declaration files below a scanned path are left out; a named one is kept.
func TestDiscoverSkipsDeclarationFiles(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.php", "ext/hash.stub.php", "mod/mod.api.php", "mod/mod.module.php"} {
		writeFile(t, filepath.Join(dir, f), "<?php")
	}
	got, err := Discover([]string{dir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var rel []string
	for _, f := range got {
		r, _ := filepath.Rel(dir, f)
		rel = append(rel, filepath.ToSlash(r))
	}
	if want := []string{"a.php", "mod/mod.module.php"}; !reflect.DeepEqual(rel, want) {
		t.Fatalf("Discover: got %v want %v", rel, want)
	}
	named, err := Discover([]string{filepath.Join(dir, "ext/hash.stub.php")}, nil)
	if err != nil || len(named) != 1 {
		t.Fatalf("a named declaration file: %v %v", named, err)
	}
}
