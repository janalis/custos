package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestCloneArgumentTree(t *testing.T) {
	f := parse(t, `<?php clone(withProperties: overrides(), object: source(), extra: nested($v));`, phpversion.PHP85)
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	n := firstExpr(t, f).(*Clone)
	seen := map[Node]bool{}
	calls := map[string]bool{}
	Inspect(n, func(child Node) bool {
		if seen[child] {
			t.Fatalf("duplicate visit of %T", child)
		}
		seen[child] = true
		if !n.Span().Contains(child.Span()) {
			t.Fatalf("child %T outside clone span", child)
		}
		if child != n && (child.Parent() == nil || !child.Parent().Span().Contains(child.Span())) {
			t.Fatalf("invalid parent of %T", child)
		}
		if call, ok := child.(*FuncCall); ok {
			calls[call.Name.(*Name).Value] = true
		}
		return true
	})
	if !calls["source"] || !calls["overrides"] || !calls["nested"] || len(calls) != 3 {
		t.Fatalf("lost call subtree: %v", calls)
	}
	if n.Args.Parent() != n || n.Expr.Parent() != n.Args.Args[1] || n.With.Parent() != n.Args.Args[0] {
		t.Fatal("aliases must retain their argument parents")
	}
	arg := n.Args.Args[2].(*Arg)
	if arg.Name.Value != "extra" || arg.Parent() != n.Args {
		t.Fatal("extra argument metadata lost")
	}
	f = parse(t, `<?php clone(...$args);`, phpversion.PHP85)
	unpacked := firstExpr(t, f).(*Clone).Args.Args[0].(*Arg)
	if !unpacked.Unpack || unpacked.Value.(*Variable).Name != "args" {
		t.Fatal("unpack metadata lost")
	}
}
