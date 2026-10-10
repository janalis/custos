// Package sessiondestroyleaveslocalauthentication implements the native SessionDestroyLeavesLocalAuthentication inspection.
package sessiondestroyleaveslocalauthentication

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clear local session data after destroying the session."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionDestroyLeavesLocalAuthentication" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	dim := n.(*syntax.ArrayDimFetch)
	v, ok := dim.Var.(*syntax.Variable)
	if !ok || v.Name != "_SESSION" {
		return
	}
	branch, ok := dim.Parent().(*syntax.If)
	if !ok || branch.Cond != dim {
		return
	}
	state, known := ctx.Flow().GlobalStateBefore(dim, "session")
	if !known || state.Operation != "session_destroy" {
		return
	}
	prev, ok := astquery.PrevStmt(ctx.File, branch)
	if !ok {
		return
	}
	st, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	call, ok := st.Expr.(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltin(ctx, call, "session_destroy") {
		ctx.ReportNode(dim, message)
	}
}
