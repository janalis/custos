package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// BenchmarkGeneratorCompletion includes fresh environments to exercise body inference.
func BenchmarkGeneratorCompletion(b *testing.B) {
	f := syntax.Parse("g.php", []byte(`<?php
function inner($b) { yield 'item' => 1; if ($b) return 'done'; return null; }
function outer() { return yield from inner(true); }
/** @return \Generator<int,string,float,bool> */
function documented() { $sent = yield 'item'; return true; }
outer()->getReturn(); documented()->send(1.5);
`), syntax.Options{Version: phpver.PHP85})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	ns := names.New(f)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, ns, ix, phpver.PHP85)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if expr, ok := n.(syntax.Expr); ok {
				env.TypeOf(expr)
			}
			return true
		})
	}
}
