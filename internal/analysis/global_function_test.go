package analysis

import (
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestGlobalFunctionName(t *testing.T) {
	src := `<?php
namespace A { function strlen($s) { return 0; } strlen($x); STRLEN($x); Trim($x); }
namespace B { use function C\trim; trim($x); \Trim($x); D\count($x); count($x); $f($x); }
namespace { StrLen($x); }
`
	e, err := NewEngine(nil, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ctx := NewTestContext(e, f)
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			got = append(got, ctx.GlobalFunctionName(call))
		}
		return true
	})
	want := []string{"", "", "trim", "", "trim", "", "count", "", "strlen"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
	if !ctx.IsGlobalFunctionCall(firstCall(f, "StrLen"), "strlen") {
		t.Fatal("StrLen in the global namespace should resolve to strlen")
	}
}

func firstCall(f *syntax.File, written string) *syntax.FuncCall {
	var out *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok && out == nil {
			if nm, ok := call.Name.(*syntax.Name); ok && nm.Value == written {
				out = call
			}
		}
		return out == nil
	})
	return out
}
