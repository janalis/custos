package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// BenchmarkMissingMember looks up a method a typed receiver lacks, on a
// non-final class with a subclass and a trait, as UndefinedMethod does.
func BenchmarkMissingMember(b *testing.B) {
	const src = `<?php
namespace App;
trait Stamps { public int $at = 0; public function stamp(): void {} }
class Order { use Stamps; public int $cents = 0; public function pay(): void {} }
final class Rush extends Order { public function hurry(): void {} }
function settle(Order $order) {
    probe($order->refund());
}`
	has := func(c *index.Class) bool { return c.Methods["refund"] != nil || c.Methods["__call"] != nil }
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		if name, ok := call.Name.(*syntax.Name); !ok || name.Value != "probe" {
			return
		}
		method := call.Args.Args[0].(*syntax.Arg).Value.(*syntax.MethodCall)
		if label, ok := MissingMember(ctx, MemberReceiverClasses(ctx, method.Var), false, has); !ok || label != `App\Order` {
			b.Fatalf("missing member = %q, %v", label, ok)
		}
		if MemberGuarded(ctx, method, "refund", false) {
			b.Fatal("unguarded call reported guarded")
		}
		ran = true
		b.ReportAllocs()
		for b.Loop() {
			MissingMember(ctx, MemberReceiverClasses(ctx, method.Var), false, has)
			MemberGuarded(ctx, method, "refund", false)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("bench.php", []byte(src), syntax.Options{}))
	if !ran {
		b.Fatal("no member access checked")
	}
}
