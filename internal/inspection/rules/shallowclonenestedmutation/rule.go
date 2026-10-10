// Package shallowclonenestedmutation implements the ShallowCloneNestedMutation inspection.
package shallowclonenestedmutation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ShallowCloneNestedMutation" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

const message = "Clone the nested object before mutating independent state."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	assign := n.(*syntax.Assign)
	outer, ok := assign.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	child, ok := outer.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	name, fixedName := child.Name.(*syntax.Identifier)
	if !fixedName {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, n)
	if len(prior) < 2 {
		return
	}
	cloneStatement, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	cloneAssign, ok := cloneStatement.Expr.(*syntax.Assign)
	if !ok || !astquery.Equivalent(ctx.File, cloneAssign.Var, child.Var) {
		return
	}
	value, ok := syntax.UnwrapParens(cloneAssign.Value).(*syntax.Clone)
	if !ok || value.With != nil {
		return
	}
	classes := ctx.TypeOf(value.Expr).Classes()
	if len(classes) != 1 {
		return
	}
	for _, cls := range classes {
		cls = strings.TrimPrefix(cls, `\`)
		if ctx.Index().Class(cls, ctx.PHP) == nil || !ctx.Index().AncestorsComplete(cls, ctx.PHP) || !semanticquery.HierarchyResolved(ctx.Index(), cls, ctx.PHP) || ctx.Index().FindMethod(cls, "__clone", ctx.PHP) != nil || ctx.Index().FindMethod(cls, "__get", ctx.PHP) != nil || ctx.Index().FindMethod(cls, "__set", ctx.PHP) != nil {
			return
		}
		if property := ctx.Index().FindProperty(cls, name.Value, ctx.PHP); property != nil && (property.ReadsRunCode || property.WritesRunCode || property.Magic) {
			return
		}
	}

	statement, ok := prior[len(prior)-2].(*syntax.ExprStmt)
	if !ok {
		return
	}
	previous, ok := statement.Expr.(*syntax.Assign)
	if !ok {
		return
	}
	property, ok := previous.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	if !astquery.Equivalent(ctx.File, property.Var, value.Expr) || !astquery.Equivalent(ctx.File, property.Name, child.Name) {
		return
	}
	if _, ok := semanticquery.NativeValue(ctx, previous.Value).(*syntax.New); !ok {
		return
	}
	ctx.ReportNode(outer, message)
}

func (rule) Semantic() {}
