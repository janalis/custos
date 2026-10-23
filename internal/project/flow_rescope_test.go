package project

import (
	"context"
	"maps"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// rescopeProject writes two clusters of files, each with a recursion: an
// edit within one summarizes the other again only when it must.
func rescopeProject(t *testing.T) (map[string]string, []string, *index.Index) {
	t.Helper()
	root := t.TempDir()
	paths := map[string]string{}
	for name, src := range map[string]string{
		"helper.php": "<?php function helper($x) { return $x; }",
		"ping.php":   "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }",
		"pong.php":   "<?php function pong($n) { ping($n); }",
		"user.php":   "<?php ping($_POST['b']);",
		"shell.php":  "<?php function shell($c) { system($c); }",
		"loop.php":   "<?php function loop($n) { if ($n) { loop($n - 1); } }",
		"caller.php": "<?php shell($_GET['x']); loop(1);",
	} {
		paths[name] = filepath.Join(root, name)
		writeFile(t, paths[name], src)
	}
	var files []string
	for _, path := range paths {
		files = append(files, path)
	}
	sort.Strings(files)
	return paths, files, BuildIndex(files, syntax.Options{Version: phpversion.Default}, infer.Options{})
}

// countSummaries counts, by path, the summaries extracted until the test ends.
func countSummaries(t *testing.T) map[string]int {
	t.Helper()
	var mu sync.Mutex
	counts := map[string]int{}
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options) *flow.File {
		mu.Lock()
		counts[f.Path]++
		mu.Unlock()
		return old(f, ix, v, s, o)
	}
	return counts
}

// updateLikeFull applies an edit through Update, then requires the build a
// build of all files from none makes, record included. counts, when given,
// keeps only the summaries the update extracted.
func updateLikeFull(t *testing.T, build *FlowBuild, path, src string, ix *index.Index, counts map[string]int) *FlowBuild {
	t.Helper()
	next, ok := build.Update(map[string][]byte{path: []byte(src)}, ix)
	if !ok {
		t.Fatal("update refused")
	}
	kept := maps.Clone(counts)
	writeFile(t, path, src)
	full := BuildFlow(context.Background(), build.files, nil, build.opt, ix, build.options)
	maps.Copy(counts, kept)
	for path := range counts {
		if _, ok := kept[path]; !ok {
			delete(counts, path)
		}
	}
	if !next.Snapshot.Equal(full.Snapshot) || next.traced && (!reflect.DeepEqual(next.seen, full.seen) || !reflect.DeepEqual(next.declares, full.declares)) {
		t.Fatalf("%s: update differs from a full build: snapshot %v traced %v seen %v declares %v", filepath.Base(path), next.Snapshot.Equal(full.Snapshot), next.traced, reflect.DeepEqual(next.seen, full.seen), reflect.DeepEqual(next.declares, full.declares))
	}
	return next
}

func TestFlowUpdateRescopesRecursion(t *testing.T) {
	paths, files, ix := rescopeProject(t)
	opt := syntax.Options{Version: phpversion.Default}
	build := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	counts := countSummaries(t)
	// The other cluster is left as it is.
	build = updateLikeFull(t, build, paths["ping.php"], "<?php function ping($n) { if ($n) { pong(helper($n)); } passthru($n); }", ix, counts)
	if !build.traced || counts[paths["shell.php"]] != 0 || counts[paths["helper.php"]] == 0 {
		t.Fatalf("scoped update: traced=%v summaries=%v", build.traced, counts)
	}
	// A call into the other cluster brings what it reads along.
	clear(counts)
	build = updateLikeFull(t, build, paths["ping.php"], "<?php function ping($n) { if ($n) { pong(helper($n)); } shell($n); }", ix, counts)
	if !build.traced || counts[paths["shell.php"]] == 0 || counts[paths["loop.php"]] != 0 {
		t.Fatalf("call into the other cluster: summaries=%v", counts)
	}
	// A contract consulted on the way that the edit's first summary did
	// not show: every file is summarized again.
	missed := build.results[build.at[paths["ping.php"]]]
	inner := extractFlow
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options) *flow.File {
		extractFlow = inner
		inner(f, ix, v, s, o) // counted
		return missed
	}
	clear(counts)
	build = updateLikeFull(t, build, paths["ping.php"], "<?php function ping($n) { if ($n) { pong(helper($n)); } loop($n); shell($n); }", ix, counts)
	if !build.traced || counts[paths["caller.php"]] == 0 {
		t.Fatalf("contract outside the scope: summaries=%v", counts)
	}
}

func TestFlowUpdateRescopeFallsBack(t *testing.T) {
	paths, files, ix := rescopeProject(t)
	opt := syntax.Options{Version: phpversion.Default}
	edit := "<?php function ping($n) { if ($n) { pong(helper($n)); } passthru($n); }"
	for _, tc := range []struct {
		name  string
		setup func(build *FlowBuild) *FlowBuild
	}{
		{"untraced", func(build *FlowBuild) *FlowBuild {
			// An update outside any recursion loses the record.
			next := updateLikeFull(t, build, paths["caller.php"], "<?php shell($_GET['y']); loop(2);", ix, nil)
			if next.traced {
				t.Fatal("record kept")
			}
			return next
		}},
		{"unreached file summarized differently", func(build *FlowBuild) *FlowBuild {
			build.results[build.at[paths["helper.php"]]] = &flow.File{Path: paths["helper.php"], Summaries: map[string]flow.Summary{"helper": {}}}
			return build
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for name, path := range paths {
				if name == "ping.php" {
					writeFile(t, path, "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }")
				}
			}
			build := tc.setup(BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{}))
			counts := countSummaries(t)
			next := updateLikeFull(t, build, paths["ping.php"], edit, ix, counts)
			if !next.traced || counts[paths["shell.php"]] == 0 {
				t.Fatalf("not rebuilt whole: %v", counts)
			}
		})
	}
}

// Another changed file whose summaries crash declares what no record says:
// every file is summarized again.
func TestFlowUpdateRescopeWithoutSummaries(t *testing.T) {
	paths, files, ix := rescopeProject(t)
	opt := syntax.Options{Version: phpversion.Default}
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options) *flow.File {
		if f.Path == paths["shell.php"] {
			return nil
		}
		return old(f, ix, v, s, o)
	}
	build := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	counts := countSummaries(t)
	edit := "<?php function ping($n) { if ($n) { pong(helper($n)); } passthru($n); }"
	next, ok := build.Update(map[string][]byte{paths["ping.php"]: []byte(edit), paths["shell.php"]: []byte("<?php function shell($c) { exec($c); }")}, ix)
	if !ok || counts[paths["loop.php"]] == 0 {
		t.Fatalf("not rebuilt whole: %v", counts)
	}
	writeFile(t, paths["ping.php"], edit)
	writeFile(t, paths["shell.php"], "<?php function shell($c) { exec($c); }")
	if !next.Snapshot.Equal(BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{}).Snapshot) {
		t.Fatal("update differs from a full build")
	}
}

func TestSameSummaries(t *testing.T) {
	file := &flow.File{Path: "a.php", Summaries: map[string]flow.Summary{"a": {}}}
	other := &flow.File{Path: "a.php", Summaries: map[string]flow.Summary{"b": {}}}
	if !sameSummaries(nil, nil) || sameSummaries(file, nil) || sameSummaries(nil, file) || !sameSummaries(file, file) || sameSummaries(file, other) {
		t.Fatal("summaries compared wrongly")
	}
}
