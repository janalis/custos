package splfixedarrayshrinkdiscardsvalues

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestStateScopeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{
		{`if($x)unset($x);probe();`, false},
		{`function first(){unset($x);}function second(){probe();}`, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			file := syntax.Parse("scope.php", []byte("<?php "+tc.source), syntax.Options{})
			var before *syntax.Unset
			var at *syntax.FuncCall
			syntax.InspectFile(file, func(n syntax.Node) bool {
				switch n := n.(type) {
				case *syntax.Unset:
					before = n
				case *syntax.FuncCall:
					at = n
				}
				return true
			})
			if before == nil || at == nil {
				t.Fatal("missing operations")
			}
			if got := stateDominates(before, at); got != tc.want {
				t.Fatalf("dominates=%v", got)
			}
		})
	}
}
