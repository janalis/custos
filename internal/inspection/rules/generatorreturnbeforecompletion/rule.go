// Package generatorreturnbeforecompletion implements the native GeneratorReturnBeforeCompletion inspection.
package generatorreturnbeforecompletion

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Finish the generator before reading its return value."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GeneratorReturnBeforeCompletion" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP70 {
		return
	}
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || name.Value != "getReturn" {
		return
	}
	v, ok := call.Var.(*syntax.Variable)
	if !ok || v.Name == "" {
		return
	}
	st, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	prev, ok := astquery.PrevStmtNoDoc(ctx.File, st)
	if !ok {
		return
	}
	assignment, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	a, ok := assignment.Expr.(*syntax.Assign)
	if !ok || a.ByRef {
		return
	}
	local, ok := a.Var.(*syntax.Variable)
	if !ok || local.Name != v.Name {
		return
	}
	closureCall, ok := syntax.UnwrapParens(a.Value).(*syntax.FuncCall)
	if !ok {
		return
	}
	closure, ok := syntax.UnwrapParens(closureCall.Name).(*syntax.Closure)
	if !ok || closure.Body == nil || len(closure.Body.Stmts) == 0 {
		return
	}
	first, ok := closure.Body.Stmts[0].(*syntax.ExprStmt)
	if !ok {
		return
	}
	if _, ok := first.Expr.(*syntax.Yield); !ok {
		return
	}
	ctx.ReportNode(call, message)
}
