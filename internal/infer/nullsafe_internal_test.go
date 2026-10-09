package infer

import (
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestOrdinaryOffsetLeavesVariableBaseUncached(t *testing.T) {
	for _, source := range []string{`<?php function f($data) { if (is_array($data)) { $data[0]; } }`, `<?php function f($data) { if (is_array($data)) { ($data)[0]; } }`} {
		f := syntax.Parse("t.php", []byte(source), syntax.Options{Version: phpver.PHP84})
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := NewEnv(f, names.New(f), ix, phpver.PHP84)
		var dim *syntax.ArrayDimFetch
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if d, ok := n.(*syntax.ArrayDimFetch); ok {
				dim = d
			}
			return true
		})
		env.TypeOf(dim)
		variable := syntax.UnwrapParens(dim.Var).(*syntax.Variable)
		if _, primed := env.cache[variable]; primed {
			t.Fatal("ordinary offset primed narrowed receiver before baseType")
		}
		if _, computed := env.bases[variable]; !computed {
			t.Fatal("ordinary offset bypassed baseType")
		}
	}
}
