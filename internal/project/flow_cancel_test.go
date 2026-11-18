package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// cancelOnSummary cancels the returned context once a summary of path is
// extracted, after extracting it.
func cancelOnSummary(t *testing.T, path string) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options, prior *flow.File, earlier *flow.Snapshot, types *infer.Env) *flow.File {
		r := old(f, ix, v, s, o, prior, earlier, types)
		if f.Path == path {
			cancel()
		}
		return r
	}
	return ctx
}

// An update whose context ends before it completes is refused, whatever
// stage it reached: what it summarized is incomplete.
func TestFlowBuildUpdateCanceled(t *testing.T) {
	paths, files, ix := rescopeProject(t)
	build := BuildFlow(context.Background(), files, nil, syntax.Options{Version: phpversion.Default}, ix, flow.Options{})
	update := func(ctx context.Context, name, src string) bool {
		t.Helper()
		next, ok := build.Update(ctx, map[string][]byte{paths[name]: []byte(src)}, ix)
		if ok != (next != nil) {
			t.Fatalf("%s: build %v, ok %v", name, next, ok)
		}
		return ok
	}
	done, cancel := context.WithCancel(context.Background())
	cancel()
	if update(done, "shell.php", "<?php function shell($c) { passthru($c); }") {
		t.Fatal("update before any summary")
	}
	// Outside a recursion: once the edited file is summarized again.
	if update(cancelOnSummary(t, paths["shell.php"]), "shell.php", "<?php function shell($c) { passthru($c); }") {
		t.Fatal("update after the edited file's summary")
	}
	// Within a recursion (ping and pong), scoped or not.
	if update(cancelOnSummary(t, paths["ping.php"]), "pong.php", "<?php function pong($n) { ping($n); exec($n); }") {
		t.Fatal("update of a recursion")
	}
	if update(cancelOnSummary(t, paths["helper.php"]), "helper.php", "<?php function helper($x) { return ping($x); }") {
		t.Fatal("update of a whole-project recursion")
	}
	if !update(context.Background(), "shell.php", "<?php function shell($c) { passthru($c); }") {
		t.Fatal("update not canceled")
	}
}

// An index build whose context ends leaves the files not started out.
func TestBuildIndexUnreadCanceled(t *testing.T) {
	_, files, _ := rescopeProject(t)
	done, cancel := context.WithCancel(context.Background())
	cancel()
	opt := syntax.Options{Version: phpversion.Default}
	dir := t.TempDir()
	cache := OpenSymbolCache(dir, filepath.Dir(files[0]), opt, infer.Options{})
	ix, _ := BuildIndexUnread(done, files, opt, infer.Options{}, cache)
	if n, _, _, _ := ix.Stats(); n != 0 {
		t.Fatalf("%d files indexed after cancellation", n)
	}
	// What it reached is not saved over the cache a whole build left.
	if _, err := os.Stat(cache.path); !os.IsNotExist(err) {
		t.Fatalf("a canceled build saved its cache: %v", err)
	}
}
