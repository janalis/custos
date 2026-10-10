package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// ExpansionBcOperands returns the numeric operand slots of a resolved BCMath
// call. Scale and exponent parameters are not decimal-value operands.
func ExpansionBcOperands(ctx *analysis.Context, c *syntax.FuncCall) []syntax.Expr {
	switch NativeBuiltinName(ctx, c) {
	case "bcadd", "bcsub", "bcmul", "bcdiv", "bcmod", "bccomp":
		return []syntax.Expr{CallArgument(c.Args, 0, "num1"), CallArgument(c.Args, 1, "num2")}
	case "bcpowmod":
		return []syntax.Expr{CallArgument(c.Args, 0, "num"), CallArgument(c.Args, 2, "modulus")}
	case "bcpow", "bcsqrt":
		return []syntax.Expr{CallArgument(c.Args, 0, "num")}
	}
	return nil
}
