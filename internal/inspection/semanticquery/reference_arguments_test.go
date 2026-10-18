package semanticquery

import (
	"fmt"
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func TestReferenceArguments(t *testing.T) {
	params := []index.Param{{Name: "out", ByRef: true}, {Name: "ignored"}}
	for _, src := range []string{`f($x,0)`, `f(ignored:0,out:$x)`, `f(...$args,out:$x)`, `f(...$args,$x)`, `f(ignored:$x)`, `f(...)`} {
		t.Run(src, func(t *testing.T) {
			f := syntax.Parse("args.php", []byte("<?php "+src+";"), syntax.Options{})
			call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
			got := ReferenceArguments(call.Args, params, true)
			at := 0
			for _, node := range call.Args.Args {
				arg, ok := node.(*syntax.Arg)
				if !ok {
					continue
				}
				ref, known := ArgumentBindsByRef(call.Args, arg, params, true)
				if ref {
					if at >= len(got) || got[at].Argument != arg || got[at].Known != known {
						t.Fatalf("binding %d: %+v", at, got)
					}
					at++
				}
			}
			if at != len(got) {
				t.Fatalf("extra bindings: %+v", got)
			}
		})
	}
	if got := ReferenceArguments(nil, params, true); got != nil {
		t.Fatal("nil list has bindings")
	}
	arg := &syntax.Arg{ByRef: true}
	list := &syntax.ArgList{Args: []syntax.Expr{arg}}
	if got := ReferenceArguments(list, nil, true); len(got) != 1 || !got[0].Known {
		t.Fatalf("call-time reference: %+v", got)
	}
	if got := ReferenceArguments(list, nil, false); len(got) != 0 {
		t.Fatal("disabled call-time reference")
	}
}

func BenchmarkReferenceArguments(b *testing.B) {
	for _, count := range []int{100, 10000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			list := &syntax.ArgList{Args: make([]syntax.Expr, count)}
			params := make([]index.Param, count)
			for i := range list.Args {
				list.Args[i] = &syntax.Arg{}
			}
			params[count-1].ByRef = true
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				ReferenceArguments(list, params, true)
			}
		})
	}
}
