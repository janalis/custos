package project

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/inspection/rules/securityadvisories"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// vendorProject writes a project whose inputs are old enough to be trusted
// by size and modification time.
func vendorProject(t *testing.T) (string, *os.Root) {
	t.Helper()
	dir := t.TempDir()
	for name, text := range map[string]string{
		"custos.json":            `{"php":"8.2","baseline":"baseline.json","rules":{"NestedNotOperators":{"enabled":true}}}`,
		"baseline.json":          `{"version":1,"entries":[]}`,
		"composer.json":          `{}`,
		"a.php":                  "<?php $ok=!!!$x;",
		"vendor/pkg/a.php":       "<?php function vendor_a(): int { return 1; }",
		"vendor/pkg/b.php":       "<?php namespace V; function dependent() { return strlen('x'); }",
		"vendor/pkg/tests/t.php": "<?php function vendor_test() {}",
		"vendor/pkg/notes.txt":   "not PHP",
	} {
		writeFile(t, filepath.Join(dir, name), text)
	}
	settle(t, dir)
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return dir, root
}

func settle(t *testing.T, dir string) {
	t.Helper()
	old := time.Now().Add(-time.Hour)
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Type()&os.ModeSymlink == 0 {
			err = os.Chtimes(p, old, old)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVendorIsIndexedWithoutRetention(t *testing.T) {
	dir, root := vendorProject(t)
	large, err := os.Create(filepath.Join(dir, "vendor/pkg/large.php"))
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate(syntax.MaxFileSize + 1); err != nil {
		t.Fatal(err)
	}
	_ = large.Close()
	if err := os.Symlink(t.TempDir(), filepath.Join(dir, "vendor/linked")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "vendor/pkg/crash.php"), "<?php function crashes() {}")
	settle(t, dir)
	original := extract
	t.Cleanup(func() { extract = original })
	extract = func(name string, src []byte, opt syntax.Options) *index.FileSymbols {
		if strings.HasSuffix(name, "crash.php") {
			panic("parser crash")
		}
		return original(name, src, opt)
	}
	ctx := context.Background()
	rules := []analysis.Rule{nestednotoperators.New()}
	s, err := BuildSnapshot(ctx, root, rules, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Vendor) != 3 || s.Sources["vendor/pkg/a.php"] != nil || s.Hashes[VendorDigest] == "" || s.VendorHashes["vendor/pkg/a.php"] == "" {
		t.Fatal(s.Vendor, s.Hashes)
	}
	if s.Index.Function("vendor_a", s.Config.PHP) == nil || s.Index.Function("vendor_test", s.Config.PHP) != nil || len(s.Results) != 1 {
		t.Fatal("vendor declarations", s.Results)
	}
	if s.SkippedCount != 3 || strings.Join(s.Skipped, "|") != "vendor/linked (symlink)|vendor/pkg/large.php (too large)|vendor/pkg/crash.php (unparsable)" {
		t.Fatal(s.Skipped)
	}
	// Reported skips stay bounded however many entries are left out.
	l := &listing{}
	for range maxSkipped + 1 {
		l.skip("x", "symlink")
	}
	if l.skippedCount != maxSkipped+1 || len(l.skipped) != maxSkipped {
		t.Fatal(l.skipped)
	}
	s.Skipped = l.skipped
	s.vendorSymbols = nil
	s.Symbols, s.Index = nil, index.New(stubs.Index())
	if _, err := AnalyzeSnapshot(ctx, s); err != nil || len(s.Skipped) != maxSkipped {
		t.Fatal(s.Skipped, err)
	}

	if b, err := s.Source("a.php"); err != nil || string(b) != "<?php $ok=!!!$x;" {
		t.Fatal(string(b), err)
	}
	if b, err := s.Source("vendor/pkg/a.php"); err != nil || !strings.Contains(string(b), "vendor_a") {
		t.Fatal(string(b), err)
	}
	if _, err := s.Source("vendor/pkg/tests/t.php"); err == nil {
		t.Fatal("unlisted source")
	}
	writeFile(t, filepath.Join(dir, "vendor/pkg/a.php"), "<?php function replaced() {}")
	if _, err := s.Source("vendor/pkg/a.php"); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "vendor/pkg/a.php")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Source("vendor/pkg/a.php"); err == nil {
		t.Fatal("removed source")
	}
}

