// Package sessionregenerationfailureunchecked implements the native SessionRegenerationFailureUnchecked inspection.
package sessionregenerationfailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check session identifier regeneration before authenticating."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionRegenerationFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "session_regenerate_id") {
		return
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, stmt)
	if !ok {
		return
	}
	st, ok := next.(*syntax.ExprStmt)
	if !ok {
		return
	}
	a, ok := st.Expr.(*syntax.Assign)
	if !ok || a.ByRef || a.Op.Kind != syntax.TEqual {
		return
	}
	truth, known := semanticquery.NativeTruth(ctx, a.Value)
	if !known || !truth {
		return
	}
	dim, ok := a.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	v, ok := dim.Var.(*syntax.Variable)
	if !ok || v.Name != "_SESSION" {
		return
	}
	key, known := semanticquery.NativeString(ctx, dim.Dim)
	if !known {
		return
	}
	for _, name := range ctx.List("authenticationKeys") {
		if key == name {
			ctx.ReportNode(call, message)
			return
		}
	}
}
