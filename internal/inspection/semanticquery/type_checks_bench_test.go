package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

const typeChecksBenchSrc = `<?php
namespace App;
interface Priced {}
final class Order implements Priced {}
function find(): Order { return new Order(); }
function check(int $qty, string $code, ?Order $order): void {
    if ($order === null) {
        return;
    }
    $mode = 'cart';
    if ($qty > 1) {
        $mode = 'bulk';
    }
    probe($order instanceof Priced, $qty === $code, $mode === 'sale', find() === null);
}`

// benchTypeChecks runs check on the probe() call of typeChecksBenchSrc.
func benchTypeChecks(b *testing.B, check func(ctx *analysis.Context, args []syntax.Expr)) {
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		if name, ok := call.Name.(*syntax.Name); ok && name.Value == "probe" {
			var args []syntax.Expr
			for _, a := range call.Args.Args {
				args = append(args, a.(*syntax.Arg).Value)
			}
			check(ctx, args)
			ran = true
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}, PHP: 804})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(typeChecksBenchSrc), syntax.Options{}))
	if !ran {
		b.Fatal("no probe() call checked")
	}
}

// BenchmarkRedundantTypeCheck judges an instanceof and a null check, in
// the three environments (default, proven, newest PHP).
func BenchmarkRedundantTypeCheck(b *testing.B) {
	benchTypeChecks(b, func(ctx *analysis.Context, args []syntax.Expr) {
		for _, x := range []syntax.Expr{args[0], args[3]} {
			if _, ok := redundantTypeCheck(ctx, x); !ok {
				b.Fatalf("%s not reported", ctx.Text(x))
			}
		}
		b.ReportAllocs()
		for b.Loop() {
			redundantTypeCheck(ctx, args[0])
			redundantTypeCheck(ctx, args[3])
		}
	})
}

// BenchmarkImpossibleTypeComparison judges a comparison of disjoint types
// and one of exact string values.
func BenchmarkImpossibleTypeComparison(b *testing.B) {
	benchTypeChecks(b, func(ctx *analysis.Context, args []syntax.Expr) {
		for _, x := range args[1:3] {
			if _, ok := impossibleTypeComparison(ctx, x); !ok {
				b.Fatalf("%s not reported", ctx.Text(x))
			}
		}
		b.ReportAllocs()
		for b.Loop() {
			impossibleTypeComparison(ctx, args[1])
			impossibleTypeComparison(ctx, args[2])
		}
	})
}
