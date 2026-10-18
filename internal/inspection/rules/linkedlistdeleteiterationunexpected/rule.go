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
	history := semanticquery.NativeCallHistory(ctx, f, f.Expr, "setIteratorMode")
	if !history.Complete {
		return
	}
	modeKnown := true
	for _, event := range history.Events {
		c := event.Node.(*syntax.MethodCall)
		value, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 0, "mode"))
		if event.Must {
			mode, modeKnown = value, known
		} else if !known || !modeKnown || value != mode {
			modeKnown = false
		}
	}
	if modeKnown && mode&1 != 0 {
		ctx.ReportNode(f.Expr, message)
	}
}
