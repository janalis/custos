// Package gmpdivisionbyknownzero implements the native GmpDivisionByKnownZero inspection.
package gmpdivisionbyknownzero

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a nonzero GMP divisor."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GmpDivisionByKnownZero" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	switch name {
	case "gmp_div_q", "gmp_div_r", "gmp_div_qr", "gmp_mod":
	default:
		return
	}
	e := semanticquery.CallArgument(c.Args, 1, "num2")
	if zero(ctx, e) {
		ctx.ReportNode(c, message)
	}
}

func zero(ctx *analysis.Context, e syntax.Expr) bool {
	if n, ok := semanticquery.NativeInt(ctx, e); ok {
		return n == 0
	}
	if s, ok := semanticquery.NativeString(ctx, e); ok {
		if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
			s = s[1:]
		}
		return s != "" && strings.Trim(s, "0") == ""
	}
	if call, ok := semanticquery.NativeValue(ctx, e).(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, call, "gmp_init") {
		base := semanticquery.CallArgument(call.Args, 1, "base")
		if base != nil {
			v, ok := semanticquery.NativeInt(ctx, base)
			if !ok || (v != 0 && (v < 2 || v > 62)) {
				return false
			}
		}
		arg := semanticquery.CallArgument(call.Args, 0, "num")
		if _, nested := semanticquery.NativeValue(ctx, arg).(*syntax.FuncCall); nested {
			return false
		}
		return zero(ctx, arg)
	}
	return false
}
