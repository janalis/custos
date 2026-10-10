// Package linkedlistdeleteiterationunexpected implements the native LinkedListDeleteIterationUnexpected inspection.
package linkedlistdeleteiterationunexpected

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve linked-list elements during iteration."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "LinkedListDeleteIterationUnexpected" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	f := n.(*syntax.Foreach)
	if _, ok := syntax.UnwrapParens(f.Expr).(*syntax.Clone); ok {
		return
	}
	classes := ctx.Types().Native().TypeOf(f.Expr).Classes()
	if len(classes) != 1 {
		return
	}
	cl := ctx.Index().Class(classes[0], ctx.PHP)
	if cl == nil || !strings.HasPrefix(cl.File, "stubs/") || !ctx.Index().IsSubtype(cl.FQN, "SplDoublyLinkedList", ctx.PHP) {
		return
	}
	mode := int64(0)
	for _, n := range semanticquery.NativePriorCalls(ctx, f, f.Expr, "setIteratorMode") {
		c := n.(*syntax.MethodCall)
		known := false
		mode, known = semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 0, "mode"))
		if !known {
			return
		}
	}
	if mode&1 != 0 {
		ctx.ReportNode(f.Expr, message)
	}
}
