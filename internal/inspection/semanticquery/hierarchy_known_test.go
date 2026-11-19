package semanticquery

import (
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/stubs"
)

func TestHierarchyResolved(t *testing.T) {
	src := `<?php
namespace N;
interface I {}
trait T {}
class A implements I { use T; }
class B extends A {}
class P extends Missing\Base {}
class Q implements Missing\Iface {}
class R { use Missing\Tr; }
class S extends P {}
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	for fqn, want := range map[string]bool{
		`N\B`: true, `N\A`: true, `N\P`: false, `N\Q`: false, `N\R`: false, `N\S`: false, `N\Unknown`: true,
	} {
		if got := HierarchyResolved(ix, fqn, 0); got != want {
			t.Errorf("HierarchyResolved(%s) = %v", fqn, got)
		}
	}
}
