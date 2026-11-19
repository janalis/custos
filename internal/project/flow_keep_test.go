package project

import (
	"context"
	"sync"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/flow"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/infer"
)

// A file summarized again in a build given its files' trees (--parse-once)
// that records no steps infers with the environment of its last summary,
// within the budget; a build without trees, the language server's recorded
// builds and a spent budget build a new one each time, and every build
// composes the same snapshot.
func TestFlowKeepsEnvironmentForNextSummary(t *testing.T) {
	files := writeChain(t, 4)
	opt := syntax.Options{Version: phpversion.Default}
	ix, srcs, trees, _ := buildIndex(context.Background(), files, len(files), opt, infer.Options{}, nil, true)
	parsedOnce := func() *FlowBuild {
		return buildFlow(context.Background(), files, srcs, opt, ix, flow.Options{}, nil, nil, false, trees)
	}
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	var mu sync.Mutex
	var seen map[*infer.Env]int
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options, prior *flow.File, earlier *flow.Snapshot, types *infer.Env) *flow.File {
		mu.Lock()
		seen[types]++
		mu.Unlock()
		if types.File != f {
			t.Errorf("environment of %s summarizes %s", types.File.Path, f.Path)
		}
		return old(f, ix, v, s, o, prior, earlier, types)
	}
	reused := func(build func() *FlowBuild) (*FlowBuild, int) {
		seen = map[*infer.Env]int{}
		b := build()
		n := 0
		for _, count := range seen {
			n += count - 1
		}
		if b.kept != nil {
			t.Fatalf("kept after the build: %v", b.kept)
		}
		return b, n
	}
	plain, n := reused(func() *FlowBuild { return BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{}) })
	if n != 0 {
		t.Fatalf("build without trees reused %d environments", n)
	}
	once, n := reused(parsedOnce)
	if n == 0 || !once.Snapshot.Equal(plain.Snapshot) {
		t.Fatal("no summary of a parse-once build reused an environment")
	}
	recorded, n := reused(func() *FlowBuild {
		return BuildFlowRecorded(context.Background(), files, nil, opt, ix, flow.Options{}, nil)
	})
	if n != 0 || !recorded.Snapshot.Equal(plain.Snapshot) {
		t.Fatalf("recorded build reused %d environments", n)
	}
	budget := keepBudget
	t.Cleanup(func() { keepBudget = budget })
	keepBudget = 0
	spent, n := reused(parsedOnce)
	if n != 0 || !spent.Snapshot.Equal(plain.Snapshot) {
		t.Fatalf("spent budget reused %d environments", n)
	}
}
