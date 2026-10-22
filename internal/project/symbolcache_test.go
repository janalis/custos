package project

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// countExtractions counts the files parsed for their declarations.
func countExtractions(t *testing.T) func() int {
	t.Helper()
	old := extract
	t.Cleanup(func() { extract = old })
	var mu sync.Mutex
	n := 0
	extract = func(path string, src []byte, opt syntax.Options, options infer.Options) *index.FileSymbols {
		mu.Lock()
		n++
		mu.Unlock()
		return old(path, src, opt, options)
	}
	return func() int {
		mu.Lock()
		defer mu.Unlock()
		defer func() { n = 0 }()
		return n
	}
}

// backdate dates files back past the racy window so the cache stores them.
func backdate(t *testing.T, files ...string) {
	t.Helper()
	past := time.Now().Add(-time.Hour)
	for _, f := range files {
		if err := os.Chtimes(f, past, past); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSymbolCacheReusesUnchangedDeclarations(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	app := filepath.Join(root, "src", "app.php")
	lib := filepath.Join(root, "vendor", "lib", "Lib.php")
	dep := filepath.Join(root, "vendor", "lib", "dep.php")
	writeFile(t, app, "<?php\nfunction app_fn() {}\n")
	writeFile(t, lib, "<?php\nnamespace Lib;\nclass Thing {}\nconst LIMIT = 1;\n")
	// Its inferred return depends on whether the project declares Lib\helper.
	writeFile(t, dep, "<?php\nnamespace Lib;\nfunction wrap() { return helper(); }\n")
	backdate(t, lib, dep)
	opt := syntax.Options{Version: phpversion.Default}
	files := []string{app, lib, dep}
	extracted := countExtractions(t)
	build := func() *index.Index {
		ix, _ := BuildIndexCached(files, 1, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{}))
		return ix
	}
	if ix := build(); extracted() != 3 || ix.Class(`Lib\Thing`, phpversion.Default) == nil {
		t.Fatal("cold run")
	}
	if ix := build(); extracted() != 1 || ix.Class(`Lib\Thing`, phpversion.Default) == nil || ix.Function(`Lib\wrap`, phpversion.Default) == nil {
		t.Fatal("warm run must parse only the analysed file")
	}
	// A changed file is parsed again; another PHP version starts afresh.
	writeFile(t, lib, "<?php\nnamespace Lib;\nclass Other {}\n")
	backdate(t, lib)
	if ix := build(); extracted() != 2 || ix.Class(`Lib\Other`, phpversion.Default) == nil {
		t.Fatal("changed file")
	}
	ix, _ := BuildIndexCached(files, 1, syntax.Options{Version: phpversion.PHP74}, infer.Options{}, OpenSymbolCache(dir, root, syntax.Options{Version: phpversion.PHP74}, infer.Options{}))
	if extracted() != 3 || ix.Class(`Lib\Other`, phpversion.PHP74) == nil {
		t.Fatal("other options")
	}
}

func TestSymbolCacheSkipsRecentFiles(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	lib := filepath.Join(root, "Lib.php")
	writeFile(t, lib, "<?php\nclass Lib {}\n")
	opt := syntax.Options{Version: phpversion.Default}
	extracted := countExtractions(t)
	for range 2 {
		BuildIndexCached([]string{lib}, 0, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{}))
	}
	if n := extracted(); n != 2 {
		t.Fatalf("a file written just now must not be trusted: %d", n)
	}
	// A vanished file is skipped without an entry.
	if ix, _ := BuildIndexCached([]string{filepath.Join(root, "gone.php")}, 0, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{})); ix == nil {
		t.Fatal("no index")
	}
}

func TestSymbolCacheRelocatesDeclarations(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	lib := filepath.Join(root, "Lib.php")
	writeFile(t, lib, "<?php\nclass Lib {}\nfunction lib() {}\nconst LIB = 1;\n")
	backdate(t, lib)
	opt := syntax.Options{Version: phpversion.Default}
	BuildIndexCached([]string{lib}, 0, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{}))
	t.Chdir(root)
	ix, _ := BuildIndexCached([]string{"Lib.php"}, 0, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{}))
	if c := ix.Class("Lib", phpversion.Default); c == nil || c.File != "Lib.php" || ix.Function("lib", phpversion.Default).File != "Lib.php" {
		t.Fatal("declarations keep the earlier spelling of their file")
	}
	fs := &index.FileSymbols{Path: "a.php", Classes: []*index.Class{{File: "other.php"}}, Functions: []*index.Function{{File: "other.php"}}, Constants: []*index.Constant{{File: "other.php"}, {File: "a.php"}}}
	relocate(fs, "b.php")
	if fs.Path != "b.php" || fs.Classes[0].File != "other.php" || fs.Functions[0].File != "other.php" || fs.Constants[0].File != "other.php" || fs.Constants[1].File != "b.php" {
		t.Fatalf("%+v", fs)
	}
}

