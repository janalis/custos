package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// BenchmarkDeclarationViews types the same declarations through an Env,
// its native and proven twins and the T-rules: documented property reads,
// and calls of functions whose return docs name class constants, as rules
// asking several views of one file do.
func BenchmarkDeclarationViews(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php class K { const A_X = 'x'; const A_Y = 'y'; }\ninterface I {} class M implements I {}\nclass G {\n")
	for p := range 10 {
		fmt.Fprintf(&src, "    /** @var list<M> */ public array $p%d = [];\n    /** @var M */ public I $q%d;\n", p, p)
	}
	src.WriteString("}\n")
	for f := range 10 {
		fmt.Fprintf(&src, "/** @return K::A_* */ function f%d(): string { return 'x'; }\n", f)
	}
	src.WriteString("function main(G $g): void {\n")
	for i := range 30 {
		fmt.Fprintf(&src, "    $g->p%[1]d; $g->q%[1]d; f%[1]d(); strlen('x'); array_keys([]);\n", i%10)
	}
	src.WriteString("}\n")
	file := syntax.Parse("bench.php", []byte(src.String()), syntax.Options{Version: phpversion.PHP84})
	if len(file.Errors) > 0 {
		b.Fatal(file.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(file))
	resolver := names.New(file)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(file, resolver, ix, phpversion.PHP84, infer.Options{})
		views := []*infer.Env{env, env.Native(), env.Proven(), env.WithOptions(infer.Options{Assertions: true})}
		tr := infer.NewTRules(env)
		syntax.InspectFile(file, func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.PropertyFetch:
				for _, v := range views {
					v.TypeOf(x)
				}
			case *syntax.FuncCall:
				for _, v := range views {
					v.TypeOf(x)
				}
				tr.TypeOf(x)
			}
			return true
		})
	}
}
