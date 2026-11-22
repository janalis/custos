package project

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/project/config"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/infer"
)

// worldFiles are the files of a project whose world is complete.
func worldFiles() map[string]string {
	return map[string]string{
		"composer.json": `{
			"require": {"php": ">=8.1", "ext-json": "*", "acme/http": "^1", "psr/log-implementation": "*"},
			"require-dev": {"acme/test": "^1"},
			"suggest": {"ext-intl": "for locales", "acme/http": "already required"},
			"autoload": {"psr-4": {"App\\": "src/", "App\\Extra\\": ["lib/", "src/Extra"]}, "files": ["src/helpers.php"]},
			"autoload-dev": {"psr-4": {"App\\Tests\\": "tests/"}, "classmap": ["src/legacy"]}
		}`,
		"vendor/composer/installed.json": `{"packages": [
			{"name": "acme/http", "install-path": "../acme/http", "autoload": {"psr-4": {"Acme\\Http\\": "src/", "Acme\\Http\\Testing\\": "tests/"}}, "provide": {"psr/log-implementation": "1.0"}},
			{"name": "acme/test", "autoload": {"classmap": ["src/"]}, "replace": {"acme/old": "*"}, "require": {"ext-pdo_mysql": "*", "ext-zend-opcache": "*", "ext-mysqlnd": "*", "ext-SimpleXML": "*", "ext-opcache": "*"}},
			{"name": "acme/assets"}
		], "dev": true}`,
	}
}

// symbols extracts the declarations of the PHP files of files, as the
// index holds them.
func symbols(files map[string]string) func(string) *index.FileSymbols {
	return func(name string) *index.FileSymbols {
		src, ok := files[name]
		if !ok || !strings.HasSuffix(name, ".php") {
			return nil
		}
		return index.Extract(syntax.Parse(name, []byte(src), syntax.Options{}))
	}
}

func reader(files map[string]string) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		if s, ok := files[name]; ok {
			return []byte(s), nil
		}
		return nil, os.ErrNotExist
	}
}

// addPackage lists one more installed package.
func addPackage(files map[string]string, pkg string) {
	installed := files["vendor/composer/installed.json"]
	files["vendor/composer/installed.json"] = strings.Replace(installed, `"packages": [`, `"packages": [`+pkg+",", 1)
}

// worldSkipped is what the walks of the world left out.
var worldSkipped = Skipped{Dirs: []string{"vendor/acme/http/tests", "vendor/acme/http/test"}}

var worldIndexed = []string{"src/A.php", "src/helpers.php", "vendor/acme/http/src/Client.php", "vendor/acme/test/src/Case.php"}

