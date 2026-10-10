// Package readonlypropertyreassignment implements the ReadonlyPropertyReassignment inspection.
package readonlypropertyreassignment

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReadonlyPropertyReassignment" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

const message = "Initialize this readonly property only once."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	assign := n.(*syntax.Assign)
	fetch, ok := assign.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	name, ok := fetch.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	classes := ctx.TypeOf(fetch.Var).Classes()
	if len(classes) != 1 {
		return
	}
	property := ctx.Index().FindProperty(strings.TrimPrefix(classes[0], `\`), name.Value, ctx.PHP)
	if property == nil || !property.Readonly {
		return
	}
	method, ok := syntax.EnclosingFuncLike(assign).(*syntax.Method)
	if ok && strings.EqualFold(method.Name.Value, "__clone") && ctx.PHP >= phpversion.PHP83 {
		return
	}
	statement, ok := assign.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	_, ok = statement.Parent().(*syntax.Block)
	if !ok {
		return
	}
	priorStatements := flowquery.NativePriorStatements(ctx.File, assign)
	if len(priorStatements) == 0 {
		return
	}
	expression, ok := priorStatements[len(priorStatements)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	prior, ok := expression.Expr.(*syntax.Assign)
	if !ok || prior.ByRef {
		return
	}
	if _, ok := syntax.UnwrapParens(fetch.Var).(*syntax.Variable); !ok {
		return
	}
	// Unknown evaluation can replace the receiver or unset its property.
	switch syntax.UnwrapParens(prior.Value).(type) {
	case *syntax.Literal, *syntax.Variable:
	default:
		return
	}
	if astquery.Equivalent(ctx.File, prior.Var, assign.Var) {
		ctx.ReportNode(fetch, message)
	}
}

func (rule) Semantic() {}
