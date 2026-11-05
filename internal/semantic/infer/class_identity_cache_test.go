package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

func TestClassIdentitiesStayFileLocal(t *testing.T) {
	for _, path := range []string{"first.php", "second.php"} {
		file := syntax.Parse(path, []byte(`<?php namespace Example;class Named{} $a=new class{};`), syntax.Options{Version: phpversion.Max})
		resolver := names.New(file)
		env := infer.NewEnv(file, resolver, index.New(nil), phpversion.Max, infer.Options{})
		if env.ClassFQN(nil) != "" {
			t.Fatal("nil declaration acquired identity")
		}
		syntax.InspectFile(file, func(node syntax.Node) bool {
			if class, ok := node.(*syntax.ClassLike); ok {
				want := resolver.SymbolFQN(class)
				for range 2 {
					if got := env.ClassFQN(class); got != want {
						t.Fatalf("%s: %q != %q", path, got, want)
					}
				}
				if got := env.Native().ClassFQN(class); got != want {
					t.Fatal("native environment changed class identity")
				}
			}
			return true
		})
	}
}

func BenchmarkClassIdentity(b *testing.B) {
	file := syntax.Parse("bench.php", []byte(`<?php namespace Example;class Named{}`), syntax.Options{Version: phpversion.Max})
	resolver := names.New(file)
	env := infer.NewEnv(file, resolver, index.New(nil), phpversion.Max, infer.Options{})
	var class *syntax.ClassLike
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if candidate, ok := node.(*syntax.ClassLike); ok {
			class = candidate
		}
		return true
	})
	env.ClassFQN(class)
	b.ReportAllocs()
	for b.Loop() {
		if env.ClassFQN(class) != "Example\\Named" {
			b.Fatal("class identity changed")
		}
	}
}
