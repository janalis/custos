// Package fractionalmodulooperand implements the native FractionalModuloOperand inspection.
package fractionalmodulooperand

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use fmod for fractional operands."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FractionalModuloOperand" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TMod {
		return
	}
	for _, e := range []syntax.Expr{b.Left, b.Right} {
		v := semanticquery.NativeValue(ctx, e)
		if u, ok := v.(*syntax.Unary); ok && (u.Op.Kind == syntax.TMinus || u.Op.Kind == syntax.TPlus) {
			v = syntax.UnwrapParens(u.Expr)
		}
		if l, ok := v.(*syntax.Literal); ok && l.LitKind == syntax.LitFloat {
			f, err := strconv.ParseFloat(strings.ReplaceAll(l.Raw, "_", ""), 64)
			if err == nil && !math.IsInf(f, 0) && !math.IsNaN(f) && math.Trunc(f) != f {
				ctx.ReportNode(b, message)
				return
			}
		}
	}
}