func TestSymbolCacheIgnoresUnusableFiles(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	opt := syntax.Options{Version: phpversion.Default}
	c := OpenSymbolCache(dir, root, opt, infer.Options{})
	if err := os.WriteFile(c.path, []byte("not gob"), 0o600); err != nil {
		t.Fatal(err)
	}
	if c := OpenSymbolCache(dir, root, opt, infer.Options{}); c == nil || c.old != nil {
		t.Fatal("a corrupt cache starts empty")
	}
	old := executable
	t.Cleanup(func() { executable = old })
	executable = func() (string, error) { return "", errors.New("unknown") }
	if OpenSymbolCache(dir, root, opt, infer.Options{}) != nil {
		t.Fatal("unidentified binary")
	}
	executable = func() (string, error) { return filepath.Join(root, "gone"), nil }
	if OpenSymbolCache(dir, root, opt, infer.Options{}) != nil {
		t.Fatal("vanished binary")
	}
	var none *SymbolCache
	if none.lookup("a.php", nil) != nil || none.Save() != nil {
		t.Fatal("nil cache")
	}
	none.store("a.php", nil, nil)
}

func TestSymbolCacheSaveFailures(t *testing.T) {
	root := t.TempDir()
	lib := filepath.Join(root, "Lib.php")
	writeFile(t, lib, "<?php\nclass Lib {}\n")
	backdate(t, lib)
	info, err := os.Stat(lib)
	if err != nil {
		t.Fatal(err)
	}
	opt := syntax.Options{Version: phpversion.Default}
	open := func(dir string) *SymbolCache {
		c := OpenSymbolCache(dir, root, opt, infer.Options{})
		c.store(lib, info, &index.FileSymbols{Path: lib})
		return c
	}
	// The directory cannot be created under a file.
	if err := open(filepath.Join(lib, "cache")).Save(); err == nil {
		t.Fatal("mkdir under a file")
	}
	readOnly := filepath.Join(t.TempDir(), "ro")
	if err := os.Mkdir(readOnly, 0o500); err != nil {
		t.Fatal(err)
	}
	if err := open(readOnly).Save(); err == nil {
		t.Fatal("write in a read-only directory")
	}
	// The cache path is taken by a non-empty directory: rename fails and
	// the written copy is removed.
	dir := t.TempDir()
	c := open(dir)
	writeFile(t, filepath.Join(c.path, "keep"), "")
	if err := c.Save(); err == nil {
		t.Fatal("rename over a directory")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("leftovers: %v", entries)
	}
	// A written cache reloads; left unused, it is rewritten without entries.
	c = open(t.TempDir())
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	again := OpenSymbolCache(filepath.Dir(c.path), root, opt, infer.Options{})
	if !reflect.DeepEqual(again.old[lib].Size, info.Size()) || again.Save() != nil {
		t.Fatal("reload")
	}
	if _, err := os.Stat(again.path); err != nil {
		t.Fatal(err)
	}
}

func TestProjectUsesSymbolCache(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "a.php"), "<?php\n$x = new Lib\\Thing();\n")
	lib := filepath.Join(root, "vendor", "lib", "Lib.php")
	writeFile(t, lib, "<?php\nnamespace Lib;\nclass Thing {}\n")
	backdate(t, lib)
	opt := syntax.Options{Version: phpversion.Default}
	p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: []string{filepath.Join(root, "src")}, Cache: t.TempDir(), Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: opt})
	if err != nil {
		t.Fatal(err)
	}
	extracted := countExtractions(t)
	p.AnalyzeReport()
	cold := extracted()
	if p.AnalyzeReport(); cold != 2 || extracted() != 1 {
		t.Fatalf("cold %d, warm must parse only the analysed file", cold)
	}
}

