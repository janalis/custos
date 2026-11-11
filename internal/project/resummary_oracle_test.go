package project

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// Every summary computed again in compose comes out as a full summary of
// the file would: the same contracts, read from the same keys. A project
// given by CUSTOS_ORACLE_CORPUS (its src directory is analyzed) is checked
// instead of the fixture.
func TestResummaryOracle(t *testing.T) {
	root, paths := oracleFixture(t)
	if corpus := os.Getenv("CUSTOS_ORACLE_CORPUS"); corpus != "" {
		root, paths = corpus, []string{filepath.Join(corpus, "src")}
	}
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	var mu sync.Mutex
	again, kept := 0, 0
	var differ []string
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options, prior *flow.File, earlier *flow.Snapshot) *flow.File {
		got := old(f, ix, v, s, o, prior, earlier)
		if prior == nil || earlier == nil || got == nil {
			return got
		}
		want := flow.Extract(f, ix, v, s, o)
		mu.Lock()
		defer mu.Unlock()
		again++
		if reflect.DeepEqual(got.Summaries, prior.Summaries) {
			kept++
		}
		switch {
		case !reflect.DeepEqual(got.Summaries, want.Summaries):
			differ = append(differ, f.Path+": summaries")
		case !slices.Equal(sortedKeys(got.Consulted()), sortedKeys(want.Consulted())):
			differ = append(differ, fmt.Sprintf("%s: consulted %v, want %v", f.Path, sortedKeys(got.Consulted()), sortedKeys(want.Consulted())))
		case !slices.Equal(got.Reads(), want.Reads()):
			differ = append(differ, fmt.Sprintf("%s: reads %v, want %v", f.Path, got.Reads(), want.Reads()))
		}
		return got
	}
	p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: paths, Parse: syntax.Options{Version: phpversion.PHP84}, Analysis: analysis.Config{PHP: phpversion.PHP84}})
	if err != nil {
		t.Fatal(err)
	}
	p.AnalyzeReport()
	t.Logf("summaries computed again: %d, unchanged: %d", again, kept)
	if again == 0 {
		t.Fatal("no summary was computed again")
	}
	for _, d := range differ {
		t.Error(d)
	}
}

// oracleFixture writes a recursion across two files, one of whose methods
// reads nothing of the other, so that compose summarizes it again.
func oracleFixture(t *testing.T) (string, []string) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "a.php"), `<?php
final class A {
	public function loop($v) { return (new B())->back($v); }
	public function plain() { return 'constant'; }
	public function same($v) { return $this->plain() . $v; }
}`)
	writeFile(t, filepath.Join(root, "src", "b.php"), `<?php
final class B {
	public function back($v) { return $v ? (new A())->loop($v) : $_GET['x']; }
}`)
	return root, []string{filepath.Join(root, "src")}
}

func sortedKeys(keys []string) []string {
	out := slices.Clone(keys)
	slices.Sort(out)
	return out
}
