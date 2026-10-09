package flowquery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestVarAccessesByNameMatches checks the one-walk variant against the
// per-name walk on every own fixture, for every scope and name.
func TestVarAccessesByNameMatches(t *testing.T) {
	paths, _ := filepath.Glob("../../../testdata/rules/*/*.php")
	if len(paths) == 0 {
		t.Skip("no fixtures")
	}
	paths = append(paths, "") // extra inline source
	for _, p := range paths {
		src := []byte("<?php function f($a) { $x = fn($a) => $a + $x + fn($x) => $x + $y; $y = 1; $u = function () use ($x, &$y) { $z = 1; }; }")
		if p != "" {
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			src = b
		}
		f := syntax.Parse(p, src, syntax.Options{Version: phpversion.PHP84, Permissive: true})
		scopes := []syntax.Node{nil}
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
				scopes = append(scopes, n)
			}
			return true
		})
		names := map[string]bool{}
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil {
				names[v.Name] = true
			}
			return true
		})
		for _, sc := range scopes {
			all := VarAccessesByName(f, sc)
			for name := range names {
				want := VarAccesses(f, sc, name)
				if got := all[name]; !(len(got) == 0 && len(want) == 0) && !reflect.DeepEqual(got, want) {
					t.Fatalf("%s: scope %T name %s: got %d accesses, want %d", p, sc, name, len(got), len(want))
				}
			}
		}
	}
}