// TestSymbolCacheReplacedByNewBuild: another binary reads the same file,
// sets its entries aside, and overwrites it.
func TestSymbolCacheReplacedByNewBuild(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	lib := filepath.Join(root, "Lib.php")
	writeFile(t, lib, "<?php\nclass Lib {}\n")
	backdate(t, lib)
	opt := syntax.Options{Version: phpversion.Default}
	BuildIndexCached([]string{lib}, 0, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{}))
	old := executable
	t.Cleanup(func() { executable = old })
	executable = func() (string, error) { return lib, nil }
	c := OpenSymbolCache(dir, root, opt, infer.Options{})
	if c.old != nil {
		t.Fatal("entries of another build")
	}
	BuildIndexCached([]string{lib}, 0, opt, infer.Options{}, c)
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("cache files: %v", entries)
	}
}

func TestCacheDir(t *testing.T) {
	old := userCacheDir
	t.Cleanup(func() { userCacheDir = old })
	userCacheDir = func() (string, error) { return "/home/u/.cache", nil }
	t.Setenv("CUSTOS_CACHE_DIR", "")
	if got := CacheDir(); got != filepath.Join("/home/u/.cache", "custos") {
		t.Fatal(got)
	}
	t.Setenv("CUSTOS_CACHE_DIR", "/ci/cache")
	if got := CacheDir(); got != "/ci/cache" {
		t.Fatal(got)
	}
	t.Setenv("CUSTOS_CACHE_DIR", "")
	userCacheDir = func() (string, error) { return "", errors.New("no home") }
	if got := CacheDir(); got != "" {
		t.Fatal(got)
	}
}

func TestSymbolCacheFlowOrder(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	opt := syntax.Options{Version: phpversion.Default}
	open := func() *SymbolCache { return OpenSymbolCache(dir, root, opt, infer.Options{}) }
	var none *SymbolCache
	if none.FlowOrder(nil, flow.Options{}) != nil || none.SaveFlowOrder(nil, flow.Options{}) != nil {
		t.Fatal("nil cache")
	}
	a := filepath.Join(root, "a.php")
	if open().FlowOrder([]string{a}, flow.Options{}) != nil {
		t.Fatal("no order saved yet")
	}
	if err := open().SaveFlowOrder(map[string]int{a: 2}, flow.Options{}); err != nil {
		t.Fatal(err)
	}
	// Read back under another spelling of the path; files it lacks have none.
	t.Chdir(root)
	c := open()
	if got := c.FlowOrder([]string{"a.php", "b.php"}, flow.Options{}); !reflect.DeepEqual(got, map[string]int{"a.php": 2}) {
		t.Fatalf("order: %v", got)
	}
	// The same order is not written again.
	backdate(t, c.orderPath())
	info, _ := os.Stat(c.orderPath())
	if err := c.SaveFlowOrder(map[string]int{"a.php": 2}, flow.Options{}); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.Stat(c.orderPath()); !again.ModTime().Equal(info.ModTime()) {
		t.Fatal("unchanged order rewritten")
	}
	// Other flow options, or an unreadable file, have no order.
	if open().FlowOrder([]string{a}, flow.Options{Assertions: true}) != nil {
		t.Fatal("order of other options")
	}
	writeFile(t, c.orderPath(), "not gob")
	if open().FlowOrder([]string{a}, flow.Options{}) != nil {
		t.Fatal("corrupt order")
	}
}

// With FlowOrder, a run saves the order its flow build settled in, and the
// next one starts from it.
func TestProjectReusesFlowOrder(t *testing.T) {
	root, dir := t.TempDir(), t.TempDir()
	files := []string{filepath.Join(root, "a.php"), filepath.Join(root, "b.php")}
	writeFile(t, files[0], "<?php function a() { return $_GET['v']; }")
	writeFile(t, files[1], "<?php function b() { system(a()); }")
	opt := syntax.Options{Version: phpversion.Default}
	open := func(order bool) *Project {
		p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: []string{root}, Cache: dir, FlowOrder: order, Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: opt})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	cache := OpenSymbolCache(dir, root, opt, infer.Options{})
	open(false).AnalyzeReport()
	if _, err := os.Stat(cache.orderPath()); err == nil {
		t.Fatal("an order saved without FlowOrder")
	}
	for range 2 {
		found := 0
		for _, r := range open(true).AnalyzeReport() {
			found += len(r.Findings)
		}
		if found != 1 {
			t.Fatalf("%d findings", found)
		}
	}
	if got := cache.FlowOrder(files, flow.Options{}); got[files[1]] != 1 {
		t.Fatalf("order: %v", got)
	}
}
