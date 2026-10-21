// Package catchvariableoverwriteslocal implements the native CatchVariableOverwritesLocal inspection.
package catchvariableoverwriteslocal

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a different catch variable to preserve the local value."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CatchVariableOverwritesLocal" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTry} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	tr := n.(*syntax.Try)
	if tr.Finally != nil && syntax.Terminates(tr.Finally.Body) {
		return
	}
	prev, ok := astquery.PrevStmtNoDoc(ctx.File, tr)
	if !ok {
		return
	}
	st, ok := prev.(*syntax.ExprStmt)
	if !ok {
		return
	}
	a, ok := st.Expr.(*syntax.Assign)
	if !ok || a.ByRef || a.Op.Kind != syntax.TEqual {
		return
	}
	v, ok := a.Var.(*syntax.Variable)
	if !ok || v.Name == "" {
		return
	}
	if _, ok := semanticquery.NativeValue(ctx, a.Value).(*syntax.Literal); !ok {
		return
	}
	read, ok := astquery.NextEcho(ctx.File, tr).(*syntax.Variable)
	if !ok || read.Name != v.Name {
		return
	}
	throwing := false
	for _, s := range tr.Body.Stmts {
		expr, ok := s.(*syntax.ExprStmt)
		if !ok {
			return
		}
		switch expr.Expr.(type) {
		case *syntax.FuncCall, *syntax.Throw:
			throwing = true
		default:
			return
		}
	}
	if !throwing {
		return
	}
	for _, c := range tr.Catches {
		if c.Var != nil && c.Var.Name == v.Name && !syntax.Terminates(c.Body) && !replaced(c.Body, v.Name) && (tr.Finally == nil || !replaced(tr.Finally.Body, v.Name)) {
			ctx.ReportNode(c.Var, message)
		}
	}
}

func replaced(body syntax.Node, name string) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			found = found || astquery.MentionsVariable(x.Var, name)
		case *syntax.IncDec:
			found = found || astquery.MentionsVariable(x.Var, name)
		case *syntax.Foreach:
			found = found || astquery.MentionsVariable(x.Value, name)
		case *syntax.Unset:
			found = found || astquery.MentionsVariable(x, name)
		}
		return !found
	})
	return found
}
