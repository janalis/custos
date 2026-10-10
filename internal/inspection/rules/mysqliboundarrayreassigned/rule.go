// Package mysqliboundarrayreassigned implements the native MysqliBoundArrayReassigned inspection.
package mysqliboundarrayreassigned

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Update the bound array element without replacing its array."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliBoundArrayReassigned" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli_stmt", "execute") || c.Args != nil && len(c.Args.Args) > 0 {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	assignmentStmt, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	assignment, ok := assignmentStmt.Expr.(*syntax.Assign)
	if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return
	}
	array, ok := assignment.Var.(*syntax.Variable)
	if !ok || semanticquery.NativeArray(ctx, assignment.Value) == nil {
		return
	}
	bindStmt, ok := prior[len(prior)-2].(*syntax.ExprStmt)
	if !ok {
		return
	}
	bind, ok := bindStmt.Expr.(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, bind, "mysqli_stmt", "bind_param") {
		return
	}
	for i, node := range bind.Args.Args {
		if i == 0 {
			continue
		}
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return
		}
		if semanticquery.NativeValue(ctx, c.Var) == nil || semanticquery.NativeValue(ctx, c.Var) != semanticquery.NativeValue(ctx, bind.Var) {
			return
		}
		dim, ok := arg.Value.(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		base, ok := dim.Var.(*syntax.Variable)
		if !ok || base.Name != array.Name {
			continue
		}
		if _, known := semanticquery.NativeArrayKey(ctx, dim.Dim); known && semanticquery.NativeArray(ctx, base) != nil {
			ctx.ReportNode(assignment, message)
			return
		}
	}
}
