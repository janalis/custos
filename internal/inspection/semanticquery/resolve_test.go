package semanticquery

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestResolvesToGlobalFunction(t *testing.T) {
	src := `<?php
namespace A { function time() {} time(); \time(); mktime(); }
namespace B { use function A\time; time(); nope(); }
`
	f := Parse(t, src)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	r := names.New(f)
	var calls []*syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			calls = append(calls, c)
		}
		return true
	})
	want := []struct {
		global string
		ok     bool
	}{{"time", false}, {"time", true}, {"mktime", true}, {"time", false}, {"nope", false}}
	if len(calls) != len(want) {
		t.Fatalf("got %d calls", len(calls))
	}
	for i, w := range want {
		if got := ResolvesToGlobalFunction(r, ix, phpversion.Max, calls[i], w.global); got != w.ok {
			t.Errorf("call %d (%s): got %v, want %v", i, Text(f, calls[i]), got, w.ok)
		}
	}
}

func TestResolvedFunctionFQNDynamic(t *testing.T) {
	f := Parse(t, `<?php $f();`)
	if _, ok := ResolvedFunctionFQN(names.New(f), index.New(stubs.Index()), phpversion.Max, FirstExpr(t, f).(*syntax.FuncCall)); ok {
		t.Fatal("dynamic call")
	}
}
