package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeAuditSharedProofBoundaries(t *testing.T) {
	probeNative(t, "$x='same';probe($x,$x);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if !SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("uniquely reaching assigned value lost")
		}
		if SameOperand(ctx, nil, CallArgument(c.Args, 0, "")) {
			t.Fatal("missing path must be unknown")
		}
	})
	probeNative(t, "probe($unknown,$unknown);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if SameOperand(ctx, CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")) {
			t.Fatal("unknown variables cannot prove identity")
		}
	})
	probeNative(t, "probe(mb_strlen($body,chooseEncoding()));", func(_ *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "").(*syntax.FuncCall)
		if ArgumentsRemovable(call) {
			t.Fatal("effectful optional argument must remain evaluated")
		}
	})
}
