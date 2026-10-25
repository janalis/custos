// Package sessionmutationafterclose implements the native SessionMutationAfterClose inspection.
package sessionmutationafterclose

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

const message = "Persist session changes before closing the session."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionMutationAfterClose" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign, syntax.KIncDec} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !sessionMutation(n) {
		return
	}
	state, known := ctx.Flow().GlobalStateBefore(n, "session")
	if known && (state.Operation == "session_write_close" || state.Operation == "session_commit" || state.Operation == "session_abort") {
		ctx.ReportNode(n, message)
	}
}

func sessionMutation(n syntax.Node) bool {
	var expr syntax.Expr
	if assign, ok := n.(*syntax.Assign); ok {
		if arr, empty := syntax.UnwrapParens(assign.Value).(*syntax.Array); empty && len(arr.Items) == 0 && assign.Op.Kind == syntax.TEqual {
			if v, whole := assign.Var.(*syntax.Variable); whole && v.Name == "_SESSION" {
				return false // E3: clearing the closed session's in-memory copy
			}
		}
		expr = assign.Var
	} else {
		expr = n.(*syntax.IncDec).Var
	}
	for {
		fetch, ok := expr.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		expr = fetch.Var
	}
	v, ok := expr.(*syntax.Variable)
	return ok && v.Name == "_SESSION"
}
