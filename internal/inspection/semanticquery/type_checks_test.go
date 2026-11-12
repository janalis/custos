package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// probeArgs runs check on the arguments of every probe() call of src.
func probeArgs(t *testing.T, src string, check func(*analysis.Context, *syntax.FuncCall, []syntax.Expr)) {
	t.Helper()
	ran := false
	probeNative(t, src, func(ctx *analysis.Context, c *syntax.FuncCall) {
		var args []syntax.Expr
		for _, a := range c.Args.Args {
			args = append(args, a.(*syntax.Arg).Value)
		}
		check(ctx, c, args)
		ran = true
	})
	if !ran {
		t.Fatal("no probe() call")
	}
}

// Nodes that are not checks are answered no.
func TestTypeChecksOtherNodes(t *testing.T) {
	probeArgs(t, `function f(int $a) { probe($a); }`, func(ctx *analysis.Context, c *syntax.FuncCall, args []syntax.Expr) {
		if _, ok := DecidedTypeCheck(ctx, c); ok {
			t.Error("probe() is no check")
		}
		if _, ok := DecidedTypeCheck(ctx, args[0]); ok {
			t.Error("a variable is no check")
		}
	})
}

// A write changes the value it targets, the values that one reads through
// and its parts; destructuring writes each slot.
func TestTypeChecksWritesTo(t *testing.T) {
	src := `function f($a, $o, $other) { probe($a, $a['k'], [$a, $b], list(, $c), $o->p, $o->m()->q, $other, ($a)); }`
	probeArgs(t, src, func(ctx *analysis.Context, _ *syntax.FuncCall, x []syntax.Expr) {
		a, elem, pair, list, prop, chained, other, paren := x[0], x[1], x[2], x[3], x[4], x[5], x[6], x[7]
		for _, c := range []struct {
			target, x syntax.Expr
			want      bool
		}{
			{a, elem, true},
			{elem, a, true},
			{pair, a, true},
			{list, a, false},
			{other, a, false},
			{prop, chained, false},
			{syntax.UnwrapParens(paren), a, true},
			{nil, a, false},
		} {
			if got := writesTo(ctx, c.target, c.x); got != c.want {
				t.Errorf("writesTo(%s, %s) = %v", ctx.Text(c.target), ctx.Text(c.x), got)
			}
		}
		if accessBase(chained) == nil || accessBase(other) != nil {
			t.Error("accessBase")
		}
	})
}

// Exact operands: literals, and locals holding nothing else.
func TestTypeChecksExactOperand(t *testing.T) {
	src := `function f($c) { $a = 'x'; $i = 1; $i++; if ($c) { $u = 'y'; } $v = 'x'; if ($c) { $v = f($c); } probe($a, $i, $u, f($c), -1, (true), $v); }`
	probeArgs(t, src, func(ctx *analysis.Context, _ *syntax.FuncCall, x []syntax.Expr) {
		for i, want := range []bool{true, false, false, false, true, true, false} {
			if got := exactOperand(ctx, x[i]); got != want {
				t.Errorf("exactOperand(%s) = %v", ctx.Text(x[i]), got)
			}
		}
	})
	probeArgs(t, `$top = 'x'; probe($top);`, func(ctx *analysis.Context, _ *syntax.FuncCall, x []syntax.Expr) {
		if exactOperand(ctx, x[0]) {
			t.Error("a top-level local is not exact")
		}
	})
}

// Atom judgments of type tests and identity comparisons.
func TestTypeChecksAtoms(t *testing.T) {
	probeArgs(t, `final class A {} interface I {} function f() { probe(); }`, func(ctx *analysis.Context, _ *syntax.FuncCall, _ []syntax.Expr) {
		rel := typeRelations(ctx)
		y, n, m := types.Yes, types.No, types.Maybe
		for _, c := range []struct {
			pred, atom string
			want       types.Verdict
		}{
			{"null", "mixed", m},
			{"bool", "true", y},
			{`\I`, `\A`, n},
			{`\I`, "object", m},
			{`\I`, "int", n},
			{"iterable", "array", y},
			{"iterable", `\A`, n},
			{"iterable", "callable", m},
			{"iterable", "int", n},
			{"object", `\A`, y},
			{"object", "iterable", m},
			{"object", "int", n},
			{"scalar", "float", y},
			{"scalar", "callable", m},
			{"scalar", "null", n},
			{"string", "callable", m},
			{"array", "iterable", m},
			{"array", "int[]", y},
			{"int", "float", n},
		} {
			if got := atomPredicate(c.pred, c.atom, rel); got != c.want {
				t.Errorf("atomPredicate(%s, %s) = %v", c.pred, c.atom, got)
			}
		}
		if predicateVerdict("int", types.Of("int", "string"), rel) != m {
			t.Error("int|string is not always an int")
		}
		for _, c := range []struct {
			x, y string
			meet bool
		}{
			{"resource", "int", true},
			{"int", "int", true},
			{"bool", "true", true},
			{"false", "true", false},
			{"true", "bool", true},
			{"array", "callable", true},
			{"object", `\A`, true},
			{"int", "string", false},
			{`\A`, `\A`, true},
			{`\A`, `\I`, false},
		} {
			if got, _ := atomsMeet(c.x, c.y, types.Of(c.x), types.Of(c.y), rel); got != c.meet {
				t.Errorf("atomsMeet(%s, %s) = %v", c.x, c.y, got)
			}
		}
		if !sharesInt([]types.IntRange{{Lo: 1, Hi: 1}}, []types.IntRange{{Lo: 1, Hi: 1}}) || !sharesString([]string{"a"}, []string{"a"}) {
			t.Error("shared values")
		}
	})
}

// Declarations missing at a version print no return type; dynamic member
// names have none.
func TestTypeChecksVersionHelpers(t *testing.T) {
	if functionReturn(nil) != "" || methodReturn(nil) != "" {
		t.Error("missing declarations")
	}
	probeArgs(t, `function f($o, $m) { probe($o->$m()); }`, func(_ *analysis.Context, _ *syntax.FuncCall, x []syntax.Expr) {
		if identName(x[0].(*syntax.MethodCall).Name) != "" {
			t.Error("a dynamic name")
		}
	})
}
