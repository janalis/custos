// Package fractionalarraykeytruncation implements the native FractionalArrayKeyTruncation inspection.
package fractionalarraykeytruncation

import (
	"math"
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

const message = "Use a string key to preserve the fractional value."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FractionalArrayKeyTruncation" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayItem} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	item := n.(*syntax.ArrayItem)
	key := syntax.UnwrapParens(item.Key)
	if u, ok := key.(*syntax.Unary); ok {
		if u.Op.Kind != syntax.TMinus && u.Op.Kind != syntax.TPlus {
			return
		}
		key = syntax.UnwrapParens(u.Expr)
	}
	lit, ok := key.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitFloat {
		return
	}
	f, err := strconv.ParseFloat(lit.Raw, 64)
	if err == nil && !math.IsInf(f, 0) && !math.IsNaN(f) && math.Trunc(f) != f {
		ctx.ReportNode(item.Key, message)
	}
}
