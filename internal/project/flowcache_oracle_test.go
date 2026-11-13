package project

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// flowCacheRun is what one run of the oracle saw: its flow build and its
// findings, rendered.
type flowCacheRun struct {
	build    *FlowBuild
	findings string
}

// runFlowCache analyses the src directory of root as the command line does,
// with the cache directory dir ("" for none).
func runFlowCache(t *testing.T, root, dir string, opt syntax.Options) flowCacheRun {
	t.Helper()
	var build *FlowBuild
	old := preparedFlow
	defer func() { preparedFlow = old }()
	preparedFlow = func(b *FlowBuild) { build = b }
	src := filepath.Join(root, "src")
	p, err := Open(catalogue.All(), Options{Root: root, Paths: []string{src}, Scope: []string{src}, Cache: dir, Analysis: analysis.Config{PHP: opt.Version, EnableAll: true}, Parse: opt})
	if err != nil {
		t.Fatal(err)
	}
	results := p.AnalyzeReport()
	if build == nil {
		t.Fatal("no flow build")
	}
	return flowCacheRun{build, renderResults(results)}
}

// renderResults renders what a report shows of results.
func renderResults(results []diagnostic.FileResult) string {
	var b strings.Builder
	for _, r := range results {
		fmt.Fprintf(&b, "%s %v %v\n", r.Path, r.Incomplete, r.Err)
		for _, f := range r.Findings {
			var titles []string
			for _, fix := range f.Fixes {
				titles = append(titles, fix.Title)
			}
			fmt.Fprintf(&b, "  %s %v %v %q %q\n", f.Rule, f.Severity, f.Span, f.Message, titles)
		}
	}
	return b.String()
}

// sameFlowRuns fails unless the cached run is the one without the cache:
// the same summaries of each file, which make the snapshot (Snapshot.Equal
// would also compare the vendor providers, one per run), and the same
// findings.
func sameFlowRuns(t *testing.T, label string, cached, fresh flowCacheRun) {
	t.Helper()
	if !slices.Equal(cached.build.files, fresh.build.files) {
		t.Fatalf("%s: the files differ from a run without the cache", label)
	}
	for i, r := range cached.build.results {
		f := fresh.build.results[i]
		if (r == nil) != (f == nil) || r != nil && !reflect.DeepEqual(r.Summaries, f.Summaries) {
			t.Fatalf("%s: the summaries of %s differ from a run without the cache", label, cached.build.files[i])
		}
	}
	if cached.findings != fresh.findings {
		t.Fatalf("%s: the findings differ from a run without the cache", label)
	}
}

// A run with the flow steps cache comes out as a run without it: after
// edits outside and within a recursion, a revert, a change of declarations
// and a change of a dependency file. A project given by
// CUSTOS_ORACLE_CORPUS (its src directory, analysed with every rule) is
// checked instead of the fixture, editing on disk the files
// CUSTOS_ORACLE_EDITS lists (comma-separated, relative to the corpus) or
// else its three largest files: a trailing comment, a sink at the start of
// the first function body, then the original source, which the test
// restores in any case.
func TestFlowCacheOracle(t *testing.T) {
	opt := syntax.Options{Version: phpversion.PHP84}
	type edit struct{ path, src string }
	var root string
	var edits []edit
	if corpus := os.Getenv("CUSTOS_ORACLE_CORPUS"); corpus != "" {
		root = corpus
		files, _ := Discover([]string{filepath.Join(root, "src")}, nil)
		for _, e := range corpusEdits(t, root, files, opt, false) {
			edits = append(edits, edit(e))
		}
	} else {
		root = t.TempDir()
		src := func(name string) string { return filepath.Join(root, "src", name) }
		for name, code := range map[string]string{
			"helper.php": "<?php function helper($x) { return $x; }",
			"ping.php":   "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }",
			"pong.php":   "<?php function pong($n) { ping($n); }",
			"user.php":   "<?php function main() { ping($_POST['b']); }",
			"shell.php":  "<?php function shell($c) { \\Lib\\run($c); }",
			"caller.php": "<?php shell($_GET['x']);",
			// Which method relay calls depends on a declaration alone, the
			// return type of make, not on its summary.
			"runner.php":  "<?php final class Runner { public function go($c) { system($c); } }",
			"other.php":   "<?php final class Other { public function go($c) { return $c; } }",
			"factory.php": "<?php function make(): Runner { return $GLOBALS['r']; }",
			"relay.php":   "<?php function relay($c) { make()->go($c); }",
			"entry.php":   "<?php relay($_GET['r']);",
		} {
			writeFile(t, src(name), code)
		}
		lib := filepath.Join(root, "vendor", "lib", "run.php")
		writeFile(t, lib, "<?php\nnamespace Lib;\nfunction run($c) { system($c); }\n")
		backdate(t, lib)
		edits = []edit{
			{src("ping.php"), "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }\n// edit"},
			{src("ping.php"), "<?php function ping($n) { if ($n) { pong(helper($n)); } passthru($n); }"},
			{src("shell.php"), "<?php function shell($c) { \\Lib\\run($c); exec($c); }"},
			{src("ping.php"), "<?php function ping($n) { if ($n) { pong(helper($n)); } exec($n); }"},
			// Declarations change: the steps are not taken.
			{src("factory.php"), "<?php function make(): Other { return $GLOBALS['r']; }"},
			// So does a dependency file the vendor provider summarizes.
			{lib, "<?php\nnamespace Lib;\nfunction run($c) { return $c; }\n"},
		}
	}
	originals := map[string][]byte{}
	for _, e := range edits {
		if _, ok := originals[e.path]; !ok {
			src, err := os.ReadFile(e.path)
			if err != nil {
				t.Fatal(err)
			}
			originals[e.path] = src
		}
	}
	t.Cleanup(func() {
		for path, src := range originals {
			_ = os.WriteFile(path, src, 0o644)
		}
	})
	dir := t.TempDir()
	cold := runFlowCache(t, root, dir, opt)
	sameFlowRuns(t, "cold", cold, runFlowCache(t, root, "", opt))
	again := runFlowCache(t, root, dir, opt)
	if made := again.build.made.Load(); made != 0 {
		t.Fatalf("a run with nothing changed summarized %d files", made)
	}
	sameFlowRuns(t, "unchanged", again, cold)
	made := make([][2]int64, len(edits))
	for k, e := range edits {
		if err := os.WriteFile(e.path, []byte(e.src), 0o644); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(e.path, string(filepath.Separator)+"src"+string(filepath.Separator)) {
			backdate(t, e.path)
		}
		cached := runFlowCache(t, root, dir, opt)
		fresh := runFlowCache(t, root, "", opt)
		made[k] = [2]int64{cached.build.made.Load(), fresh.build.made.Load()}
		t.Logf("%s: %d summaries with the cache, %d without", filepath.Base(e.path), made[k][0], made[k][1])
		sameFlowRuns(t, filepath.Base(e.path), cached, fresh)
	}
	if os.Getenv("CUSTOS_ORACLE_CORPUS") == "" {
		// The edits within function bodies took the steps of the files they
		// do not reach; a change of declarations or of a dependency took
		// none.
		for k := range 4 {
			if made[k][0] >= made[k][1] {
				t.Errorf("edit %d: the cache saved no summary (%v)", k, made[k])
			}
		}
		for k := 4; k < len(edits); k++ {
			if made[k][0] != made[k][1] {
				t.Errorf("edit %d: the cache was taken (%v)", k, made[k])
			}
		}
	}
}