func TestVendorBudgetAndFailures(t *testing.T) {
	dir, root := vendorProject(t)
	ctx := context.Background()
	rules := []analysis.Rule{nestednotoperators.New()}
	files := maxVendorFiles
	maxVendorFiles = 1
	_, err := CaptureSnapshot(ctx, root, rules, nil, nil)
	maxVendorFiles = files
	if err == nil || !strings.Contains(err.Error(), "vendor input budget exceeded: 2 files") {
		t.Fatal(err)
	}
	for stop := 1; stop < 12; stop++ {
		_, _ = BuildSnapshot(&stepContext{Context: ctx, stop: stop}, root, rules, nil, nil)
	}
	if err := os.Symlink(filepath.Join(dir, "a.php"), filepath.Join(dir, "link.php")); err != nil {
		t.Fatal(err)
	}
	// A symlink below the levels that are listed one directory at a time.
	if err := os.MkdirAll(filepath.Join(dir, "deep/a/b/c"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "a.php"), filepath.Join(dir, "deep/a/b/c/link.php")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"missing", "a.php/child", "link.php", "deep"} {
		if _, err := CaptureSnapshot(ctx, root, rules, []string{name}, nil); err == nil {
			t.Fatal(name)
		}
	}
	if err := os.Remove(filepath.Join(dir, "link.php")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "deep")); err != nil {
		t.Fatal(err)
	}
	// A configuration file that a rule also inspects is captured once.
	if s, err := CaptureSnapshot(ctx, root, append(rules, securityadvisories.New()), nil, nil); err != nil || len(s.Files) != 2 {
		t.Fatal(s, err)
	}
	// A manifest captured without declarations rejects a vendor file that
	// changed before analysis.
	captured, err := CaptureSnapshot(ctx, root, rules, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AnalyzeSnapshot(&stepContext{Context: ctx, stop: 1}, captured); err == nil {
		t.Fatal("cancelled vendor extraction")
	}
	writeFile(t, filepath.Join(dir, "vendor/pkg/a.php"), "<?php function vendor_changed() {}")
	if _, err := AnalyzeSnapshot(ctx, captured); err == nil || !strings.Contains(err.Error(), "changed during analysis") {
		t.Fatal(err)
	}
	// A directory removed between listing and reading fails the capture.
	counter := &stepContext{Context: ctx}
	if _, err := CaptureSnapshot(counter, root, rules, nil, nil); err != nil {
		t.Fatal(err)
	}
	vanish := &stepContext{Context: ctx, run: func(n int) {
		if n == counter.steps {
			_ = os.RemoveAll(filepath.Join(dir, "vendor/pkg"))
		}
	}}
	if _, err := CaptureSnapshot(vanish, root, rules, nil, nil); err == nil {
		t.Fatal("vanished vendor directory")
	}
	writeFile(t, filepath.Join(dir, "vendor/pkg/a.php"), "<?php")
	if err := os.Chmod(filepath.Join(dir, "vendor/pkg/a.php"), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := CaptureSnapshot(ctx, root, rules, nil, nil); err == nil {
		t.Fatal("unreadable vendor file")
	}
	if err := os.Chmod(filepath.Join(dir, "vendor/pkg"), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "vendor/pkg"), 0o700) })
	if _, err := CaptureSnapshot(ctx, root, rules, nil, nil); err == nil {
		t.Fatal("unreadable vendor directory")
	}
}

