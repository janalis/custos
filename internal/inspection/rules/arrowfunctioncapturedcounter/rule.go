// Package arrowfunctioncapturedcounter implements the native ArrowFunctionCapturedCounter inspection.
package arrowfunctioncapturedcounter

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use persistent shared state for this counter."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrowFunctionCapturedCounter" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrowFunction} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	arrow := n.(*syntax.ArrowFunction)
	if ctx.PHP < phpversion.PHP74 {
		return
	}
	inc, ok := syntax.UnwrapParens(arrow.Expr).(*syntax.IncDec)
	if !ok {
		return
	}
	counter, ok := inc.Var.(*syntax.Variable)
	if !ok || counter.NameExpr != nil {
		return
	}
	for _, p := range arrow.Params {
		if p.Var != nil && p.Var.Name == counter.Name {
			return
		}
	}
	if !ctx.TypeOf(counter).OnlyOf("int", "float") {
		return
	}
	a, ok := arrow.Parent().(*syntax.Assign)
	if !ok {
		return
	}
	v, ok := a.Var.(*syntax.Variable)
	if !ok {
		return
	}
	st, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	after, ok := astquery.NextStmt(ctx.File, st)
	if !ok {
		return
	}
	echo, ok := after.(*syntax.Echo)
	if !ok {
		return
	}
	count := 0
	for _, expr := range echo.Exprs {
		fc, ok := expr.(*syntax.FuncCall)
		if !ok {
			continue
		}
		name, ok := fc.Name.(*syntax.Variable)
		if ok && name.Name == v.Name {
			count++
		}
	}
	if count >= 2 {
		ctx.ReportNode(arrow, message)
	}
}
