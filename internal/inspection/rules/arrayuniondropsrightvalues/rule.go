// Package arrayuniondropsrightvalues implements the native ArrayUnionDropsRightValues inspection.
package arrayuniondropsrightvalues

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use array merging to preserve the right-hand list values."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayUnionDropsRightValues" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TPlus {
		return
	}
	a := semanticquery.NativeArray(ctx, b.Left)
	c := semanticquery.NativeArray(ctx, b.Right)
	if !implicit(a) || !implicit(c) {
		return
	}
	assignment, ok := b.Parent().(*syntax.Assign)
	if !ok || assignment.ByRef {
		return
	}
	local, ok := assignment.Var.(*syntax.Variable)
	if !ok || local.Name == "" {
		return
	}
	stmt, ok := assignment.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, stmt)
	if !ok {
		return
	}
	loop, ok := next.(*syntax.Foreach)
	if !ok || loop.Key != nil || loop.ByRef {
		return
	}
	v, ok := loop.Expr.(*syntax.Variable)
	if ok && v.Name == local.Name {
		ctx.ReportNode(b, message)
	}
}

func implicit(a *syntax.Array) bool {
	if a == nil || len(a.Items) == 0 {
		return false
	}
	for _, item := range a.Items {
		if item.Key != nil {
			return false
		}
	}
	return true
}
