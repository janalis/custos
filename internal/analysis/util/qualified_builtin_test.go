package util

import (
	"testing"

	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestQualifiedBuiltin(t *testing.T) {
	src := `<?php
namespace A { function dirname($p) {} dirname($x); }
namespace B { use function C\is_dir; is_dir($x); \mkdir($x); mkdir($x); }
namespace C { use function is_dir; is_dir($x); }
namespace { dirname($x); }
`
	e, err := analysis.NewEngine(nil, analysis.Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ctx := analysis.NewTestContext(e, f)
	var calls []*syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			calls = append(calls, c)
		}
		return true
	})
	want := []struct{ name, got string }{
		{"dirname", `\dirname`}, {"is_dir", `\is_dir`}, {"mkdir", `\mkdir`}, {"mkdir", "mkdir"},
		{"is_dir", "is_dir"}, {"dirname", "dirname"},
	}
	if len(calls) != len(want) {
		t.Fatalf("got %d calls", len(calls))
	}
	for i, w := range want {
		if got := QualifiedBuiltinFor(ctx, w.name, calls[i]); got != w.got {
			t.Errorf("call %d: got %q, want %q", i, got, w.got)
		}
	}
	if got := QualifiedBuiltin(ctx, "strlen", calls[0].Span().Start); got != "strlen" {
		t.Errorf("strlen in A: got %q", got)
	}
}
