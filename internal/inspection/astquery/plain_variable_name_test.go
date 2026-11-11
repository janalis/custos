package astquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestVariableName(t *testing.T) {
	f := syntax.Parse("v.php", []byte(`<?php f($a, (($b)), $$c, $d[0], 1);`), syntax.Options{})
	var args []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			for _, a := range call.Args.Args {
				args = append(args, a.(*syntax.Arg).Value)
			}
		}
		return true
	})
	want := []string{"a", "b", "", "", ""}
	for i, a := range args {
		if got := VariableName(a); got != want[i] {
			t.Errorf("argument %d: %q, want %q", i, got, want[i])
		}
	}
	if VariableName(nil) != "" || VariableName(f) != "" {
		t.Error("a node that is no expression has no variable name")
	}
}
