package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

func TestResolvedFunctionCacheTracksDependencies(t *testing.T) {
	f := syntax.Parse("caller.php", []byte(`<?php namespace App; target();`), syntax.Options{})
	call := f.Stmts[0].(*syntax.Namespace).Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	base := index.New(nil)
	ix := index.New(base)
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
	for range 2 {
		if env.ResolveFunction(call) != nil {
			t.Fatal("negative cache must start unresolved")
		}
	}
	base.Add(&index.FileSymbols{Path: "global.php", Functions: []*index.Function{{FQN: "target", Return: "int"}}})
	for range 2 {
		if fn := env.ResolveFunction(call); fn == nil || fn.FQN != "target" {
			t.Fatalf("global fallback lost: %+v", fn)
		}
	}
	ix.Add(&index.FileSymbols{Path: "local.php", Functions: []*index.Function{{FQN: `App\target`, Return: "string"}}})
	if fn := env.ResolveFunction(call); fn == nil || fn.FQN != `App\target` {
		t.Fatalf("local shadow ignored: %+v", fn)
	}
	ix.Remove("local.php")
	if fn := env.ResolveFunction(call); fn == nil || fn.FQN != "target" {
		t.Fatalf("removed shadow retained: %+v", fn)
	}
	env.Index = index.New(nil)
	if env.ResolveFunction(call) != nil {
		t.Fatal("cached declaration survived index replacement")
	}
}

func TestResolvedFunctionCacheTracksPHPVersion(t *testing.T) {
	f := syntax.Parse("version.php", []byte(`<?php versioned();`), syntax.Options{})
	call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	ix := index.New(nil)
	ix.Add(&index.FileSymbols{Path: "versioned.php", Functions: []*index.Function{{FQN: "versioned", Return: "string", RetVer: []index.VerType{{Until: phpversion.PHP80, Type: "int"}}}}})
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP74, infer.Options{})
	if fn := env.ResolveFunction(call); fn.Return != "int" {
		t.Fatal(fn.Return)
	}
	env.PHP = phpversion.PHP84
	if fn := env.ResolveFunction(call); fn.Return != "string" {
		t.Fatal(fn.Return)
	}
}

func BenchmarkResolvedFunction(b *testing.B) {
	file := syntax.Parse("caller.php", []byte(`<?php namespace App; target();`), syntax.Options{})
	call := file.Stmts[0].(*syntax.Namespace).Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	ix := index.New(nil)
	ix.Add(&index.FileSymbols{Path: "target.php", Functions: []*index.Function{{FQN: "target", Return: "int"}}})
	env := infer.NewEnv(file, names.New(file), ix, phpversion.PHP84, infer.Options{})
	b.ReportAllocs()
	for b.Loop() {
		env.ResolveFunction(call)
	}
}
