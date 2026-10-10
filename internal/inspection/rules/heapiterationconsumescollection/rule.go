// Package heapiterationconsumescollection implements the native HeapIterationConsumesCollection inspection.
package heapiterationconsumescollection

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

const message = "Iterate a clone to preserve the heap."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HeapIterationConsumesCollection" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	f := n.(*syntax.Foreach)
	if _, ok := syntax.UnwrapParens(f.Expr).(*syntax.Clone); ok {
		return
	}
	classes := ctx.Types().Native().TypeOf(f.Expr).Classes()
	if len(classes) != 1 {
		return
	}
	c := ctx.Index().Class(classes[0], ctx.PHP)
	if c == nil || !strings.HasPrefix(c.File, "stubs/") || (!ctx.Index().IsSubtype(c.FQN, "SplHeap", ctx.PHP) && !ctx.Index().IsSubtype(c.FQN, "SplPriorityQueue", ctx.PHP)) {
		return
	}
	ctx.ReportNode(f.Expr, message)
}
