package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// BenchmarkPropertyAssignmentMismatch checks one rejected write to a typed
// property of a non-final class with a subclass, through a typed receiver.
func BenchmarkPropertyAssignmentMismatch(b *testing.B) {
	const src = `<?php
declare(strict_types=1);
namespace App;
class Order { public int $cents = 0; public ?string $note = null; }
final class Rush extends Order {}
function pay(Order $order) {
    $label = 'x';
    probe($order->cents = $label);
}`
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		if name, ok := call.Name.(*syntax.Name); !ok || name.Value != "probe" {
			return
		}
		assign := call.Args.Args[0].(*syntax.Arg).Value
		if got, ok := propertyAssignmentMismatch(ctx, assign); !ok || !got.Proven {
			b.Fatalf("mismatch = %+v, %v", got, ok)
		}
		ran = true
		b.ReportAllocs()
		for b.Loop() {
			propertyAssignmentMismatch(ctx, assign)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(src), syntax.Options{}))
	if !ran {
		b.Fatal("no property write checked")
	}
}

// BenchmarkForeachPropertyMismatches checks a loop binding rejected keys
// and destructured elements to typed properties through a typed receiver.
func BenchmarkForeachPropertyMismatches(b *testing.B) {
	const src = `<?php
declare(strict_types=1);
namespace App;
class Order { public int $cents = 0; public ?string $note = null; }
final class Rush extends Order {}
function pay(Order $order) {
    foreach ([['x', 1]] as $order->note => [$order->cents, $order->note]) { probe(); }
}`
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		if name, ok := call.Name.(*syntax.Name); !ok || name.Value != "probe" {
			return
		}
		var fe *syntax.Foreach
		for n := syntax.Node(call); fe == nil; n = n.Parent() {
			fe, _ = n.(*syntax.Foreach)
		}
		if got := foreachMismatches(ctx, fe); len(got) != 3 || !got[1].Proven {
			b.Fatalf("mismatches = %+v", got)
		}
		ran = true
		b.ReportAllocs()
		for b.Loop() {
			foreachMismatches(ctx, fe)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(src), syntax.Options{}))
	if !ran {
		b.Fatal("no property write checked")
	}
}
