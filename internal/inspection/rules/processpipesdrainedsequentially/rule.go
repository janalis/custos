// Package processpipesdrainedsequentially implements the native ProcessPipesDrainedSequentially inspection.
package processpipesdrainedsequentially

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Drain process output pipes concurrently."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ProcessPipesDrainedSequentially" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "stream_get_contents") {
		return
	}
	stderr, ok := semanticquery.CallArgument(c.Args, 0, "stream").(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	number, known := semanticquery.NativeInt(ctx, stderr.Dim)
	array, ok := stderr.Var.(*syntax.Variable)
	if !known || number != 2 || !ok {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	read, ok := callStatement(prior[len(prior)-1])
	if !ok || !semanticquery.NativeBuiltin(ctx, read, "stream_get_contents") {
		return
	}
	stdout, ok := semanticquery.CallArgument(read.Args, 0, "stream").(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	index, known := semanticquery.NativeInt(ctx, stdout.Dim)
	base, ok := stdout.Var.(*syntax.Variable)
	if !known || index != 1 || !ok || base.Name != array.Name {
		return
	}
	open, ok := callStatement(prior[len(prior)-2])
	if !ok || !semanticquery.NativeBuiltin(ctx, open, "proc_open") {
		return
	}
	pipes, ok := semanticquery.CallArgument(open.Args, 2, "pipes").(*syntax.Variable)
	if !ok || pipes.Name != array.Name {
		return
	}
	descriptors, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(open.Args, 1, "descriptor_spec"))
	if !known {
		return
	}
	for _, key := range []string{"i:1", "i:2"} {
		tuple, known := semanticquery.NativeArrayEntries(ctx, descriptors[key])
		if !known {
			return
		}
		kind, kk := semanticquery.NativeString(ctx, tuple["i:0"])
		mode, mk := semanticquery.NativeString(ctx, tuple["i:1"])
		if !kk || !mk || kind != "pipe" || mode != "w" {
			return
		}
	}
	ctx.ReportNode(c, message)
}

func callStatement(st syntax.Stmt) (*syntax.FuncCall, bool) {
	statement, ok := st.(*syntax.ExprStmt)
	if !ok {
		return nil, false
	}
	expr := statement.Expr
	if a, ok := expr.(*syntax.Assign); ok && !a.ByRef && a.Op.Kind == syntax.TEqual {
		expr = a.Value
	}
	call, ok := expr.(*syntax.FuncCall)
	return call, ok
}
