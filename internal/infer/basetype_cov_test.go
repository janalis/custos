package infer

import (
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// baseType on a variable being typed (recursion) is unknown and records
// nothing, as TypeOf.
func TestBaseTypeRecursionGuard(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php $a = ['k' => 1]; $a['j'] = 2; echo $a['k'];`), syntax.Options{Version: phpver.PHP84})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	e := NewEnv(f, names.New(f), ix, phpver.PHP84)
	var last *syntax.Variable
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok {
			last = v
		}
		return true
	})
	e.busy[last] = true
	if got := e.baseType(last); !got.IsUnknown() {
		t.Errorf("busy: %s", got)
	}
	if _, ok := e.bases[last]; ok {
		t.Error("busy: nothing recorded")
	}
	delete(e.busy, last)
	if got := e.baseType(last).ShapeString(); got != "int[]{k: int}" {
		t.Errorf("base: %s", got)
	}
	if got := e.TypeOf(last).ShapeString(); got != "non-empty int[]" {
		t.Errorf("whole: %s", got)
	}
}
