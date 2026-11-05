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

func BenchmarkLocalArgumentPack(b *testing.B) {
	file := syntax.Parse("bench.php", []byte(`<?php $args=[CURLOPT_SSL_VERIFYPEER,false];curl_setopt($h,...$args);`), syntax.Options{Version: phpversion.Max})
	ix := index.New(stubs.Index())
	resolver := names.New(file)
	call := file.Stmts[1].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	b.ReportAllocs()
	for b.Loop() {
		infer.NewEnv(file, resolver, ix, phpversion.Max, infer.Options{}).BuiltinArguments(call)
	}
}