func TestVendorCacheReusesSettledDeclarations(t *testing.T) {
	dir, root := vendorProject(t)
	ctx := context.Background()
	rules := []analysis.Rule{nestednotoperators.New()}
	cache := &VendorCache{}
	first, err := BuildSnapshot(ctx, root, rules, nil, cache)
	if err != nil {
		t.Fatal(err)
	}
	// Declarations with return dependencies are rewritten per index and are
	// therefore extracted again rather than shared.
	if cache.files["vendor/pkg/a.php"].symbols == nil || cache.files["vendor/pkg/b.php"].symbols != nil {
		t.Fatal(cache.files)
	}
	second, err := BuildSnapshot(ctx, root, rules, nil, cache)
	if err != nil {
		t.Fatal(err)
	}
	if first.vendorSymbols["vendor/pkg/a.php"] != second.vendorSymbols["vendor/pkg/a.php"] || first.vendorSymbols["vendor/pkg/b.php"] == second.vendorSymbols["vendor/pkg/b.php"] {
		t.Fatal("cache reuse")
	}
	if first.Hashes[VendorDigest] != second.Hashes[VendorDigest] {
		t.Fatal("digest changed")
	}
	// A hash-only capture needs no read at all for settled entries, and a
	// later analysis extracts only what the cache cannot supply.
	captured, err := CaptureSnapshot(ctx, root, rules, nil, cache)
	if err != nil {
		t.Fatal(err)
	}
	analyzed, err := AnalyzeSnapshot(ctx, captured)
	if err != nil || analyzed.vendorSymbols["vendor/pkg/a.php"] != first.vendorSymbols["vendor/pkg/a.php"] {
		t.Fatal(err)
	}
	// Rewriting a file with identical bytes changes its stat, not its hash:
	// its declarations are kept.
	writeFile(t, filepath.Join(dir, "vendor/pkg/a.php"), "<?php function vendor_a(): int { return 1; }")
	third, err := BuildSnapshot(ctx, root, rules, nil, cache)
	if err != nil || third.vendorSymbols["vendor/pkg/a.php"] != first.vendorSymbols["vendor/pkg/a.php"] || cache.files["vendor/pkg/a.php"].settled {
		t.Fatal(err)
	}
	// Different parse options invalidate every cached declaration.
	writeFile(t, filepath.Join(dir, "custos.json"), `{"php":"7.4","baseline":"baseline.json"}`)
	fourth, err := BuildSnapshot(ctx, root, rules, nil, cache)
	if err != nil || fourth.vendorSymbols["vendor/pkg/a.php"] == first.vendorSymbols["vendor/pkg/a.php"] {
		t.Fatal(err)
	}
}

func TestChangedInputs(t *testing.T) {
	dir, root := vendorProject(t)
	ctx := context.Background()
	rules := []analysis.Rule{nestednotoperators.New()}
	build := func() *Snapshot {
		t.Helper()
		settle(t, dir)
		s, err := BuildSnapshot(ctx, root, rules, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	expect := func(s *Snapshot, want string) {
		t.Helper()
		changed, err := ChangedInputs(ctx, s)
		if err != nil || strings.Join(changed, ",") != want {
			t.Fatalf("changed %v, want %s: %v", changed, want, err)
		}
	}
	s := build()
	expect(s, "")
	// Same bytes with a new modification time are hashed, not trusted.
	writeFile(t, filepath.Join(dir, "a.php"), "<?php $ok=!!!$x;")
	expect(s, "")
	writeFile(t, filepath.Join(dir, "a.php"), "<?php $ok=!!!$y;")
	writeFile(t, filepath.Join(dir, "new.php"), "<?php")
	writeFile(t, filepath.Join(dir, "vendor/pkg/b.php"), "<?php function other() {}")
	if err := os.Remove(filepath.Join(dir, "vendor/pkg/a.php")); err != nil {
		t.Fatal(err)
	}
	expect(s, "a.php,new.php,vendor/pkg/a.php,vendor/pkg/b.php")

	s = build()
	writeFile(t, filepath.Join(dir, "baseline.json"), `{"version":1,"entries":[] }`)
	expect(s, "baseline.json")
	if err := os.Remove(filepath.Join(dir, "baseline.json")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "custos.json"), `{"php":"8.3"}`)
	writeFile(t, filepath.Join(dir, "composer.json"), `{"name":"a/b"}`)
	expect(s, "custos.json,composer.json,baseline.json")

	s = build()
	if _, err := ChangedInputs(&stepContext{Context: ctx, stop: 1}, s); err == nil {
		t.Fatal("cancelled walk")
	}
	for _, name := range []string{"a.php", "vendor/pkg/b.php"} {
		p := filepath.Join(dir, name)
		writeFile(t, p, "<?php // unreadable")
		if err := os.Chmod(p, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := ChangedInputs(ctx, s); err == nil {
			t.Fatal("unreadable input", name)
		}
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(dir, "composer.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(dir, "composer.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := ChangedInputs(ctx, s); err == nil {
		t.Fatal("linked configuration")
	}
}
