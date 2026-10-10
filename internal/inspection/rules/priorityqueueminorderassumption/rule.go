// Package priorityqueueminorderassumption implements the native PriorityQueueMinOrderAssumption inspection.
package priorityqueueminorderassumption

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reverse priorities for smallest-first extraction."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PriorityQueueMinOrderAssumption" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || (!strings.EqualFold(id.Value, "extract") && !strings.EqualFold(id.Value, "top") && !strings.EqualFold(id.Value, "current")) || !semanticquery.NativeMethod(ctx, c, "SplPriorityQueue", id.Value) || semanticquery.NativeConstruction(ctx, c.Var, "SplPriorityQueue") == nil {
		return
	}
	var priorities []float64
	for _, n := range semanticquery.NativePriorCalls(ctx, c, c.Var, "insert", "extract") {
		p := n.(*syntax.MethodCall)
		if strings.EqualFold(p.Name.(*syntax.Identifier).Value, "extract") {
			return
		}
		v, known := number(ctx, semanticquery.CallArgument(p.Args, 1, "priority"))
		if !known || v < 0 {
			return
		}
		priorities = append(priorities, v)
	}
	if len(priorities) < 2 {
		return
	}
	for _, p := range priorities[1:] {
		if p != priorities[0] {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func number(ctx *analysis.Context, e syntax.Expr) (float64, bool) {
	if n, known := semanticquery.NativeInt(ctx, e); known {
		return float64(n), true
	}
	e = semanticquery.NativeValue(ctx, e)
	negative := false
	if u, ok := e.(*syntax.Unary); ok {
		if u.Op.Kind != syntax.TMinus && u.Op.Kind != syntax.TPlus {
			return 0, false
		}
		negative = u.Op.Kind == syntax.TMinus
		e = syntax.UnwrapParens(u.Expr)
	}
	lit, ok := e.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitFloat {
		return 0, false
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
	if negative {
		n = -n
	}
	return n, err == nil
}
