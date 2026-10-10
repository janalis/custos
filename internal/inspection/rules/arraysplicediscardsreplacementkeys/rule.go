// Package arraysplicediscardsreplacementkeys implements the native ArraySpliceDiscardsReplacementKeys inspection.
package arraysplicediscardsreplacementkeys

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve replacement keys before reading the inserted key."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArraySpliceDiscardsReplacementKeys" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_splice") {
		return
	}
	local := semanticquery.CallArgument(call.Args, 0, "array")
	if variable(local) == "" {
		return
	}
	original, known := semanticquery.NativeArrayEntries(ctx, local)
	replacement, ok := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 3, "replacement"))
	if !known || !ok {
		return
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	access, ok := echoExpr(ctx, stmt).(*syntax.ArrayDimFetch)
	if !ok || variable(access.Var) != variable(local) {
		return
	}
	key, known := semanticquery.NativeArrayKey(ctx, access.Dim)
	if !known || len(key) < 2 || key[:2] != "s:" {
		return
	}
	if _, exists := original[key]; exists {
		return
	}
	if _, exists := replacement[key]; exists {
		ctx.ReportNode(call, message)
	}
}

func echoExpr(ctx *analysis.Context, s syntax.Stmt) syntax.Expr {
	next, ok := astquery.NextStmt(ctx.File, s)
	if !ok {
		return nil
	}
	echo, ok := next.(*syntax.Echo)
	if !ok || len(echo.Exprs) != 1 {
		return nil
	}
	return syntax.UnwrapParens(echo.Exprs[0])
}

func variable(e syntax.Expr) string {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return ""
	}
	return v.Name
}
