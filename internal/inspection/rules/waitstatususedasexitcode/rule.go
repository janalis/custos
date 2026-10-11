// Package waitstatususedasexitcode implements the native WaitStatusUsedAsExitCode inspection.
package waitstatususedasexitcode

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Decode the child wait status."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "WaitStatusUsedAsExitCode" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
	default:
		return
	}
	value, ok := semanticquery.NativeInt(ctx, b.Right)
	if !ok || value < 1 || value > 255 {
		return
	}
	wait := semanticquery.ExpansionCOutputCall(ctx, b, b.Left, []string{"pcntl_waitpid"}, 1, "status")
	if wait == nil {
		wait = semanticquery.ExpansionCOutputCall(ctx, b, b.Left, []string{"pcntl_wait"}, 0, "status")
	}
	if wait != nil && !rewritten(ctx, b, wait) {
		ctx.ReportNode(b, message)
	}
}

// rewritten reports whether the compared variable can hold something other
// than the raw wait status: it is rebound (assignment, destructuring, foreach
// target, ++/--, unset, global/static, catch), written through an element, or
// bound to a reference after the wait call. A write later in a loop that
// repeats the comparison without repeating the wait call counts as well.
func rewritten(ctx *analysis.Context, b *syntax.Binary, wait *syntax.FuncCall) bool {
	status := syntax.UnwrapParens(b.Left).(*syntax.Variable)
	scope := syntax.EnclosingVariableScope(b)
	from, to := wait.Span().End, b.Span().Start
	for p := b.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			if sp := p.Span(); sp.Start > from {
				to = sp.End
			}
		}
	}
	for _, access := range flowquery.VarAccesses(ctx.File, scope, status.Name) {
		if at := access.Var.Span().Start; at < from || at >= to {
			continue
		}
		if access.Write || access.ElemWrite {
			return true
		}
		if use, ok := access.By.(*syntax.ClosureUse); ok && use.ByRef {
			return true
		}
		parent, child := astquery.ParentSkipParens(access.Var)
		if a, ok := parent.(*syntax.Assign); ok && a.ByRef && a.Value == child {
			return true
		}
	}
	return false
}
