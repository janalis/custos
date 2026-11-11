package project

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"custos/internal/project/config"
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
	w := composerWorld(reader(worldFiles()), []string{"src", "lib"}, worldIndexed, worldSkipped)
	if !w.Complete || !slices.Equal(w.Blind, []string{`app\tests\`, `acme\http\testing\`}) {
		t.Fatalf("world: %+v", w)
	}
	if !w.Unseen(`\App\Tests\Unit\CaseTest`) || w.Unseen(`App\Model`) {
		t.Fatal("blind prefixes")
	}
	// Composer 1 writes a bare list, and the whole project in scope.
	files := worldFiles()
	files["vendor/composer/installed.json"] = `[
		{"name": "acme/http", "autoload": {"psr-4": {"Acme\\Http\\": "src/"}}, "provide": {"psr/log-implementation": "1.0"}},
		{"name": "acme/test"}
	]`
	if w := composerWorld(reader(files), []string{"."}, worldIndexed, worldSkipped); !w.Complete || len(w.Blind) != 0 {
		t.Fatalf("composer 1: %+v", w)
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
			if w := composerWorld(reader(files), []string{"src", "lib"}, worldIndexed, worldSkipped); w.Complete {
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
		// Big.php was too large to index; a declaration-only file names no
		// class PSR-4 can load.
		Files: []string{"vendor/acme/http/src/Big.php", "src/api.stub.php"},
	}
	w := composerWorld(reader(files), []string{"src", "lib"}, indexed, skipped)
	want := []string{`app\legacy\`, `app\tests\`, `acme\old\`, `acme\http\tests\`, `acme\http\testing\`}
	if !w.Complete || !slices.Equal(sortedCopy(w.Blind), sortedCopy(want)) || !slices.Equal(w.Names, []string{`acme\http\big`}) {
		t.Fatalf("world: %+v", w)
	}
	// A skipped file hides its own class, not the names it starts.
	if !w.Unseen(`Acme\Http\Big`) || w.Unseen(`Acme\Http\BigQuerry`) || w.Unseen(`Acme\Http\Big\Part`) {
		t.Fatalf("unseen: %+v", w)
	}
	// The project root as a PSR-4 directory.
	root := map[string]string{"composer.json": `{"autoload": {"psr-4": {"App\\": ""}}}`, "vendor/composer/installed.json": `{"packages": []}`}
	if w := composerWorld(reader(root), []string{"."}, nil, Skipped{Dirs: []string{"src/Gen"}}); !w.Complete || !slices.Equal(w.Blind, []string{`app\src\gen\`}) {
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
			if w := composerWorld(reader(files), []string{"src", "lib"}, ix, s); w.Complete {
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
	if w := ComposerWorld(root, scope, indexed, Skipped{Dirs: []string{filepath.Join(root, "vendor", "acme", "http", "tests")}}); !w.Complete {
		t.Fatalf("world: %+v", w)
	}
	if w := ComposerWorld(filepath.Join(root, "missing"), scope, indexed, Skipped{}); w.Complete {
		t.Fatal("world without composer.json")
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
	if w := s.world(); w.Complete {
		t.Fatal("world without installed.json")
	}
}
