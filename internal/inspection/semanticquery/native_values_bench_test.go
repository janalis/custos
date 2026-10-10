package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkNativeArrayEntries(b *testing.B) {
	src := "<?php probe([" + strings.Repeat("'record',", 100) + "]);"
	file := syntax.Parse("array.php", []byte(src), syntax.Options{})
	var ctx *analysis.Context
	var input syntax.Expr
	probe := nativeProbe{check: func(c *analysis.Context, call *syntax.FuncCall) {
		ctx, input = c, CallArgument(call.Args, 0, "value")
	}}
	engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	engine.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		NativeArrayEntries(ctx, input)
	}
}
