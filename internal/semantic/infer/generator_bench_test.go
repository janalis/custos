package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// BenchmarkGeneratorCompletion includes fresh environments to exercise body inference.
func BenchmarkGeneratorCompletion(b *testing.B) {
	f := syntax.Parse("g.php", []byte(`<?php
function inner($b) { yield 'item' => 1; if ($b) return 'done'; return null; }
function outer() { return yield from inner(true); }
/** @return \Generator<int,string,float,bool> */
function documented() { $sent = yield 'item'; return true; }
outer()->getReturn(); documented()->send(1.5);
`), syntax.Options{Version: phpversion.PHP85})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	ns := names.New(f)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, ns, ix, phpversion.PHP85)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if expr, ok := n.(syntax.Expr); ok {
				env.TypeOf(expr)
			}
			return true
		})
	}
}
