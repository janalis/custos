package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkBuiltinDispatchMismatch(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		b.ReportAllocs()
		for b.Loop() {
			NativeBuiltin(ctx, call, "strlen")
			GlobalCall(ctx, call, "strlen")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("dispatch.php", []byte(`<?php namespace App; trim('value');`), syntax.Options{}))
}

func TestGlobalDispatchRejectsMatchingProjectFunction(t *testing.T) {
	probeNative(t, `function probe($value) {} probe('value');`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		if found, name := GlobalCall(ctx, call, "probe"); found != nil || name != "" {
			t.Fatal("project function was accepted as a builtin")
		}
		if NativeBuiltin(ctx, call, "probe") {
			t.Fatal("project function was accepted by native dispatch")
		}
	})
}
