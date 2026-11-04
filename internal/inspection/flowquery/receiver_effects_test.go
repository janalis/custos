package flowquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestMethodReceiverEffect(t *testing.T) {
	for _, tc := range []struct {
		expr, root  string
		conditional bool
	}{
		{`$sink->push($row)`, `$sink`, false},
		{`$sink?->push($row)`, `$sink`, true},
		{`($sink->children[0])->push($row)`, `$sink`, false},
		{`$sink?->children->fluent()->push($row)`, `$sink`, true},
		{`$sink?->fluent()->push($row)`, `$sink`, true},
	} {
		f := Parse(t, `<?php `+tc.expr+`;`)
		call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.MethodCall)
		got := MethodReceiverEffect(call)
		if Text(f, got.Root) != tc.root || got.Conditional != tc.conditional {
			t.Fatalf("%s: %+v", tc.expr, got)
		}
	}
	if MethodReceiverEffect(nil).Root != nil {
		t.Fatal("nil call")
	}
	var root syntax.Expr = &syntax.Variable{Name: "x"}
	for range MaxAssignScan {
		root = &syntax.MethodCall{Var: root}
	}
	if MethodReceiverEffect(&syntax.MethodCall{Var: root}).Root != nil {
		t.Fatal("unbounded chain")
	}
}

func BenchmarkMethodReceiverEffect(b *testing.B) {
	f := Parse(b, `<?php $sink?->children->fluent()->push($row);`)
	call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.MethodCall)
	b.ReportAllocs()
	for b.Loop() {
		MethodReceiverEffect(call)
	}
}
