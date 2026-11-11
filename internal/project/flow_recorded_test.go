package project

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// recordedEdit is one edit the recorded update oracle applies.
type recordedEdit struct {
	path, src string
}

// Every update of a recorded build comes out as a build of all files from
// none: the summaries an earlier build made are taken only where the
// contracts they read still hold. A project given by CUSTOS_ORACLE_CORPUS
// (its src directory) is checked instead of the fixture, editing the files
// CUSTOS_ORACLE_EDITS lists (comma-separated, relative to the corpus) or
// else its three largest files: a trailing comment, a sink at the start of
// the first function body, then the original source.
func TestRecordedUpdateOracle(t *testing.T) {
	var (
		files   []string
		ix      *index.Index
		opt     = syntax.Options{Version: phpversion.PHP84}
		options flow.Options
		edits   []recordedEdit
	)
	if root := os.Getenv("CUSTOS_ORACLE_CORPUS"); root != "" {
		files, _ = Discover([]string{filepath.Join(root, "src")}, nil)
		ix = BuildIndex(IndexSources(root, files, nil, nil), opt, infer.Options{})
		options = WithVendorFlow(root, ix, opt, flow.Options{})
		edits = corpusEdits(t, root, files, opt)
	} else {
		var paths map[string]string
		paths, files, ix = rescopeProject(t)
		opt = syntax.Options{Version: phpversion.Default}
		ping := paths["ping.php"]
		edits = []recordedEdit{
			{ping, "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }\n// edit"},
			{ping, "<?php function ping($n) { if ($n) { pong(helper($n)); } passthru($n); }"},
			{paths["pong.php"], "<?php function pong($n) { ping($n); shell($n); }"},
			// Outside any recursion, then within it again.
			{paths["shell.php"], "<?php function shell($c) { exec($c); }"},
			{paths["user.php"], "<?php function main() { ping($_GET['b']); }"},
			{ping, "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }"},
		}
	}
	var calls atomic.Int64
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options, prior *flow.File, earlier *flow.Snapshot) *flow.File {
		calls.Add(1)
		return old(f, ix, v, s, o, prior, earlier)
	}
	build := BuildFlowRecorded(context.Background(), files, nil, opt, ix, options, nil)
	sources := map[string][]byte{}
	for _, e := range edits {
		sources[e.path] = []byte(e.src)
		calls.Store(0)
		start := time.Now()
		next, ok := build.Update(map[string][]byte{e.path: []byte(e.src)}, ix)
		took, made := time.Since(start), calls.Load()
		if !ok {
			t.Fatalf("%s: update refused", e.path)
		}
		srcs := make([][]byte, len(files))
		for i, path := range files {
			srcs[i] = sources[path]
		}
		calls.Store(0)
		start = time.Now()
		full := BuildFlowOrdered(context.Background(), files, srcs, opt, ix, options, nil)
		t.Logf("%s: update %v (%d summaries), full build %v (%d)", filepath.Base(e.path), took, made, time.Since(start), calls.Load())
		if !next.Snapshot.Equal(full.Snapshot) {
			t.Fatalf("%s: update differs from a full build", e.path)
		}
		for i, r := range next.results {
			if (r == nil) != (full.results[i] == nil) || r != nil && !reflect.DeepEqual(r.Summaries, full.results[i].Summaries) {
				t.Fatalf("%s: summaries of %s differ from a full build", e.path, files[i])
			}
		}
		build = next
	}
}

// firstBody finds the opening brace of the first function body.
var firstBody = regexp.MustCompile(`function\s+\w+\s*\([^)]*\)\s*(:\s*[\w\\?|]+\s*)?\{`)

// corpusEdits returns the edits TestRecordedUpdateOracle applies to the
// corpus at root, keeping those that leave the declarations as they are
// (Update's contract).
func corpusEdits(t *testing.T, root string, files []string, opt syntax.Options) []recordedEdit {
	t.Helper()
	var targets []string
	if list := os.Getenv("CUSTOS_ORACLE_EDITS"); list != "" {
		for rel := range strings.SplitSeq(list, ",") {
			targets = append(targets, filepath.Join(root, rel))
		}
	} else {
		sized := slices.Clone(files)
		size := func(path string) int64 {
			info, err := os.Stat(path)
			if err != nil {
				return 0
			}
			return info.Size()
		}
		slices.SortFunc(sized, func(a, b string) int { return int(size(b) - size(a)) })
		targets = sized[:min(3, len(sized))]
	}
	var edits []recordedEdit
	for _, path := range targets {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		same := func(edited string) bool {
			was := ExtractSymbols(path, src, opt, infer.Options{})
			return index.SameDeclarations(was, ExtractSymbols(path, []byte(edited), opt, infer.Options{}))
		}
		candidates := []string{string(src) + "\n// edit\n"}
		if at := firstBody.FindIndex(src); at != nil {
			candidates = append(candidates, string(src[:at[1]])+" exec($_GET['custos']); "+string(src[at[1]:]))
		}
		for _, edited := range candidates {
			if same(edited) {
				edits = append(edits, recordedEdit{path, edited})
			}
		}
		edits = append(edits, recordedEdit{path, string(src)})
	}
	return edits
}

// A recorded build keeps no step of a file it could not read or whose
// summaries crashed: a later update summarizes it again, as a build of all
// files from none does.
func TestRecordedUpdateWithoutSummaries(t *testing.T) {
	paths, files, ix := rescopeProject(t)
	files = append(files, filepath.Join(filepath.Dir(paths["ping.php"]), "missing.php"))
	opt := syntax.Options{Version: phpversion.Default}
	build := BuildFlowRecorded(context.Background(), files, nil, opt, ix, flow.Options{}, nil)
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options, prior *flow.File, earlier *flow.Snapshot) *flow.File {
		if f.Path == paths["shell.php"] {
			return nil
		}
		return old(f, ix, v, s, o, prior, earlier)
	}
	build = updateLikeFull(t, build, paths["shell.php"], "<?php function shell($c) { exec($c); }", ix, nil)
	if build.steps[build.at[paths["shell.php"]]] != nil {
		t.Fatal("a crashed summary was recorded")
	}
	build = updateLikeFull(t, build, paths["ping.php"], "<?php function ping($n) { if ($n) { pong(helper($n)); } shell($n); }", ix, nil)
	if build.results[build.at[paths["shell.php"]]] != nil || build.steps[len(files)-1] != nil {
		t.Fatal("steps kept for a file without summaries")
	}
}
