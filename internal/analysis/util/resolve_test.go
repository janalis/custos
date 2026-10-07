package util

import (
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestResolvesToGlobalFunction(t *testing.T) {
	src := `<?php
namespace A { function time() {} time(); \time(); mktime(); }
namespace B { use function A\time; time(); nope(); }
`
	f := parse(t, src)
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
		if got := ResolvesToGlobalFunction(r, ix, phpver.Max, calls[i], w.global); got != w.ok {
			t.Errorf("call %d (%s): got %v, want %v", i, text(f, calls[i]), got, w.ok)
		}
	}
}
