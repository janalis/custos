package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestStringCallable(t *testing.T) {
	src := `<?php
namespace Shop { use Vendor as Lib; use function Tools\sum; function max() {} $x = 1; }
namespace { $y = 1; }
`
	e, err := analysis.NewEngine(nil, analysis.Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ctx := analysis.NewTestContext(e, f)
	var vars []uint32
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok {
			vars = append(vars, v.Span().Start)
		}
		return true
	})
	ns, global := vars[0], vars[1]
	fn := []struct {
		in   string
		at   uint32
		want string
	}{
		{"max", ns, `\max`},
		{"sum", ns, `\sum`},
		{"min", ns, "min"},
		{"Lib\\fmt", ns, `\Lib\fmt`},
		{"Util\\fmt", ns, `\Util\fmt`},
		{"\\Util\\fmt", ns, `\Util\fmt`},
		{"Util\\fmt", global, `Util\fmt`},
	}
	for _, c := range fn {
		if got := StringCallableFunction(ctx, c.in, c.at); got != c.want {
			t.Errorf("function %q: got %q, want %q", c.in, got, c.want)
		}
	}
	cls := []struct {
		in   string
		at   uint32
		want string
	}{
		{"Repo", ns, `\Repo`}, {"Lib", ns, `\Lib`}, {"self", ns, "self"}, {"Repo", global, "Repo"}, {"\\A\\B", ns, `\A\B`},
	}
	for _, c := range cls {
		if got := StringCallableClass(ctx, c.in, c.at); got != c.want {
			t.Errorf("class %q: got %q, want %q", c.in, got, c.want)
		}
	}
}
