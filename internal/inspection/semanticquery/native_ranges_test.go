package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeIntegerRanges(t *testing.T) {
	probeNative(t, `if($flag){$n=-2;}else{$n=-2;}probe($n);`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		if ctx.Text(c.Name) != "probe" {
			return
		}
		x := CallArgument(c.Args, 0, "value")
		lo, hi, known := NativeIntRange(ctx, x)
		if !known || lo != -2 || hi != -2 {
			t.Fatalf("singleton range [%d,%d],%v", lo, hi, known)
		}
		if _, exact := NativeInt(ctx, x); exact {
			t.Fatal("equal values from different expressions must not change the exact-expression contract")
		}
	})
	for _, src := range []string{`probe(-2);`, `const LIMIT=-2;probe(LIMIT);`, `class C{const LIMIT=-2;}probe(C::LIMIT);`} {
		probeNative(t, src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if ctx.Text(c.Name) != "probe" {
				return
			}
			lo, hi, known := NativeIntRange(ctx, CallArgument(c.Args, 0, "value"))
			if !known || lo != -2 || hi != -2 {
				t.Fatalf("range [%d,%d],%v", lo, hi, known)
			}
		})
	}
	probeNative(t, `function f(&$n,&$m){$n=-2;$m=5;probe($n);}`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		if ctx.Text(c.Name) != "probe" {
			return
		}
		if _, _, known := NativeIntRange(ctx, CallArgument(c.Args, 0, "value")); known {
			t.Fatal("possibly aliased reference parameters must remain unknown")
		}
	})
}

func TestNativeIntegerRangeContractFallback(t *testing.T) {
	probeNative(t, `probe(128|1);`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		if ctx.Text(c.Name) != "probe" {
			return
		}
		lo, hi, known := NativeIntRange(ctx, CallArgument(c.Args, 0, "value"))
		if !known || lo != 129 || hi != 129 {
			t.Fatalf("contract fallback=[%d,%d],%v", lo, hi, known)
		}
	})
}
