// Package arrayslicediscardsrequiredkeys implements the native ArraySliceDiscardsRequiredKeys inspection.
package arrayslicediscardsrequiredkeys

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve numeric keys before reading the original key."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArraySliceDiscardsRequiredKeys" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_slice") {
		return
	}
	preserve := semanticquery.CallArgument(call.Args, 3, "preserve_keys")
	if preserve != nil {
		yes, known := semanticquery.NativeTruth(ctx, preserve)
		if !known || yes {
			return
		}
	}
	a := semanticquery.NativeArray(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	offset, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "offset"))
	if a == nil || !ok {
		return
	}
	effective, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	if !known || len(effective) != len(a.Items) {
		return
	}
	for _, item := range a.Items {
		if item.Key == nil {
			return
		}
		if _, ok := semanticquery.NativeInt(ctx, item.Key); !ok {
			return
		}
	}
	size := int64(len(a.Items))
	if offset < 0 {
		offset += size
	}
	if offset < 0 {
		offset = 0
	}
	if offset > size {
		offset = size
	}
	end := size
	length := semanticquery.CallArgument(call.Args, 2, "length")
	if length != nil {
		n, known := semanticquery.NativeInt(ctx, length)
		if !known {
			return
		}
		if n < 0 {
			end = size + n
		} else if n < size-offset {
			end = offset + n
		}
	}
	if end < offset {
		end = offset
	}
	assignment, ok := call.Parent().(*syntax.Assign)
	if !ok || assignment.ByRef {
		return
	}
	stmt, ok := assignment.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	access, ok := echoExpr(ctx, stmt).(*syntax.ArrayDimFetch)
	if !ok || variable(access.Var) == "" || variable(access.Var) != variable(assignment.Var) {
		return
	}
	key, known := semanticquery.NativeInt(ctx, access.Dim)
	if !known || (key >= 0 && key < end-offset) {
		return
	}
	for _, item := range a.Items[offset:end] {
		original, _ := semanticquery.NativeInt(ctx, item.Key)
		if original != 0 && original == key {
			ctx.ReportNode(call, message)
			return
		}
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
