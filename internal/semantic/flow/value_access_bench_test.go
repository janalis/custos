package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func BenchmarkExpressionStoreRead(b *testing.B) {
	for _, width := range []int{1, 3, 16} {
		b.Run(fmt.Sprint(width), func(b *testing.B) {
			exprs := make([]syntax.Expr, 128)
			store := newValueStore(width)
			for i := range exprs {
				exprs[i] = &syntax.Literal{LitKind: syntax.LitInt}
				store.set(exprs[i], Value{Complete: true, Identity: uint32(i + 1)})
			}
			b.ReportAllocs()
			for b.Loop() {
				for i, expr := range exprs {
					if store.at(expr).Identity != uint32(i+1) {
						b.Fatal("value changed")
					}
				}
			}
		})
	}
}

func BenchmarkResolveNestedRead(b *testing.B) {
	e, file := environment(b, `function f(){$x=42;probe(`+strings.Repeat("(", 32)+`$x`+strings.Repeat(")", 32)+`);}`, nil)
	expr := syntax.UnwrapParens(argument(callsNamed(file, "probe")[0]))
	if _, known := e.Resolve(expr); !known {
		b.Fatal("missing initial resolution")
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, known := e.Resolve(expr); !known {
			b.Fatal("resolution changed")
		}
	}
}
