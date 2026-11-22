package index_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/stubs"
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
	// A world naming its working directory resolves relative paths against
	// it, not against the current one.
	named := index.World{Vendor: abs.Vendor, WorkDir: cwd}
	t.Chdir(t.TempDir())
	if !named.Dependency("vendor/acme/Base.php") || named.Dependency("src/Base.php") {
		t.Error("relative paths resolve against WorkDir")
	}
	t.Chdir(dir)
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

// BenchmarkWorldDependency: a relative file against the absolute vendor
// directory ComposerWorld records, asked per ancestor of every class a
// type check judges.
func BenchmarkWorldDependency(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	w := index.World{Vendor: filepath.Join(cwd, "vendor"), WorkDir: cwd}
	b.ReportAllocs()
	for b.Loop() {
		if w.Dependency("src/Model/Order.php") || !w.Dependency("vendor/acme/lib/Base.php") {
			b.Fatal("wrong answer")
		}
	}
}

// TestExtractAutoloaders: the namespaces registering autoloaders, once
// each; what their callbacks test the class name for, and whether they
// name a phar.
func TestExtractAutoloaders(t *testing.T) {
	src := `<?php
namespace Tool { spl_autoload_register(fn ($c) => 1); \spl_autoload_register(fn ($c) => "phar://{$c}"); }
namespace { spl_autoload_register(fn ($c) => 3); $p = 'phar://' . __DIR__; }`
	fs := index.Extract(syntax.Parse("a.php", []byte(src), syntax.Options{}))
	if !slices.Equal(fs.Autoloaders, []string{`tool\`, ""}) || !fs.Phar {
		t.Errorf("autoloaders %q, phar %v", fs.Autoloaders, fs.Phar)
	}
	for name, src := range map[string]string{
		"comment":         "<?php // spl_autoload_register('phar://x');\n$s = 'none';",
		"phar elsewhere":  "<?php spl_autoload_register(fn ($c) => null); $p = 'phar://x';",
		"not of the file": "<?php spl_autoload_register([Other::class, 'load']); spl_autoload_register('other'); spl_autoload_register($f); spl_autoload_register(); spl_autoload_register(...$a); $g = spl_autoload_register(...);",
		"no class name":   "<?php spl_autoload_register(function () { strpos($x, 'Acme'); });",
		"not first-class": "<?php final class L { function load($c) { strpos($c, 'Acme'); } function reg() { spl_autoload_register($this->load($x)); spl_autoload_register(self::load($x, $y)); spl_autoload_register(load()); spl_autoload_register($this->$m(...)); spl_autoload_register(self::$m(...)); spl_autoload_register($f(...)); } }",
		"other receiver":  "<?php final class L { function load($c) { strpos($c, 'Acme'); } function reg($o) { spl_autoload_register($o->load(...)); spl_autoload_register($this?->load(...)); spl_autoload_register($$v->load(...)); spl_autoload_register([$o, 'load']); spl_autoload_register([parent::class, 'load']); spl_autoload_register([$this, $m]); spl_autoload_register([$this]); } }",
		"other class":     "<?php final class L { static function load($c) { strpos($c, 'Acme'); } } spl_autoload_register(Other::load(...)); spl_autoload_register('Other::load'); spl_autoload_register([Other::class, 'load']); spl_autoload_register([L::{$k}, 'load']); spl_autoload_register(self::load(...)); new class { function load($c) { strpos($c, 'Acme'); } function r() { spl_autoload_register($this->load(...)); } };",
		"other kind":      "<?php function zed($s) { strpos($s, 'Zed'); } final class L { function load($c) { strpos($c, 'Acme'); } } spl_autoload_register(load(...)); spl_autoload_register('load'); spl_autoload_register([L::class, 'zed']);",
		"other argument":  "<?php function load($c) { strpos($c, 'Acme'); } spl_autoload_register(prepend: true, throw: load(...)); spl_autoload_register(fn () => 1, load(...));",
	} {
		fs := index.Extract(syntax.Parse("b.php", []byte(src), syntax.Options{}))
		if fs.Phar || fs.AutoloadPrefixes != nil {
			t.Errorf("%s: prefixes %q, phar %v", name, fs.AutoloadPrefixes, fs.Phar)
		}
	}
	// The prefixes a registered callback tests its class name for, by
	// literal, once each, not too short; not those of other strings.
	for name, src := range map[string]string{
		"closure": `<?php spl_autoload_register(function ($c) {
	if (strpos($c, '\\Acme\\') === 0 || \str_starts_with($c, 'ACME\\') || strpos($c, $x) || strpos($c) || strpos($c, '\\') || strpos($c, '_A') || Foo::strpos($c, 'x') || strpos($c, ...$a) || in_array($c, ['Acme']) || strpos($other, 'Other') || strpos(...$c)) {}
});`,
		"method":          `<?php namespace Tool; final class Loader { public static function load($class) { strpos($class, 'Acme\\'); } } spl_autoload_register([Loader::class, 'load']);`,
		"static string":   `<?php final class Loader { public static function load($class) { strpos($class, 'Acme\\'); } } spl_autoload_register('Loader::load');`,
		"this method":     `<?php function load($s) { strpos($s, 'Zed'); } final class Loader { function load($class) { strpos($class, 'Acme\\'); } function register() { spl_autoload_register($this->load(...)); } }`,
		"self method":     `<?php namespace Tool; final class Loader { static function load($class) { strpos($class, 'Acme\\'); } function register() { spl_autoload_register(self::load(...)); } }`,
		"named class":     `<?php namespace Tool; final class Loader { static function load($class) { strpos($class, 'Acme\\'); } } spl_autoload_register(\Tool\Loader::load(...));`,
		"array this":      `<?php final class Loader { function load($class) { strpos($class, 'Acme\\'); } function register() { spl_autoload_register([$this, 'load']); } }`,
		"array new":       `<?php final class Loader { function load($class) { strpos($class, 'Acme\\'); } static function register() { spl_autoload_register([new static(), 'load']); } }`,
		"array magic":     `<?php final class Loader { function load($class) { strpos($class, 'Acme\\'); } function register() { spl_autoload_register([__CLASS__, 'load']); } }`,
		"array string":    `<?php final class Loader { static function load($class) { strpos($class, 'Acme\\'); } } spl_autoload_register(['\Loader', 'load']);`,
		"named argument":  `<?php function acme_load($class) { strpos($class, 'Acme\\'); } spl_autoload_register(prepend: true, callback: 'acme_load');`,
		"function first":  `<?php namespace Tool; function acme_load($class) { strpos($class, 'Acme\\'); } spl_autoload_register(acme_load(...));`,
		"global fallback": `<?php namespace Tool { spl_autoload_register(acme_load(...)); } namespace { function acme_load($class) { strpos($class, 'Acme\\'); } }`,
		"function":        `<?php function acme_load($class) { str_starts_with($class, 'Acme\\'); } function other($s) { strpos($s, 'Zed'); } spl_autoload_register('acme_load');`,
	} {
		fs := index.Extract(syntax.Parse("l.php", []byte(src), syntax.Options{}))
		if !slices.Equal(fs.AutoloadPrefixes, []string{`acme\`}) {
			t.Errorf("%s: prefixes %q", name, fs.AutoloadPrefixes)
		}
	}
	var many strings.Builder
	many.WriteString("<?php spl_autoload_register(function ($c) {")
	for i := range 20 {
		fmt.Fprintf(&many, " strpos($c, 'Pre%d');", i)
	}
	many.WriteString(" });")
	if fs := index.Extract(syntax.Parse("m.php", []byte(many.String()), syntax.Options{})); len(fs.AutoloadPrefixes) != 16 {
		t.Errorf("%d prefixes kept", len(fs.AutoloadPrefixes))
	}
	ix := index.New(nil)
	ix.Add(index.Extract(syntax.Parse("c.php", []byte(src), syntax.Options{})))
	if top := index.New(ix); top.Symbols("c.php") == nil || top.Symbols("d.php") != nil {
		t.Error("symbols by path")
	}
}