func TestComposerWorldComplete(t *testing.T) {
	w := composerWorld(reader(worldFiles()), symbols(worldFiles()), []string{"src", "lib"}, worldIndexed, worldSkipped)
	if !w.Complete || !slices.Equal(w.Blind, []string{`app\tests\`, `acme\http\testing\`}) {
		t.Fatalf("world: %+v", w)
	}
	if !w.Unseen(`\App\Tests\Unit\CaseTest`) || w.Unseen(`App\Model`) {
		t.Fatal("blind prefixes")
	}
	// Required extensions, as the stubs name them.
	if got := slices.Sorted(maps.Keys(w.Extensions)); !slices.Equal(got, []string{"intl", "json", "mysqlnd", "pdo", "simplexml", "zend-opcache"}) {
		t.Errorf("extensions: %q", got)
	}
	// Composer 1 writes a bare list, and the whole project in scope.
	files := worldFiles()
	files["vendor/composer/installed.json"] = `[
		{"name": "acme/http", "autoload": {"psr-4": {"Acme\\Http\\": "src/"}}, "provide": {"psr/log-implementation": "1.0"}},
		{"name": "acme/test"}
	]`
	if w := composerWorld(reader(files), symbols(files), []string{"."}, worldIndexed, worldSkipped); !w.Complete || len(w.Blind) != 0 {
		t.Fatalf("composer 1: %+v", w)
	}
}

// TestComposerWorldPackageAutoloader: a package's files entry registering
// an autoloader (PHPStan's, loading from its phar) makes the namespace it
// declares unseen; one without autoloader, or unreadable, changes nothing.
func TestComposerWorldPackageAutoloader(t *testing.T) {
	for src, blind := range map[string]bool{
		"<?php namespace PHPStan; spl_autoload_register(function ($c) {});\n":                                         true,
		"<?php\nspl_autoload_register(function ($c) {}); // a global one hides nothing\n":                             false,
		"<?php declare(strict_types = 1);\n\nnamespace PHPStan;\n\nspl_autoload_register([Loader::class, 'load']);\n": true,
		"<?php\nnamespace PHPStan;\n// calls spl_autoload_register() on 'phar://' in a comment only\n":                false,
		"<?php\nnamespace Acme\\Http;\nfunction helper() {}\n":                                                        false,
		"": false,
	} {
		files := worldFiles()
		addPackage(files, `{"name": "acme/z", "install-path": "../acme/http", "autoload": {"files": ["src/Client.php"]}}`)
		if src != "" {
			files["vendor/acme/http/src/Client.php"] = src
		}
		w := composerWorld(reader(files), symbols(files), []string{"src", "lib"}, worldIndexed, worldSkipped)
		if !w.Complete || w.UnseenClass(`PHPStan\Type\ObjectType`) != blind || w.Unseen(`PHPStan\helper`) || len(w.Hosts) != 0 {
			t.Errorf("%q: %+v", src, w)
		}
	}
	// A global autoloader hides only the class name prefixes it tests.
	files := worldFiles()
	addPackage(files, `{"name": "acme/z", "install-path": "../acme/http", "autoload": {"files": ["src/Client.php"]}}`)
	files["vendor/acme/http/src/Client.php"] = "<?php\nspl_autoload_register(function ($c) { if (strpos($c, 'RectorPrefix') === 0 || str_starts_with($c, 'Rector\\\\')) { require 'x.php'; } });\n"
	w := composerWorld(reader(files), symbols(files), []string{"src", "lib"}, worldIndexed, worldSkipped)
	if !w.Complete || !w.UnseenClass(`RectorPrefix202611\Foo`) || !w.UnseenClass(`Rector\Config\RectorConfig`) || w.UnseenClass(`Rectangle`) || w.UnseenClass(`App\Model`) || w.Unseen(`Rector\helper`) {
		t.Errorf("prefixes: %+v", w)
	}
	// Loading from a phar, the package hosts code importing from it.
	files = worldFiles()
	addPackage(files, `{"name": "acme/z", "install-path": "../acme/http", "autoload": {"files": ["src/Client.php"]}}`)
	files["vendor/acme/http/src/Client.php"] = "<?php\nnamespace PHPStan;\nspl_autoload_register(function ($c) { require 'phar://' . __DIR__ . '/x.phar/a.php'; });\n"
	if w = composerWorld(reader(files), symbols(files), []string{"src", "lib"}, worldIndexed, worldSkipped); !w.Complete || !slices.Equal(w.Hosts, []string{`phpstan\`}) {
		t.Errorf("phar: %+v", w)
	}
}

func TestComposerWorldIncomplete(t *testing.T) {
	for name, edit := range map[string]func(map[string]string){
		"no composer.json":         func(f map[string]string) { delete(f, "composer.json") },
		"broken composer.json":     func(f map[string]string) { f["composer.json"] = "{" },
		"not installed":            func(f map[string]string) { delete(f, "vendor/composer/installed.json") },
		"broken installed":         func(f map[string]string) { f["vendor/composer/installed.json"] = "{" },
		"missing requirement":      func(f map[string]string) { f["composer.json"] = `{"require": {"acme/other": "*"}}` },
		"missing dev":              func(f map[string]string) { f["composer.json"] = `{"require-dev": {"acme/other": "*"}}` },
		"missing suggestion":       func(f map[string]string) { f["composer.json"] = `{"suggest": {"acme/other": "*"}}` },
		"extension the stubs lack": func(f map[string]string) { f["composer.json"] = `{"require": {"ext-phalcon": "*"}}` },
		"suggested extension":      func(f map[string]string) { f["composer.json"] = `{"suggest": {"ext-yac": "*"}}` },
		"package extension":        func(f map[string]string) { addPackage(f, `{"name": "acme/x", "require": {"ext-sphinx": "*"}}`) },
		"fallback directory":       func(f map[string]string) { f["composer.json"] = `{"autoload": {"psr-4": {"": "other/"}}}` },
		"class name prefix":        func(f map[string]string) { f["composer.json"] = `{"autoload": {"psr-0": {"Acme_": "other/"}}}` },
		"classmap out":             func(f map[string]string) { f["composer.json"] = `{"autoload": {"classmap": ["other/"]}}` },
		"files out":                func(f map[string]string) { f["composer.json"] = `{"autoload-dev": {"files": ["other/x.php"]}}` },
		"package not indexed":      func(f map[string]string) { addPackage(f, `{"name": "acme/lost", "autoload": {"files": ["f.php"]}}`) },
		"package fallback in tests": func(f map[string]string) {
			addPackage(f, `{"name": "acme/x", "install-path": "../acme/http", "autoload": {"psr-4": {"": "tests/"}}}`)
		},
		"package classmap in tests": func(f map[string]string) {
			addPackage(f, `{"name": "acme/y", "install-path": "../acme/http", "autoload": {"classmap": ["test/"]}}`)
		},
	} {
		t.Run(name, func(t *testing.T) {
			files := worldFiles()
			edit(files)
			if w := composerWorld(reader(files), symbols(files), []string{"src", "lib"}, worldIndexed, worldSkipped); w.Complete {
				t.Fatalf("complete world: %+v", w)
			}
		})
	}
}

// What the walks and the indexing skipped below an autoload directory is
// blind (a PSR-4 namespace, a class name, a whole PSR-0 prefix) or, below
// a fallback, a classmap or a files entry, makes the world incomplete.
func TestComposerWorldGaps(t *testing.T) {
	files := worldFiles()
	addPackage(files, `{"name": "acme/old", "autoload": {"psr-0": {"Acme\\Old\\": "lib/"}}}`)
	indexed := append(slices.Clone(worldIndexed), "vendor/acme/old/lib/Acme/Old/A.php")
	skipped := Skipped{
		Dirs: append(slices.Clone(worldSkipped.Dirs), "src/Legacy", "vendor/acme/http/src/Tests", "vendor/acme/old/lib/Acme/Old/Fixtures"),
		// Big.php and Huge.php were too large to index; a declaration-only
		// file names no class PSR-4 can load.
		Files: []string{"vendor/acme/http/src/Big.php", "src/Model/Huge.php", "src/Model/Other.php", "src/api.stub.php"},
	}
	w := composerWorld(reader(files), symbols(files), []string{"src", "lib"}, indexed, skipped)
	want := []string{`app\legacy\`, `app\tests\`, `acme\old\`, `acme\http\tests\`, `acme\http\testing\`}
	if !w.Complete || !slices.Equal(sortedCopy(w.Blind), sortedCopy(want)) ||
		!slices.Equal(sortedCopy(w.BlindExact), []string{`acme\http\`, `app\model\`}) {
		t.Fatalf("world: %+v", w)
	}
	// A skipped file hides its namespace and those below it: once loaded it
	// may declare more than the class it is named after, anywhere there.
	for fqn, unseen := range map[string]bool{
		`Acme\Http\Big`: true, `Acme\Http\BigQuery`: true, `Acme\Http\helper`: true,
		`Acme\Http\Client\Pool`: true, `Acme\Http\Tests\Case`: true,
		`App\Model\User`: true, `App\Model\Cast\Json`: true, `App\Modeling\User`: false, `App\Kernel`: false,
	} {
		if w.Unseen(fqn) != unseen {
			t.Errorf("unseen %s: %v", fqn, !unseen)
		}
	}
	// A skipped file at the root of a PSR-4 directory hides the mapped
	// namespace: it may declare anything below it.
	kernel := Skipped{Files: []string{"src/Kernel.php"}}
	w = composerWorld(reader(map[string]string{"composer.json": `{"autoload": {"psr-4": {"App\\": "src/"}}}`, "vendor/composer/installed.json": `{"packages": []}`}), symbols(nil), []string{"src"}, nil, kernel)
	if !w.Complete || len(w.Blind) != 0 || !w.Unseen(`App\Kernel`) || !w.Unseen(`\App\Other`) || !w.Unseen(`App\Http\Missing`) || w.Unseen(`Kernel`) {
		t.Fatalf("root file: %+v", w)
	}
	// The project root as a PSR-4 directory.
	root := map[string]string{"composer.json": `{"autoload": {"psr-4": {"App\\": ""}}}`, "vendor/composer/installed.json": `{"packages": []}`}
	if w := composerWorld(reader(root), symbols(root), []string{"."}, nil, Skipped{Dirs: []string{"src/Gen"}}); !w.Complete || !slices.Equal(w.Blind, []string{`app\src\gen\`}) {
		t.Fatalf("root directory: %+v", w)
	}
	for name, edit := range map[string]func(*Skipped, *[]string){
		"classmap holds a skipped file": func(s *Skipped, _ *[]string) { s.Files = append(s.Files, "src/legacy/Old.inc") },
		"classmap in a skipped dir":     func(s *Skipped, _ *[]string) { s.Dirs = append(s.Dirs, "vendor") },
		"files entry unread":            func(s *Skipped, _ *[]string) { s.Files = append(s.Files, "src/helpers.php") },
		"files entry not indexed": func(_ *Skipped, ix *[]string) {
			*ix = slices.DeleteFunc(*ix, func(f string) bool { return f == "src/helpers.php" })
		},
		"fallback holds a skipped dir": func(s *Skipped, _ *[]string) {
			files["composer.json"] = `{"autoload": {"psr-4": {"": "src/"}}}`
			s.Dirs = append(s.Dirs, "src/Gen")
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() { files["composer.json"] = worldFiles()["composer.json"] }()
			s := Skipped{Dirs: slices.Clone(skipped.Dirs), Files: slices.Clone(skipped.Files)}
			ix := slices.Clone(indexed)
			edit(&s, &ix)
			if w := composerWorld(reader(files), symbols(files), []string{"src", "lib"}, ix, s); w.Complete {
				t.Fatalf("complete world: %+v", w)
			}
		})
	}
}

func TestComposerWorldOnDisk(t *testing.T) {
	root := t.TempDir()
	for name, src := range worldFiles() {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var indexed []string
	for _, f := range worldIndexed {
		indexed = append(indexed, filepath.Join(root, filepath.FromSlash(f)))
	}
	scope := []string{filepath.Join(root, "src"), filepath.Join(root, "lib")}
	if w := ComposerWorld(root, scope, indexed, Skipped{Dirs: []string{filepath.Join(root, "vendor", "acme", "http", "tests")}}, index.New(nil)); !w.Complete {
		t.Fatalf("world: %+v", w)
	}
	if w := ComposerWorld(filepath.Join(root, "missing"), scope, indexed, Skipped{}, index.New(nil)); w.Complete || w.Vendor != filepath.Join(root, "missing", "vendor") {
		t.Fatalf("world without composer.json: %+v", w)
	}
}

// TestComposerWorldVendorDir: the world names the vendor directory
// composer.json configures, in the form the vendor walk gives its files,
// even when the world is incomplete.
func TestComposerWorldVendorDir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "composer.json"), `{"config": {"vendor-dir": "libraries/deps"}}`)
	writeFile(t, filepath.Join(root, "libraries", "deps", "acme", "A.php"), "<?php class A {}")
	writeFile(t, filepath.Join(root, "vendor", "B.php"), "<?php class B {}")
	indexed, skipped := IndexSourcesSkipping(root, nil, nil, nil)
	w := ComposerWorld(root, nil, indexed, skipped, index.New(nil))
	if w.Complete || w.Vendor != filepath.Join(root, "libraries", "deps") {
		t.Fatalf("world: %+v", w)
	}
	if len(indexed) != 1 || !w.Dependency(indexed[0]) || w.Dependency(filepath.Join(root, "vendor", "B.php")) {
		t.Fatalf("indexed %v under %s", indexed, w.Vendor)
	}
}

func TestSnapshotWorld(t *testing.T) {
	files := worldFiles()
	s := &Snapshot{
		Sources:   map[string][]byte{"composer.json": []byte(files["composer.json"])},
		Config:    &config.Config{Paths: []string{"src", "lib"}},
		OwnFiles:  []string{"src/A.php", "src/helpers.php"},
		unseen:    worldSkipped,
		Vendor:    []string{"vendor/acme/http/src/Client.php", "vendor/acme/test/src/Case.php"},
		installed: []byte(files["vendor/composer/installed.json"]),
	}
	if w := s.world(); !w.Complete {
		t.Fatalf("snapshot world: %+v", w)
	}
	s.Paths = []string{"src"}
	if w := s.world(); !w.Complete || !w.Unseen(`App\Extra\Thing`) {
		t.Fatalf("lib/ is autoloaded but not captured: %+v", w)
	}
	s.Paths, s.SkippedCount = nil, 1
	if w := s.world(); w.Complete {
		t.Fatal("skipped vendor files may declare anything")
	}
	s.SkippedCount, s.installed = 0, nil
	if w := s.world(); w.Complete || w.Vendor != "vendor" {
		t.Fatalf("world without installed.json: %+v", w)
	}
	s.Sources["composer.json"] = []byte(`{"config": {"vendor-dir": "deps"}}`)
	if w := s.world(); w.Vendor != "deps" || !w.Dependency("deps/acme/A.php") {
		t.Fatalf("snapshot vendor-dir: %+v", w)
	}
}

// TestComposerWorldVendorPathForms: whether the root is absolute or
// relative to the working directory, a vendor file reaches the index as
// a dependency in either form.
func TestComposerWorldVendorPathForms(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "composer.json"), `{}`)
	t.Chdir(root)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{cwd, "."} {
		w := ComposerWorld(r, nil, nil, Skipped{}, index.New(nil))
		for file, want := range map[string]bool{
			"vendor/acme/lib/Base.php":                              true,
			filepath.Join(cwd, "vendor", "acme", "lib", "Base.php"): true,
			"src/vendor/Base.php":                                   false,
			filepath.Join(cwd, "src", "Base.php"):                   false,
		} {
			if got := w.Dependency(file); got != want {
				t.Errorf("root %s: Dependency(%s) = %v, want %v", r, file, got, want)
			}
		}
	}
}

// TestComposerWorldIndexedAutoloader: ComposerWorld finds a package's
// autoloader through the index, whichever form, absolute or relative to
// the working directory, the indexed paths take.
func TestComposerWorldIndexedAutoloader(t *testing.T) {
	root := t.TempDir()
	for name, src := range map[string]string{
		"composer.json":                  `{"require-dev": {"acme/tool": "*"}}`,
		"vendor/composer/installed.json": `{"packages": [{"name": "acme/tool", "install-path": "../acme/tool", "autoload": {"files": ["bootstrap.php"]}}]}`,
		"vendor/acme/tool/bootstrap.php": "<?php\nspl_autoload_register(function ($class) { if (strpos($class, 'ToolPrefix') === 0) { require 'x.php'; } });\n",
		"src/A.php":                      "<?php\nfinal class A {}\n",
	} {
		writeFile(t, filepath.Join(root, filepath.FromSlash(name)), src)
	}
	files := []string{filepath.Join(root, "src", "A.php"), filepath.Join(root, "vendor", "acme", "tool", "bootstrap.php")}
	t.Chdir(root)
	for _, form := range [][]string{files, {"src/A.php", "vendor/acme/tool/bootstrap.php"}} {
		ix, unread := BuildIndexUnread(context.Background(), form, syntax.Options{}, infer.Options{}, nil)
		if len(unread) != 0 {
			t.Fatal(unread)
		}
		w := ComposerWorld(root, []string{filepath.Join(root, "src")}, form, Skipped{}, ix)
		if !w.Complete || !w.UnseenClass(`ToolPrefix123\Thing`) || w.UnseenClass(`Other\Thing`) {
			t.Errorf("%q: %+v", form, w)
		}
	}
}
