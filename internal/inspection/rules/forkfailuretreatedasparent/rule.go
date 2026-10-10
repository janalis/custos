// Package forkfailuretreatedasparent implements the native ForkFailureTreatedAsParent inspection.
package forkfailuretreatedasparent

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle fork failure separately."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ForkFailureTreatedAsParent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	branch := n.(*syntax.If)
	input := syntax.UnwrapParens(branch.Cond)
	if u, ok := input.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		input = u.Expr
	}
	c, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if previous, exists := astquery.PrevStmt(ctx.File, branch); exists {
		if guard, ok := previous.(*syntax.If); ok && guard.Else == nil && syntax.Terminates(guard.Body) {
			if b, ok := syntax.UnwrapParens(guard.Cond).(*syntax.Binary); ok && b.Op.Kind == syntax.TIsIdentical && semanticquery.ExpansionCSame(ctx, b.Left, input) {
				if v, known := semanticquery.NativeInt(ctx, b.Right); known && v == -1 {
					return
				}
			}
		}
	}

	if ok && semanticquery.NativeBuiltin(ctx, c, "pcntl_fork") {
		ctx.ReportNode(branch.Cond, message)
	}
}
