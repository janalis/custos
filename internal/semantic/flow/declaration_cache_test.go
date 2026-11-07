package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func TestMemberDeclarationCacheInvalidation(t *testing.T) {
	caller := syntax.Parse("caller.php", []byte(`<?php function caller(R $r){$r->take(1);}`), syntax.Options{Version: phpversion.Max})
	var call *syntax.MethodCall
	syntax.InspectFile(caller, func(node syntax.Node) bool {
		if method, ok := node.(*syntax.MethodCall); ok {
			call = method
		}
		return true
	})
	makeDefinitions := func(parameter string) *index.FileSymbols {
		file := syntax.Parse("definitions.php", []byte(`<?php final class R{function take($`+parameter+`){}}`), syntax.Options{Version: phpversion.Max})
		symbols := index.Extract(file)
		symbols.Classes[0].Methods["take"].Avail = index.Avail{From: phpversion.PHP81}
		return symbols
	}
	ix := index.New(stubs.Index())
	ix.Add(makeDefinitions("first"))
	e := New(caller, resourceTypes(caller, ix), nil, Options{})
	check := func(want string) {
		t.Helper()
		params, name, known := e.declaration(call)
		if want == "" {
			if known {
				t.Fatal("unavailable target remained cached")
			}
			return
		}
		if !known || name != "r::take" || len(params) != 1 || params[0].Name != want {
			t.Fatalf("%+v %s %v", params, name, known)
		}
	}
	check("first")
	check("first")
	e.types.PHP = phpversion.PHP74
	check("")
	e.types.PHP = phpversion.Max
	check("first")
	ix.Remove("definitions.php")
	ix.Add(makeDefinitions("updated"))
	check("updated")
	replacement := index.New(stubs.Index())
	replacement.Add(makeDefinitions("replacement"))
	e.types.Index = replacement
	check("replacement")
}

func TestUnresolvedMemberDeclarationsCanRefine(t *testing.T) {
	e, file := environment(t, `class R{} function f(R $r){$r->missing();}`, nil)
	var call *syntax.MethodCall
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if method, ok := node.(*syntax.MethodCall); ok {
			call = method
		}
		return true
	})
	if _, _, known := e.declaration(call); known {
		t.Fatal("missing method resolved")
	}
	if len(e.declarations.members) != 0 {
		t.Fatal("unresolved lookup retained")
	}
	if _, _, known := (&Env{}).declaration(&syntax.StaticCall{}); known {
		t.Fatal("malformed target resolved")
	}
}

func BenchmarkMemberDeclaration(b *testing.B) {
	e, file := environment(b, `final class R{function take($value){}}function f(R $r){$r->take(1);}`, nil)
	var call *syntax.MethodCall
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if method, ok := node.(*syntax.MethodCall); ok {
			call = method
		}
		return true
	})
	e.declaration(call)
	b.ReportAllocs()
	for b.Loop() {
		if _, _, known := e.declaration(call); !known {
			b.Fatal("lost declaration")
		}
	}
}
