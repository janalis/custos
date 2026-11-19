package infer_test

import (
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/infer"
	"github.com/janalis/custos/internal/semantic/names"
	"github.com/janalis/custos/internal/semantic/stubs"
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
