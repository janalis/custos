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

// loaded returns the class name prefixes the autoloaders of the first of
// files load, with those of the callbacks the files declare, and whether
// one names a phar.
func loaded(files ...string) (prefixes []string, phar bool) {
	ix := index.New(stubs.Index())
	var fs *index.FileSymbols
	for i, src := range files {
		f := index.Extract(syntax.Parse(fmt.Sprintf("f%d.php", i), []byte(src), syntax.Options{}))
		ix.Add(f)
		if i == 0 {
			fs = f
		}
	}
	prefixes, phar = fs.AutoloadPrefixes, fs.Phar
	for _, k := range fs.AutoloadCallbacks {
		if l := ix.LoaderFacts(k); l != nil {
			prefixes = append(prefixes, l.Prefixes...)
			phar = phar || l.Phar
		}
	}
	return prefixes, phar
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
	// BODY stands for a loader testing for Acme\, which may hand the
	// include to another function.
	const l = `{ if (strpos($c, 'Acme\\') === 0) { self::includeFile($c); } }`
	for name, src := range map[string]string{
		"comment":         "<?php // spl_autoload_register('phar://x');\n$s = 'none';",
		"phar elsewhere":  "<?php spl_autoload_register(fn ($c) => null); $p = 'phar://x';",
		"not of the file": "<?php spl_autoload_register([Other::class, 'load']); spl_autoload_register('other'); spl_autoload_register($f); spl_autoload_register(); spl_autoload_register(...$a); $g = spl_autoload_register(...);",
		"no class name":   "<?php spl_autoload_register(function () { strpos($x, 'Acme'); });",
		"not first-class": "<?php final class L { function load($c) BODY function reg() { spl_autoload_register($this->load($x)); spl_autoload_register(self::load($x, $y)); spl_autoload_register(load()); spl_autoload_register($this->$m(...)); spl_autoload_register(self::$m(...)); spl_autoload_register($f(...)); } }",
		"other receiver":  "<?php final class L { function load($c) BODY function reg($o) { spl_autoload_register($o->load(...)); spl_autoload_register($this?->load(...)); spl_autoload_register($$v->load(...)); spl_autoload_register([$o, 'load']); spl_autoload_register([parent::class, 'load']); spl_autoload_register([$this, $m]); spl_autoload_register([$this]); } }",
		"other class":     "<?php final class L { static function load($c) BODY } spl_autoload_register(Other::load(...)); spl_autoload_register('Other::load'); spl_autoload_register([Other::class, 'load']); spl_autoload_register([L::{$k}, 'load']); spl_autoload_register(self::load(...)); spl_autoload_register('parent::load'); new class { function load($c) BODY function r() { spl_autoload_register($this->load(...)); } };",
		"other kind":      "<?php function zed($c) { if (strpos($c, 'Zed') === 0) { require 'z.php'; } } final class L { function load($c) BODY } spl_autoload_register(load(...)); spl_autoload_register('load'); spl_autoload_register([L::class, 'zed']);",
		"other argument":  "<?php function load($c) BODY spl_autoload_register(prepend: true, throw: load(...)); spl_autoload_register(fn () => 1, load(...));",
		"not a loader":    "<?php $p = 'phar://x'; function load() { strpos($c, 'Acme'); } function other($c) { strpos($x, 'Acme'); } function third($c) { require $c; $f = function ($c) { strpos($c, 'Acme'); return 'phar:// unsupported'; }; } spl_autoload_register('load'); spl_autoload_register('other'); spl_autoload_register('third'); new class { function load($c) { return 'phar://x'; } };",
		"no method body":  "<?php interface I { function load($c); } /** @method void load(string $c) */ abstract class A implements I {} spl_autoload_register([A::class, 'load']); spl_autoload_register([A::class, 'missing']);",
		"parent cycle":    "<?php class A extends B {} class B extends A {} spl_autoload_register([A::class, 'load']);",
		"builtin":         "<?php spl_autoload_register('strlen');",
		"computed":        "<?php spl_autoload_register(__DIR__ . '/load'); spl_autoload_register(__NAMESPACE__ . $f); spl_autoload_register($a . 'load'); spl_autoload_register(__NAMESPACE__ + 'load');",
	} {
		src = strings.ReplaceAll(src, " BODY ", " "+l+" ")
		if prefixes, phar := loaded(src); phar || prefixes != nil {
			t.Errorf("%s: prefixes %q, phar %v", name, prefixes, phar)
		}
	}
	// The prefixes a registered callback tests its class name for, by
	// literal, once each, not too short; not those of other strings.
	for name, src := range map[string]string{
		"closure": `<?php spl_autoload_register(function ($c) {
	if (strpos($c, '\\Acme\\') === 0 || \str_starts_with($c, 'ACME\\') || strpos($c, $x) || strpos($c) || strpos($c, '\\') || strpos($c, '_A') || strpos($c, '\\\\\\\\x') || Foo::strpos($c, 'x') || strpos($c, ...$a) || in_array($c, ['Acme']) || strpos($other, 'Other') || strpos(...$c)) {}
});`,
		"method":          `<?php namespace Tool; final class Loader { public static function load($c) BODY } spl_autoload_register([Loader::class, 'load']);`,
		"static string":   `<?php final class Loader { public static function load($c) BODY } spl_autoload_register('Loader::load');`,
		"this method":     `<?php function load($c) { if (strpos($c, 'Zed') === 0) { require 'z.php'; } } final class Loader { function load($c) BODY function register() { spl_autoload_register($this->load(...)); } }`,
		"self method":     `<?php namespace Tool; final class Loader { static function load($c) BODY function register() { spl_autoload_register(self::load(...)); } }`,
		"self string":     `<?php final class Loader { static function load($c) BODY function register() { spl_autoload_register('self::load'); spl_autoload_register(['static', 'load']); } }`,
		"named class":     `<?php namespace Tool; final class Loader { static function load($c) BODY } spl_autoload_register(\Tool\Loader::load(...));`,
		"array this":      `<?php final class Loader { function load($c) BODY function register() { spl_autoload_register([$this, 'load']); } }`,
		"array new":       `<?php final class Loader { function load($c) BODY static function register() { spl_autoload_register([new static(), 'load']); } }`,
		"array magic":     `<?php final class Loader { function load($c) BODY function register() { spl_autoload_register([__CLASS__, 'load']); } }`,
		"array string":    `<?php final class Loader { static function load($c) BODY } spl_autoload_register(['\Loader', 'load']);`,
		"parent":          `<?php namespace Tool; class Base { static function load($c) BODY } final class Loader extends Base { function register() { spl_autoload_register(parent::load(...)); spl_autoload_register([parent::class, 'load']); spl_autoload_register(['parent', 'load']); } }`,
		"trait":           `<?php trait T { function load($c) BODY } final class Loader { use T; function register() { spl_autoload_register([$this, 'load']); } }`,
		"nested trait":    `<?php trait T2 { function load($c) BODY } trait T { use T2; } class P { function load($c) { strpos($c, 'Zed\\'); } } final class Loader extends P { use T; function register() { spl_autoload_register([$this, 'load']); } }`,
		"trait alias":     `<?php trait T { function boot($c) BODY } final class Loader { use T { boot as load; } function register() { spl_autoload_register([$this, 'load']); } }`,
		"tagged subclass": `<?php class Base { function load($c) BODY } /** @method void load(string $c) */ final class Loader extends Base { function register() { spl_autoload_register([$this, 'load']); } }`,
		"named argument":  `<?php function acme_load($c) BODY spl_autoload_register(prepend: true, callback: 'acme_load');`,
		"function first":  `<?php namespace Tool; function acme_load($c) BODY spl_autoload_register(acme_load(...));`,
		"global fallback": `<?php namespace Tool { spl_autoload_register(acme_load(...)); } namespace { function acme_load($c) BODY }`,
		"namespace":       `<?php namespace Tool; function acme_load($c) BODY spl_autoload_register(__NAMESPACE__ . '\acme_load');`,
		"function":        `<?php function acme_load($c) { str_starts_with($c, 'Acme\\'); } function other($s) { strpos($s, 'Zed'); } spl_autoload_register('acme_load');`,
	} {
		if prefixes, _ := loaded(strings.ReplaceAll(src, " BODY ", " "+l+" ")); !slices.Equal(prefixes, []string{`acme\`}) {
			t.Errorf("%s: prefixes %q", name, prefixes)
		}
	}
	// A callback declared elsewhere: in another file, or by a parent
	// class; its phar only when it names one, with or without a prefix.
	for name, files := range map[string][]string{
		"other file": {
			`<?php spl_autoload_register([Acme\Loader::class, 'load']);`,
			`<?php namespace Acme; final class Loader { static function load($class) { if (strpos($class, 'Acme\\') === 0) { require "phar://acme.phar/{$class}.php"; } } static function other($c) { strpos($c, 'Zed'); } }`,
		},
		"trait alias": {
			`<?php final class Loader { use Acme\Boot { Acme\Boot::boot as protected load; } function register() { spl_autoload_register([$this, 'load']); } }`,
			`<?php namespace Acme; trait Boot { function boot($c) { if (strpos($c, 'Acme\\') === 0) { require "phar://acme.phar/{$c}.php"; } } }`,
		},
		"inherited": {
			`<?php final class Loader extends Base { function register() { spl_autoload_register([$this, 'load']); } }`,
			`<?php class Base { function load($c) { if (strpos($c, 'Acme\\') === 0) { require "$c.php"; } } function path() { return 'phar://x'; } } class Root {}`,
		},
	} {
		prefixes, phar := loaded(files...)
		if !slices.Equal(prefixes, []string{`acme\`}) || phar != (name != "inherited") {
			t.Errorf("%s: prefixes %q, phar %v", name, prefixes, phar)
		}
	}
	if prefixes, phar := loaded(`<?php namespace PHPStan; function load($c) { require 'phar://x.phar/' . $c . '.php'; } spl_autoload_register('PHPStan\load');`); prefixes != nil || !phar {
		t.Errorf("phar only: prefixes %q, phar %v", prefixes, phar)
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
	many.Reset()
	many.WriteString("<?php spl_autoload_register('load'); function load($c) {")
	for i := range 20 {
		fmt.Fprintf(&many, " strpos($c, 'Pre%d'); strpos($c, 'Pre%d');", i, i)
		fmt.Fprintf(&many, " spl_autoload_register('f%d');", i)
	}
	many.WriteString(" }")
	if prefixes, _ := loaded(many.String()); len(prefixes) != 16 {
		t.Errorf("%d loader prefixes kept", len(prefixes))
	}
	if fs := index.Extract(syntax.Parse("m.php", []byte(many.String()), syntax.Options{})); len(fs.AutoloadCallbacks) != 16 {
		t.Errorf("%d callbacks kept", len(fs.AutoloadCallbacks))
	}
	ix := index.New(nil)
	ix.Add(index.Extract(syntax.Parse("c.php", []byte(src), syntax.Options{})))
	if top := index.New(ix); top.Symbols("c.php") == nil || top.Symbols("d.php") != nil {
		t.Error("symbols by path")
	}
}
