package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestExpressionStoreNarrowBlockBoundaries(t *testing.T) {
	for _, capacity := range []int{0, 1, 2, 3, 5, 7, 8, 15, 16, 17, 10000} {
		store := newValueStore(capacity)
		width := min(capacity, valueBlockLimit)
		if width == 0 {
			width = valueBlockLimit
		}
		if len(store.blocks) != 0 {
			t.Fatal("hint eagerly allocated payloads")
		}
		exprs := make([]syntax.Expr, 97)
		for i := range exprs {
			if i > 0 {
				exprs[i] = &syntax.Literal{LitKind: syntax.LitInt}
			}
			store.set(exprs[i], Value{Complete: true, Identity: uint32(i + 1)})
			if len(store.blocks) != (i+width)/width {
				t.Fatalf("capacity%d entry%d: excess blocks", capacity, i)
			}
		}
		for i, expr := range exprs {
			value, known := store.get(expr)
			if !known || value.Identity != uint32(i+1) {
				t.Fatalf("capacity%d entry%d lost", capacity, i)
			}
		}
		for _, block := range store.blocks {
			if len(block) != width {
				t.Fatalf("capacity%d changed payload width", capacity)
			}
		}
		retained := store.at(exprs[width])
		blocks := len(store.blocks)
		for _, i := range []int{0, width - 1, width, 48, 96} {
			store.set(exprs[i], Value{Complete: true, Identity: 1000})
		}
		if len(store.blocks) != blocks || retained.Identity != uint32(width+1) {
			t.Fatal("update allocated or changed a returned copy")
		}
		extra := &syntax.Literal{LitKind: syntax.LitInt}
		store.set(extra, Value{Complete: true, Identity: 2000})
		if store.at(extra).Identity != 2000 || store.at(exprs[96]).Identity != 1000 || len(store.blocks) != (98+width-1)/width {
			t.Fatal("updates moved the insertion cursor")
		}
	}
}

func TestResolveKeepsExpressionAndScopeProofsSeparate(t *testing.T) {
	e, file := environment(t, `probe($x);`, nil)
	expr := argument(callsNamed(file, "probe")[0])
	scope := syntax.EnclosingVariableScope(expr)
	result := e.scope(scope)
	literal := &syntax.Literal{LitKind: syntax.LitInt, Raw: "42"}
	result.values.set(expr, Value{Expr: literal})
	result.complete = true
	if e.Value(expr).Complete {
		t.Fatal("incomplete value became complete")
	}
	if got, known := e.Resolve(expr); !known || got != literal {
		t.Fatal("value completeness replaced expression identity proof")
	}
	result.complete = false
	if got, known := e.Resolve(expr); known || got != literal {
		t.Fatal("incomplete scope established resolution or discarded its candidate")
	}
}
