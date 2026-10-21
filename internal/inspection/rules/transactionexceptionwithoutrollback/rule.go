// Package transactionexceptionwithoutrollback implements the native TransactionExceptionWithoutRollback inspection.
package transactionexceptionwithoutrollback

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Roll back the active transaction on the caught failure path."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TransactionExceptionWithoutRollback" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KReturn} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	ret := n.(*syntax.Return)

	caught := false
	for p := ret.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.Catch); ok {
			caught = true
			break
		}
	}
	if !caught {
		return
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(ret)) {
		begin, ok := call.Node.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, begin, "PDO", "beginTransaction") || begin.Span().End >= ret.Span().Start {
			continue
		}
		state, known := ctx.Flow().StateBefore(ret, begin.Var, "begintransaction")
		if known && state.Successful && !semanticquery.CalledInFinally(ctx, ret, begin.Var, "commit", "rollback") && !rollbackCatchEnds(ctx, ret, begin.Var) {
			ctx.ReportNode(ret, message)
			return
		}
	}
}

// rollbackCatchEnds reports whether a catch-path return ends the transaction
// itself: its expression always runs commit or rollBack on the receiver, or an
// enclosing finally block does so under an inTransaction guard on it.
func rollbackCatchEnds(ctx *analysis.Context, ret *syntax.Return, receiver syntax.Expr) bool {
	identity := ctx.Flow().Value(receiver).Identity
	on := func(e syntax.Expr, names ...string) bool {
		c, ok := e.(*syntax.MethodCall)
		if !ok || identity == 0 || ctx.Flow().Value(c.Var).Identity != identity {
			return false
		}
		for _, name := range names {
			if semanticquery.NativeMethod(ctx, c, "PDO", name) {
				return true
			}
		}
		return false
	}
	ends := func(e syntax.Expr) bool { return on(e, "commit", "rollBack") }
	if rollbackCatchEvaluates(ret.Expr, ends) {
		return true
	}
	for p := ret.Parent(); p != nil; p = p.Parent() {
		tr, ok := p.(*syntax.Try)
		if !ok || tr.Finally == nil {
			continue
		}
		for _, s := range tr.Finally.Body.Stmts {
			guard, ok := s.(*syntax.If)
			if !ok || !on(guard.Cond, "inTransaction") {
				continue
			}
			stmts := []syntax.Stmt{guard.Body}
			if block, ok := guard.Body.(*syntax.Block); ok {
				stmts = block.Stmts
			}
			for _, inner := range stmts {
				if st, ok := inner.(*syntax.ExprStmt); ok && ends(st.Expr) {
					return true
				}
			}
		}
	}
	return false
}

// rollbackCatchEvaluates reports whether every evaluation of e runs an
// expression accepted by ends: directly, through parentheses or a unary
// operator, in a ternary or match subject, or in each of their arms.
func rollbackCatchEvaluates(e syntax.Expr, ends func(syntax.Expr) bool) bool {
	switch x := e.(type) {
	case *syntax.Paren:
		return rollbackCatchEvaluates(x.Expr, ends)
	case *syntax.Unary:
		return rollbackCatchEvaluates(x.Expr, ends)
	case *syntax.Ternary:
		return rollbackCatchEvaluates(x.Cond, ends) ||
			x.Then != nil && rollbackCatchEvaluates(x.Then, ends) && rollbackCatchEvaluates(x.Else, ends)
	case *syntax.Match:
		if rollbackCatchEvaluates(x.Cond, ends) {
			return true
		}
		// No arm matching throws, so only the arms decide an ordinary return.
		for _, arm := range x.Arms {
			if !rollbackCatchEvaluates(arm.Body, ends) {
				return false
			}
		}
		return true
	}
	return ends(e)
}
