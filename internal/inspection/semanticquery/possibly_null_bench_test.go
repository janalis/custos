package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

const possiblyNullBenchSrc = `<?php
declare(strict_types=1);
namespace App;
final class Order { public function total(): int { return 0; } }
function ship(Order $order, string $note) {}
function find(): ?Order { return null; }
function pay(?Order $order, string $note) {
    if ($note === '') {
        return;
    }
    ship($order, $note);
    $last = find();
    echo $last->total();
}`

// benchPossiblyNull runs check on the ship() call of possiblyNullBenchSrc.
func benchPossiblyNull(b *testing.B, check func(ctx *analysis.Context, call *syntax.FuncCall)) {
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		if name, ok := call.Name.(*syntax.Name); ok && name.Value == "ship" {
			check(ctx, call)
			ran = true
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(possiblyNullBenchSrc), syntax.Options{}))
	if !ran {
		b.Fatal("no ship() call checked")
	}
}

// BenchmarkPossiblyNullArguments checks the arguments of one call: a
// nullable parameter passed on, a narrowed string.
func BenchmarkPossiblyNullArguments(b *testing.B) {
	benchPossiblyNull(b, func(ctx *analysis.Context, call *syntax.FuncCall) {
		if got := possiblyNullArguments(ctx, call); len(got) != 1 || got[0].Param != "order" {
			b.Fatalf("arguments = %+v", got)
		}
		b.ReportAllocs()
		for b.Loop() {
			possiblyNullArguments(ctx, call)
		}
	})
}

// BenchmarkPossiblyNullAccess checks one method call on a local holding a
// nullable result.
func BenchmarkPossiblyNullAccess(b *testing.B) {
	benchPossiblyNull(b, func(ctx *analysis.Context, _ *syntax.FuncCall) {
		var access *syntax.MethodCall
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if m, ok := n.(*syntax.MethodCall); ok {
				access = m
			}
			return true
		})
		if _, ok := possiblyNullAccess(ctx, access); !ok {
			b.Fatal("access not reported")
		}
		b.ReportAllocs()
		for b.Loop() {
			possiblyNullAccess(ctx, access)
		}
	})
}
