package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func TestClassWord(t *testing.T) {
	for name, want := range map[string]string{"SplObjectStorage": "spl", "PDOStatement": "pdo", "PDO": "pdo", "DateTime": "date", "Zend_Db": "zend", "XML_Parser": "xml", "Yac": "yac", "stdClass": "std"} {
		if got := index.ClassWord(name); got != want {
			t.Errorf("ClassWord(%s) = %s, want %s", name, got, want)
		}
	}
}

func TestWorldLayers(t *testing.T) {
	project := index.New(stubs.Index())
	if project.World().Complete {
		t.Fatal("a fresh index knows nothing of its world")
	}
	project.SetWorld(index.World{Complete: true, Blind: []string{`app\tests\`}})
	project.Freeze()
	file := index.New(project)
	w := file.World()
	if !w.Complete || !w.Unseen(`\App\Tests\Case`) || w.Unseen(`App\Model`) {
		t.Fatalf("layered world: %+v", w)
	}
}

// TestWorldDependency: a file is a dependency's when it lies in the
// configured vendor directory, whatever other directory names it has; an
// unknown vendor directory makes nothing a dependency.
func TestWorldDependency(t *testing.T) {
	w := index.World{Vendor: "./libraries/deps"}
	for file, want := range map[string]bool{
		"libraries/deps/acme/A.php":     true,
		"./libraries/deps/acme/A.php":   true,
		"libraries/deps":                false,
		"libraries/depsx/A.php":         false,
		"vendor/acme/A.php":             false,
		"src/vendor/A.php":              false,
		"other/libraries/deps/acme.php": false,
	} {
		if got := w.Dependency(file); got != want {
			t.Errorf("Dependency(%s) = %v, want %v", file, got, want)
		}
	}
	if (index.World{}).Dependency("vendor/acme/A.php") {
		t.Error("no vendor directory: nothing is a dependency")
	}
}

func TestDynamicAliases(t *testing.T) {
	for _, tc := range []struct {
		src     string
		dynamic bool
	}{
		{"<?php class_alias('A', 'B');", false},
		{"<?php class_alias($original, 'B');", true},
		{"<?php class_alias(...$pair);", true},
		{"<?php function f() {}", false},
	} {
		project := index.New(stubs.Index())
		project.Add(index.Extract(syntax.Parse("a.php", []byte(tc.src), syntax.Options{})))
		project.Freeze()
		// The frozen layer answers from its memo; a layer above asks it.
		if project.DynamicAliases() != tc.dynamic || index.New(project).DynamicAliases() != tc.dynamic {
			t.Fatalf("%s: want %v", tc.src, tc.dynamic)
		}
	}
}

func TestBuiltinPrefixes(t *testing.T) {
	p := index.New(stubs.Index()).BuiltinPrefixes()
	if !p.Words["str"] || !p.Words["apcu"] || p.Words["newext"] {
		t.Fatal("leading words of global builtin functions")
	}
	if !p.Namespaces["mongodb"] || !p.Namespaces["random"] || p.Namespaces["app"] {
		t.Fatal("namespaces of builtins")
	}
}

// TestWorldDependencyPathForms: a file and a vendor directory written in
// different forms, one absolute and the other relative to the working
// directory, are compared as absolute paths. Without a working directory
// they are compared as written.
func TestWorldDependencyPathForms(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	abs := index.World{Vendor: filepath.Join(cwd, "vendor")}
	rel := index.World{Vendor: "vendor"}
	for _, c := range []struct {
		w    index.World
		file string
		want bool
	}{
		{abs, "vendor/acme/lib/Base.php", true},
		{abs, "./vendor/acme/lib/Base.php", true},
		{abs, "src/vendor/Base.php", false},
		{abs, filepath.Join(cwd, "vendor", "acme", "lib", "Base.php"), true},
		{rel, filepath.Join(cwd, "vendor", "acme", "lib", "Base.php"), true},
		{rel, filepath.Join(cwd, "src", "Base.php"), false},
		{rel, "vendor/acme/lib/Base.php", true},
	} {
		if got := c.w.Dependency(c.file); got != c.want {
			t.Errorf("World{Vendor: %s}.Dependency(%s) = %v, want %v", c.w.Vendor, c.file, got, c.want)
		}
	}
	// getcwd fails on macOS for an unreadable directory and on Linux for a
	// removed one: do both.
	gone := filepath.Join(dir, "gone")
	if err := os.Mkdir(gone, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(gone)
	t.Setenv("PWD", "")
	if err := os.Chmod(gone, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	if rel.Dependency(filepath.Join(cwd, "vendor", "A.php")) || !rel.Dependency("vendor/A.php") {
		t.Error("without a working directory: paths compare as written")
	}
}
