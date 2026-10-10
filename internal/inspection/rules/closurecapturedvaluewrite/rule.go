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
		write, returns := false, false
		syntax.Inspect(cl.Body, func(x syntax.Node) bool {
			if !semanticquery.NativeCallbackReachable(x) {
				return false
			}
			switch z := x.(type) {
			case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
				return false
			case *syntax.Return:
				returns = returns || z.Expr != nil
			case *syntax.Assign:
				if target, ok := z.Var.(*syntax.Variable); ok && target.Name == name {
					write = true
				}
			case *syntax.IncDec:
				if target, ok := z.Var.(*syntax.Variable); ok && target.Name == name {
					write = true
				}
			}
			return true
		})
		if write && !returns && astquery.MentionsVariable(echo, name) {
			ctx.ReportNode(cl, message)
			return
		}
	}
}
