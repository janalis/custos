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
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "session_destroy") {
		return
	}
	if before, ok := astquery.PrevStmt(ctx.File, st); ok && clearsLocal(ctx, before, dim) {
		return
	}
	ctx.ReportNode(dim, message)
}

// clearsLocal reports whether s, the statement directly before the destroy,
// already removed the local value read: the whole array was emptied or
// unset, or the same literal key was unset.
func clearsLocal(ctx *analysis.Context, s syntax.Stmt, read *syntax.ArrayDimFetch) bool {
	switch s := s.(type) {
	case *syntax.Unset:
		for _, target := range s.Vars {
			if isSession(target) {
				return true
			}
			if d, ok := target.(*syntax.ArrayDimFetch); ok && isSession(d.Var) && sameLiteralKey(d.Dim, read.Dim) {
				return true
			}
		}
	case *syntax.ExprStmt:
		switch e := s.Expr.(type) {
		case *syntax.FuncCall:
			return semanticquery.NativeBuiltin(ctx, e, "session_unset")
		case *syntax.Assign:
			arr, ok := e.Value.(*syntax.Array)
			return ok && len(arr.Items) == 0 && !e.ByRef && e.Op.Kind == syntax.TEqual && isSession(e.Var)
		}
	}
	return false
}

func isSession(e syntax.Expr) bool {
	v, ok := e.(*syntax.Variable)
	return ok && v.Name == "_SESSION"
}

// sameLiteralKey compares two literal array keys; string literals compare by
// value so the quote style does not matter.
func sameLiteralKey(a, b syntax.Expr) bool {
	la, ok := a.(*syntax.Literal)
	if !ok {
		return false
	}
	lb, ok := b.(*syntax.Literal)
	if !ok {
		return false
	}
	va, sa := astquery.StringLiteralValue(la.Raw)
	vb, sb := astquery.StringLiteralValue(lb.Raw)
	if sa || sb {
		return sa && sb && va == vb
	}
	return la.Raw == lb.Raw
}
