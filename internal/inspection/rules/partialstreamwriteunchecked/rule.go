// Package partialstreamwriteunchecked implements the native PartialStreamWriteUnchecked inspection.
package partialstreamwriteunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Verify that every payload byte was written."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PartialStreamWriteUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "fwrite") {
		return
	}
	if semanticquery.NativeSuccessFollower(ctx, call) {
		ctx.ReportNode(call, message)
		return
	}
	p, _ := astquery.ParentSkipParens(call)
	b, ok := p.(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TIsIdentical || syntax.UnwrapParens(b.Left) != call {
		return
	}
	truth, known := astquery.BoolConst(b.Right)
	if !known || truth {
		return
	}
	parent, _ := astquery.ParentSkipParens(b)
	condition, ok := parent.(*syntax.If)
	if !ok || condition.Else != nil || len(condition.ElseIfs) != 0 {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, condition)
	if !ok {
		return
	}
	ret, ok := next.(*syntax.Return)
	if !ok {
		return
	}
	truth, known = astquery.BoolConst(ret.Expr)
	if known && truth {
		ctx.ReportNode(call, message)
	}
}
