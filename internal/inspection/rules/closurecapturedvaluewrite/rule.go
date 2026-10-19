// Package closurecapturedvaluewrite implements the native ClosureCapturedValueWrite inspection.
package closurecapturedvaluewrite

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Capture the scalar by reference when updating outer state."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ClosureCapturedValueWrite" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KClosure} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.Closure)
	a, ok := cl.Parent().(*syntax.Assign)
	if !ok {
		return
	}
	v, ok := a.Var.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return
	}
	st, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	invoke, ok := astquery.NextStmt(ctx.File, st)
	if !ok {
		return
	}
	is, ok := invoke.(*syntax.ExprStmt)
	if !ok {
		return
	}
	call, ok := is.Expr.(*syntax.FuncCall)
	if !ok {
		return
	}
	callee, ok := call.Name.(*syntax.Variable)
	if !ok || callee.Name != v.Name {
		return
	}
	after, ok := astquery.NextStmt(ctx.File, invoke)
	if !ok {
		return
	}
	echo, ok := after.(*syntax.Echo)
	if !ok {
		return
	}
	for _, use := range cl.Uses {
		if use.ByRef || use.Var == nil || !ctx.TypeOf(use.Var).OnlyOf("int", "float", "string", "bool", "true", "false") {
			continue
		}
		name := use.Var.Name
		write, returns, read := false, false, false
		syntax.Inspect(cl.Body, func(x syntax.Node) bool {
			if !semanticquery.NativeCallbackReachable(ctx, x) {
				return false
			}
			switch z := x.(type) {
			case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
				// A nested scope that mentions the name consumes the copy.
				read = read || astquery.MentionsVariable(z, name)
				return false
			case *syntax.Return:
				returns = returns || z.Expr != nil
			case *syntax.Assign:
				if writesName(z.Var, name) {
					write = true
					read = read || !discarded(z)
					return false
				}
			case *syntax.IncDec:
				if writesName(z.Var, name) {
					write = true
					read = read || !discarded(z)
					return false
				}
			case *syntax.Variable:
				read = read || z.Name == name
			}
			return true
		})
		if write && !returns && !read && astquery.MentionsVariable(echo, name) {
			ctx.ReportNode(cl, message)
			return
		}
	}
}

// writesName reports whether target is the plain variable $name.
func writesName(target syntax.Expr, name string) bool {
	v, ok := target.(*syntax.Variable)
	return ok && v.Name == name
}

// discarded reports whether the value of write is unused: the write is a
// whole expression statement, so the closure does not consume its result.
func discarded(write syntax.Node) bool {
	parent, _ := astquery.ParentSkipParens(write)
	_, ok := parent.(*syntax.ExprStmt)
	return ok
}
