// Package execoutputarrayaccumulates implements the native ExecOutputArrayAccumulates inspection.
package execoutputarrayaccumulates

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reset the output array before collecting a separate command result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExecOutputArrayAccumulates" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "exec") {
		return
	}
	output, ok := semanticquery.CallArgument(c.Args, 1, "output").(*syntax.Variable)
	if !ok {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	previous, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	first, ok := previous.Expr.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, first, "exec") {
		return
	}
	firstOutput, ok := semanticquery.CallArgument(first.Args, 1, "output").(*syntax.Variable)
	if !ok || firstOutput.Name != output.Name {
		return
	}
	initial, ok := prior[len(prior)-2].(*syntax.ExprStmt)
	if !ok {
		return
	}
	assignment, ok := initial.Expr.(*syntax.Assign)
	if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok || variable.Name != output.Name {
		return
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, assignment.Value)
	if !known || len(entries) != 0 {
		return
	}
	st, ok := c.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, known := astquery.NextStmt(ctx.File, st)
	if !known {
		return
	}
	if each, ok := next.(*syntax.Foreach); ok {
		v, ok := each.Expr.(*syntax.Variable)
		if ok && v.Name == output.Name {
			ctx.ReportNode(c, message)
		}
		return
	}
	echo, ok := next.(*syntax.Echo)
	if !ok {
		return
	}
	for _, expr := range echo.Exprs {
		count, ok := expr.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, count, "count") {
			continue
		}
		v, ok := semanticquery.CallArgument(count.Args, 0, "value").(*syntax.Variable)
		if ok && v.Name == output.Name {
			ctx.ReportNode(c, message)
			return
		}
	}
}
