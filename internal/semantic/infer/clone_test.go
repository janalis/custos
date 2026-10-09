package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

func TestClonePHP85(t *testing.T) {
	for _, tc := range []struct{ expression, want string }{
		{`clone new C`, `\C`},
		{`clone(new C)`, `\C`},
		{`clone(withProperties: [], object: new C)`, `\C`},
		{`clone(...)`, `\Closure`},
		{`(clone(...))(new C)`, `object`},
		{`clone(...$args)`, `?unknown`},
		{`clone(new C, ...$args)`, `?unknown`},
		{`clone()`, `?unknown`},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			f := syntax.Parse("clone.php", []byte("<?php class C {} "+tc.expression+";"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) > 0 {
				t.Fatal(f.Errors)
			}
			ix := index.New(nil)
			ix.Add(index.Extract(f))
			env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP85)
			expr := f.Stmts[1].(*syntax.ExprStmt).Expr
			if got := env.TypeOf(expr).String(); got != tc.want {
				t.Errorf("got %s want %s", got, tc.want)
			}
		})
	}
}

func BenchmarkCloneInference(b *testing.B) {
	f := syntax.Parse("clone.php", []byte(`<?php class C {} clone(withProperties: [], object: new C);`), syntax.Options{Version: phpversion.PHP85})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	resolver := names.New(f)
	expr := f.Stmts[1].(*syntax.ExprStmt).Expr
	b.ReportAllocs()
	for b.Loop() {
		infer.NewEnv(f, resolver, ix, phpversion.PHP85).TypeOf(expr)
	}
}
