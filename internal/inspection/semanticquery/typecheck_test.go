package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const typecheckSrc = `declare(strict_types=1);
namespace N;
interface I {}
class A implements I {}
final class F {}
enum E { case X; }
class P extends Missing\Base {}
probe(new A());`

func TestIndexRelations(t *testing.T) {
	ran := false
	probeNative(t, typecheckSrc, func(ctx *analysis.Context, call *syntax.FuncCall) {
		ran = true
		r := NewIndexRelations(ctx.Index(), ctx.PHP)
		for c, want := range map[string]bool{`\N\A`: true, `\N\I`: true, `\N\P`: false, `\N\Nope`: false, `\Countable`: true} {
			first := r.Resolved(c)
			if cached := r.Resolved(c); first != want || cached != want {
				t.Errorf("Resolved(%s) = %v then %v, want %v", c, first, cached, want)
			}
		}
		for _, tc := range []struct {
			child, parent string
			want          types.Verdict
		}{
			{`\N\A`, `\N\I`, types.Yes},
			{`\N\A`, `\n\a`, types.Yes},
			{`\N\F`, `\N\I`, types.No},
			{`\N\P`, `\N\I`, types.Maybe},
			{`\N\A`, `\N\Nope`, types.Maybe},
		} {
			if got := r.IsSubtype(tc.child, tc.parent); got != tc.want {
				t.Errorf("IsSubtype(%s, %s) = %s", tc.child, tc.parent, got)
			}
		}
		if !r.IsFinal(`\N\F`) || !r.IsFinal(`\N\E`) || r.IsFinal(`\N\A`) || r.IsFinal(`\N\Nope`) {
			t.Error("IsFinal")
		}
		if !r.IsInterface(`\N\I`) || r.IsInterface(`\N\A`) || r.IsInterface(`\N\Nope`) {
			t.Error("IsInterface")
		}
		ac := AcceptContext(ctx, true)
		if !ac.Strict || !ac.Internal || ac.Classes == nil {
			t.Errorf("AcceptContext: %+v", ac)
		}
		arg := ctx.Types().Native().TypeOf(CallArgument(call.Args, 0, ""))
		if got := types.Accepts(types.Of(`\N\I`), arg, ac); got != types.Yes {
			t.Errorf("Accepts(I, %s) = %s", arg, got)
		}
		if got := types.Accepts(types.Of(`\N\F`), arg, ac); got != types.No {
			t.Errorf("Accepts(F, %s) = %s", arg, got)
		}
	})
	if !ran {
		t.Fatal("probe not reached")
	}
}

func BenchmarkAcceptsIndex(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, call *syntax.FuncCall) {
		ac := AcceptContext(ctx, false)
		expected, actual := types.Of(`\N\I`, "null"), types.Of(`\N\F`)
		b.ReportAllocs()
		for b.Loop() {
			types.Accepts(expected, actual, ac)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("typecheck.php", []byte("<?php "+typecheckSrc), syntax.Options{}))
}
