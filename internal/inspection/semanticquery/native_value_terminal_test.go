package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeValueMatchesCandidateDiscovery(t *testing.T) {
	probeNative(t, `class C{public $p=2;const K=3;}function f(C $c,$a,$b){probe(1,'text',[1],call(),$c->call(),C::call(),new C(),@call(),"value $a",$a+1,$a[0],fn()=>1,$a?$b:1,$a??1,C::K,$c->p,UNKNOWN);}`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		for _, node := range call.Args.Args {
			expr := node.(*syntax.Arg).Value
			values, known := DiscoverValuesKnown(ctx.Types(), expr)
			var want syntax.Expr
			if known && len(values) == 1 {
				want = NativeUnsilenced(values[0])
			}
			if got := NativeValue(ctx, expr); got != want {
				t.Fatalf("%T: got %T, want %T", expr, got, want)
			}
		}
	})
	if NativeValue(nil, &syntax.Paren{}) != nil {
		t.Fatal("missing parenthesized expression became known")
	}
	if NativeValue(nil, &syntax.Literal{}) != nil {
		t.Fatal("recovery placeholder became known")
	}
}

func TestNativeTerminalNeedsNoEnvironment(t *testing.T) {
	file := syntax.Parse("test.php", []byte(`<?php 'text';`), syntax.Options{})
	expr := file.Stmts[0].(*syntax.ExprStmt).Expr
	if NativeValue(nil, expr) != expr {
		t.Fatal("literal candidate changed")
	}
}

func BenchmarkNativeTerminalValue(b *testing.B) {
	file := syntax.Parse("test.php", []byte(`<?php 'text';`), syntax.Options{})
	expr := file.Stmts[0].(*syntax.ExprStmt).Expr
	b.ReportAllocs()
	for b.Loop() {
		if NativeValue(nil, expr) != expr {
			b.Fatal("lost candidate")
		}
	}
}
