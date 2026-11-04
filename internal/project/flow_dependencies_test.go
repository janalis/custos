package project

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/infer"
)

func TestColdFlowChainsAreOrderIndependent(t *testing.T) {
	files := writeChain(t, 100)
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex(files, opt, infer.Options{})
	counts := countSummaries(t)
	build := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	if slices.Contains(build.results, nil) {
		t.Fatal("cold chain lost a summary")
	}
	for _, path := range files {
		if counts[path] > 2 {
			t.Fatalf("%s summarized %d times", path, counts[path])
		}
	}
	reversed := slices.Clone(files)
	slices.Reverse(reversed)
	reverse := BuildFlow(context.Background(), reversed, nil, opt, ix, flow.Options{})
	if !build.Snapshot.Equal(reverse.Snapshot) {
		t.Fatal("file order changed settled summaries")
	}
	// A changed callee changes the entire chain without requiring a warm order.
	source := files[0]
	replacement := []byte("<?php function f0(){return 'safe';}")
	ix.Add(ExtractSymbols(source, replacement, opt, infer.Options{}))
	next, ok := reverse.Update(map[string][]byte{source: replacement}, ix)
	if !ok {
		t.Fatal("incremental chain update refused")
	}
	srcs := make([][]byte, len(reversed))
	srcs[len(srcs)-1] = replacement
	full := BuildFlow(context.Background(), reversed, srcs, opt, ix, flow.Options{})
	if !next.Snapshot.Equal(full.Snapshot) || slices.Contains(next.results, nil) {
		t.Fatal("updated chain differs from a fresh build")
	}
}

func BenchmarkCrossFileDependencies(b *testing.B) {
	for _, n := range []int{16, 64} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			opt := syntax.Options{Version: phpversion.Default}
			// Sources are in memory so the benchmark measures composition, not IO.
			files := make([]string, n)
			srcs := make([][]byte, n)
			for i := range n {
				files[i] = fmt.Sprintf("f%d.php", i)
				text := fmt.Sprintf("<?php function f%d(){return f%d();}", i, i-1)
				if i == 0 {
					text = "<?php function f0(){return $_GET['v'];}"
				}
				srcs[i] = []byte(text)
			}
			ix := BuildIndex(nil, opt, infer.Options{})
			for i, path := range files {
				ix.Add(ExtractSymbols(path, srcs[i], opt, infer.Options{}))
			}
			b.ReportAllocs()
			for b.Loop() {
				BuildFlow(context.Background(), files, srcs, opt, ix, flow.Options{})
			}
		})
	}
}

func TestFlowOrderHintsKeepRecursiveResults(t *testing.T) {
	_, files, ix := rescopeProject(t)
	opt := syntax.Options{Version: phpversion.Default}
	plain := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	for variant := range 4 {
		order := map[string]int{}
		for i, path := range files {
			order[path] = (i*3 + variant) % len(files)
		}
		ordered := BuildFlowOrdered(context.Background(), files, nil, opt, ix, flow.Options{}, order)
		if !ordered.Snapshot.Equal(plain.Snapshot) {
			t.Fatalf("hint %d changed recursive summaries", variant)
		}
	}
}
