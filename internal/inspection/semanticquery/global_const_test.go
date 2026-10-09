package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestGlobalConstName(t *testing.T) {
	src := `<?php
namespace A { const PHP_OS = 'x'; echo PHP_OS, PHP_VERSION, \PHP_OS, Foo\PHP_VERSION; }
namespace B { use const C\PHP_EOL; use const PHP_SAPI; echo PHP_EOL, PHP_SAPI, namespace\M_PI; }
namespace { echo PHP_VERSION, \Foo\PHP_VERSION; }
`
	e, err := analysis.NewEngine(nil, analysis.Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ctx := analysis.NewTestContext(e, f)
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ConstFetch); ok {
			got = append(got, GlobalConstName(ctx, c))
		}
		return true
	})
	want := []string{"", "PHP_VERSION", "PHP_OS", "", "", "PHP_SAPI", "", "PHP_VERSION", ""}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestGlobalConstNameNil(t *testing.T) {
	if GlobalConstName(nil, nil) != "" {
		t.Fatal("nil constant")
	}
}
