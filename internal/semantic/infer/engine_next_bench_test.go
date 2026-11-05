package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func BenchmarkBuiltinBindingAndCount(b *testing.B) {
	for _, src := range []string{`hrtime(...['as_number'=>true])`, `count(array_reverse(array_values(['a'=>1,'b'=>2])))`} {
		b.Run(src, func(b *testing.B) {
			file := syntax.Parse("bench.php", []byte("<?php "+src+";"), syntax.Options{Version: phpversion.Max})
			ix := index.New(stubs.Index())
			resolver := names.New(file)
			expr := file.Stmts[0].(*syntax.ExprStmt).Expr
			b.ReportAllocs()
			for b.Loop() {
				infer.NewEnv(file, resolver, ix, phpversion.Max, infer.Options{}).TypeOf(expr)
			}
		})
	}
}
