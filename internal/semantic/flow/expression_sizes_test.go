package flow

import (
	"sync"
	"testing"

	"custos/internal/php/syntax"
)

func TestExpressionSizesMemo(t *testing.T) {
	for _, source := range []string{
		`<?php $a = 1; function f($x) { return $x + 1; }`,
		`<?php $b = fn($v) => $v; class C { function f() { if ($x) { echo $x; } } }`,
	} {
		file := syntax.Parse("same.php", []byte(source), syntax.Options{})
		e := &Env{file: file}
		scopes := []syntax.Node{nil}
		syntax.InspectFile(file, func(n syntax.Node) bool {
			if syntax.IsVariableScope(n) {
				scopes = append(scopes, n)
			}
			return true
		})
		for _, scope := range scopes {
			values, before := e.countExpressions(scope)
			var wg sync.WaitGroup
			for range 4 {
				wg.Go(func() {
					// Separate environments share only immutable syntax-owned memo data.
					other := &Env{file: file}
					for range 3 {
						gotValues, gotBefore := other.expressionSizes(scope)
						if gotValues != values || gotBefore != before {
							t.Errorf("%T: got %d/%d, want %d/%d", scope, gotValues, gotBefore, values, before)
						}
					}
				})
			}
			wg.Wait()
		}
	}
}

func BenchmarkFlowAllocationHintsUncached(b *testing.B) {
	e, file := environment(b, `class R{function f($x){if($x){return $this->g($x);}return 1;}function g($v){return $v;}}`, nil)
	var scope syntax.Node
	syntax.InspectFile(file, func(n syntax.Node) bool {
		if method, ok := n.(*syntax.Method); ok && method.Name.Value == "f" {
			scope = method
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		e.countExpressions(scope)
	}
}
