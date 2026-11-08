package index

import (
	"testing"

	phpversion "custos/internal/php/version"
)

type memoTestKey struct{ name string }

func TestMemoKeepsValuesOnlyOnFrozenIndexes(t *testing.T) {
	calls := 0
	count := func() any { calls++; return calls }
	var none *Index
	if none.Memo(memoTestKey{"a"}, count) != 1 || none.Memo(memoTestKey{"a"}, count) != 2 {
		t.Fatal("nil index cached a value")
	}
	ix := New(nil)
	if ix.Memo(memoTestKey{"a"}, count) != 3 || ix.Memo(memoTestKey{"a"}, count) != 4 {
		t.Fatal("editable index cached a value")
	}
	ix.Freeze()
	first := ix.Memo(memoTestKey{"a"}, count)
	again := ix.Memo(memoTestKey{"a"}, count)
	if first != 5 || again != 5 || ix.Memo(memoTestKey{"b"}, count) != 6 {
		t.Fatal("frozen index did not cache by key")
	}
}

func TestMemoDropsValuesWhenABaseLayerChanges(t *testing.T) {
	base := New(nil)
	ix := New(base)
	ix.Freeze()
	methods := func() any { return ix.FindMethod("Shared", "run", phpversion.PHP84) != nil }
	if ix.Memo(memoTestKey{"run"}, methods) != false {
		t.Fatal("method found before its declaration")
	}
	base.Add(extract(t, "shared.php", `<?php class Shared{function run(){}}`))
	if ix.Memo(memoTestKey{"run"}, methods) != true {
		t.Fatal("memo kept a value computed before the base layer changed")
	}
}
