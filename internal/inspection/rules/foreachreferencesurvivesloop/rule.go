// Package foreachreferencesurvivesloop implements the native ForeachReferenceSurvivesLoop inspection.
package foreachreferencesurvivesloop

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

const message = "Unset the foreach reference before reusing the variable."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ForeachReferenceSurvivesLoop" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.Foreach)
	v, ok := loop.Value.(*syntax.Variable)
	if !loop.ByRef || !ok || v.Name == "" {
		return
	}
	iterable := syntax.UnwrapParens(loop.Expr)
	if local, ok := iterable.(*syntax.Variable); ok {
		if prev, known := astquery.PrevStmtNoDoc(ctx.File, loop); known {
			if st, ok := prev.(*syntax.ExprStmt); ok {
				if a, ok := st.Expr.(*syntax.Assign); ok && !a.ByRef && a.Op.Kind == syntax.TEqual {
					if target, ok := a.Var.(*syntax.Variable); ok && target.Name == local.Name {
						iterable = syntax.UnwrapParens(a.Value)
					}
				}
			}
		}
	}
	if array, ok := iterable.(*syntax.Array); ok && len(array.Items) == 0 {
		return
	}
	cleared := false
	syntax.Inspect(loop.Body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.Unset:
			if astquery.MentionsVariable(x, v.Name) {
				cleared = true
			}
		case *syntax.Assign:
			if x.ByRef && astquery.MentionsVariable(x.Var, v.Name) {
				cleared = true
			}
		}
		return !cleared
	})
	if cleared {
		return
	}
	list, index, ok := astquery.StmtList(ctx.File, loop)
	if !ok {
		return
	}
	for _, s := range list[index+1:] {
		if unset, ok := s.(*syntax.Unset); ok {
			for _, e := range unset.Vars {
				if x, ok := e.(*syntax.Variable); ok && x.Name == v.Name {
					return
				}
			}
			return
		}
		st, ok := s.(*syntax.ExprStmt)
		if !ok {
			return
		}
		a, ok := st.Expr.(*syntax.Assign)
		if !ok {
			return
		}
		x, ok := a.Var.(*syntax.Variable)
		if !ok {
			return
		}
		if x.Name == v.Name {
			if !a.ByRef {
				ctx.ReportNode(a, message)
			}
			return
		}
		if astquery.MentionsVariable(a, v.Name) {
			return
		}
	}
}
