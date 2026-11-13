package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// BenchmarkArgumentTypeMismatches checks the arguments of one call to a
// project function: four typed parameters, one argument rejected.
func BenchmarkArgumentTypeMismatches(b *testing.B) {
	const src = `<?php
declare(strict_types=1);
namespace App;
final class Order {}
function charge(int $cents, string $currency, Order $order, ?array $meta = null) {}
function pay(Order $order, string $currency) {
    $cents = 1200;
    charge($cents, $currency, $order, '');
}`
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		if name, ok := call.Name.(*syntax.Name); !ok || name.Value != "charge" {
			return
		}
		if got := argumentTypeMismatches(ctx, call); len(got) != 1 || !got[0].Proven {
			b.Fatalf("mismatches = %+v", got)
		}
		ran = true
		b.ReportAllocs()
		for b.Loop() {
			argumentTypeMismatches(ctx, call)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(src), syntax.Options{}))
	if !ran {
		b.Fatal("no charge() call checked")
	}
}
