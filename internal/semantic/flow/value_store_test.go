package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestValueStoreGrowthAndCopies(t *testing.T) {
	var store valueStore
	if value, known := store.get(nil); known || value.Complete {
		t.Fatal("missing entry became known")
	}
	expressions := make([]syntax.Expr, 1000)
	for i := range expressions {
		expressions[i] = &syntax.Literal{LitKind: syntax.LitInt}
		store.set(expressions[i], Value{Complete: true, Identity: uint32(i + 1)})
	}
	for i, expr := range expressions {
		if value, known := store.get(expr); !known || !value.Complete || value.Identity != uint32(i+1) {
			t.Fatalf("entry %d corrupted by growth", i)
		}
	}
	old := store.at(expressions[0])
	store.set(expressions[0], Value{Complete: true, Identity: 2000})
	if old.Identity != 1 || store.at(expressions[0]).Identity != 2000 {
		t.Fatal("updating storage changed a retained copy")
	}
}

func TestFrameRepeatedFactsRemainIsolated(t *testing.T) {
	original := frame{}
	original.resetVars()
	value := Value{Complete: true, Safe: HTML}
	original.setVar("value", value)
	branch := clone(&original)
	branch.setVar("value", value)
	branch.setVar("value", Value{Complete: true, Sources: []Source{{Kind: "request", Parameter: -1}}})
	if original.vars["value"].Safe != HTML || tainted(original.vars["value"]) || !tainted(branch.vars["value"]) {
		t.Fatal("a repeated fact broke branch isolation")
	}
	original.setVar("other", value)
	if _, known := branch.vars["other"]; known {
		t.Fatal("new binding leaked across branches")
	}
}

func BenchmarkSharedFrameRepeatedFact(b *testing.B) {
	f := frame{}
	f.resetVars()
	value := Value{Complete: true, Safe: HTML}
	f.setVar("value", value)
	b.ReportAllocs()
	for b.Loop() {
		branch := clone(&f)
		branch.setVar("value", value)
		if !branch.vars["value"].Complete {
			b.Fatal("lost fact")
		}
	}
}

func TestValueStoreReservesOnlyReachedPayloads(t *testing.T) {
	for _, capacity := range []int{0, 1, 10000} {
		store := newValueStore(capacity)
		if len(store.blocks) != 0 {
			t.Fatal("unreached expressions allocated payloads")
		}
		store.set(nil, Value{Complete: true})
		if len(store.blocks) != 1 || len(store.blocks[0]) > valueBlockLimit || !store.at(nil).Complete {
			t.Fatal("payload allocation exceeded one block")
		}
	}
}

func BenchmarkExpressionValueStore(b *testing.B) {
	expressions := make([]syntax.Expr, 128)
	for i := range expressions {
		expressions[i] = &syntax.Literal{LitKind: syntax.LitInt}
	}
	b.ReportAllocs()
	for b.Loop() {
		store := newValueStore(len(expressions))
		for i, expr := range expressions {
			store.set(expr, Value{Complete: true, Identity: uint32(i + 1)})
		}
		for i, expr := range expressions {
			if store.at(expr).Identity != uint32(i+1) {
				b.Fatal("lost value")
			}
		}
	}
}
