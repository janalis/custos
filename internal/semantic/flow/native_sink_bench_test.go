package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func BenchmarkNativeSinkSelection(b *testing.B) {
	for _, name := range []string{"strlen", "header", "curl_setopt"} {
		b.Run(name, func(b *testing.B) {
			e, file := environment(b, name+"('text');", nil)
			call := callsNamed(file, name)[0]
			values := []Value{{Complete: true, Expr: call.Args.Args[0].(*syntax.Arg).Value}}
			b.ReportAllocs()
			for b.Loop() {
				e.nativeSinks(Call{Node: call, Arguments: values})
			}
		})
	}
}
