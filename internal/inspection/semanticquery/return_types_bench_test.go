package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// BenchmarkReturnTypeMismatch judges one returned value against the
// declared return type of a project function: rejected and proven.
func BenchmarkReturnTypeMismatch(b *testing.B) {
	const src = `<?php
declare(strict_types=1);
namespace App;
final class Order {}
function total(Order $order, string $currency): int {
    $cents = 'none';
    probe();
    return $cents;
}`
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		ret := syntax.EnclosingFuncLike(call).(*syntax.Function).Body.Stmts[2]
		if got, ok := returnTypeMismatch(ctx, ret); !ok || !got.Proven {
			b.Fatalf("mismatch = %+v, %v", got, ok)
		}
		ran = true
		b.ReportAllocs()
		for b.Loop() {
			returnTypeMismatch(ctx, ret)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(src), syntax.Options{}))
	if !ran {
		b.Fatal("no return checked")
	}
}
