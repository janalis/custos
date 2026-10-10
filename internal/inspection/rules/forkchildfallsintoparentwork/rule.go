// Package forkchildfallsintoparentwork implements the native ForkChildFallsIntoParentWork inspection.
package forkchildfallsintoparentwork

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Terminate or separate the fork child branch."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ForkChildFallsIntoParentWork" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	branch := n.(*syntax.If)
	if branch.Else != nil || len(branch.ElseIfs) > 0 {
		return
	}
	cond, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
	if !ok || cond.Op.Kind != syntax.TIsIdentical {
		return
	}
	zero, known := semanticquery.NativeInt(ctx, cond.Right)
	if !known || zero != 0 {
		return
	}
	fork, ok := semanticquery.NativeValue(ctx, cond.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, fork, "pcntl_fork") {
		return
	}
	body, ok := branch.Body.(*syntax.Block)
	if !ok || len(body.Stmts) == 0 {
		return
	}
	last := body.Stmts[len(body.Stmts)-1]
	switch st := last.(type) {
	case *syntax.Return:
		return
	case *syntax.ExprStmt:
		switch st.Expr.(type) {
		case *syntax.Exit, *syntax.Throw:
			return
		}
	}
	if _, ok := astquery.NextStmt(ctx.File, branch); ok {
		ctx.ReportNode(cond, message)
	}
}
