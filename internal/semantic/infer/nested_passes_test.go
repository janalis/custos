package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
)

// Nested writes still changing the element type after the last pass leave
// it unknown.
func TestNestedWritesPassLimit(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php function f($k) { $a = []; $a[$k]['x'][] = 'v'; echo $a[$k]; }`), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	var last *syntax.ArrayDimFetch
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if d, ok := n.(*syntax.ArrayDimFetch); ok {
			last = d
		}
		return true
	})
	defer func(n int) { maxNestedPasses = n }(maxNestedPasses)
	for _, tc := range []struct {
		passes int
		want   string
	}{{1, "?unknown"}, {4, "array"}} {
		maxNestedPasses = tc.passes
		e := NewEnv(f, names.New(f), ix, phpversion.PHP84, Options{})
		if got := e.TypeOf(last).ShapeString(); got != tc.want {
			t.Errorf("%d passes: %s", tc.passes, got)
		}
	}
}
