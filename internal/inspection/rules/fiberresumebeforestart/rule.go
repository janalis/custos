// Package fiberresumebeforestart implements the native FiberResumeBeforeStart inspection.
package fiberresumebeforestart

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Start the fiber before resuming it."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberResumeBeforeStart" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || name.Value != "resume" {
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
	creation, ok := syntax.UnwrapParens(a.Value).(*syntax.New)
	if !ok || !semanticquery.NamesGlobalClass(ctx, creation.Class, "Fiber") {
		return
	}
	ctx.ReportNode(call, message)
}
