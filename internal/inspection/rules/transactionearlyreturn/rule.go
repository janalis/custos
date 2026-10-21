// Package transactionearlyreturn implements the native TransactionEarlyReturn inspection.
package transactionearlyreturn

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Close the owned transaction before returning."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TransactionEarlyReturn" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KReturn} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	ret := n.(*syntax.Return)

	caught := false
	for p := ret.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.Catch); ok {
			caught = true
			break
		}
	}
	if caught {
		return
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(ret)) {
		begin, ok := call.Node.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, begin, "PDO", "beginTransaction") || begin.Span().End >= ret.Span().Start {
			continue
		}
		state, known := ctx.Flow().StateBefore(ret, begin.Var, "begintransaction")
		if !known || !state.Successful || semanticquery.CalledInFinally(ctx, ret, begin.Var, "commit", "rollback") {
			continue
		}
		if transactionClosedByExpr(ctx, ret.Expr, begin.Var) || transactionGuardedFinally(ctx, ret, begin.Var) {
			continue
		}
		ctx.ReportNode(ret, message)
		return
	}
}

// transactionCloseCall reports whether e is a PDO commit or rollBack call on
// the connection that began the transaction.
func transactionCloseCall(ctx *analysis.Context, e, receiver syntax.Expr) bool {
	c, ok := e.(*syntax.MethodCall)
	return ok && (semanticquery.NativeMethod(ctx, c, "PDO", "commit") || semanticquery.NativeMethod(ctx, c, "PDO", "rollBack")) && semanticquery.NativeSameValue(ctx, receiver, c.Var)
}

// transactionClosedByExpr proves that evaluating a returned expression always
// ends the transaction: the closing call is evaluated unconditionally, or by
// every arm of a ternary or match. Short-circuit right operands prove nothing.
func transactionClosedByExpr(ctx *analysis.Context, e, receiver syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.MethodCall:
		return transactionCloseCall(ctx, x, receiver)
	case *syntax.Paren:
		return transactionClosedByExpr(ctx, x.Expr, receiver)
	case *syntax.Unary:
		return transactionClosedByExpr(ctx, x.Expr, receiver)
	case *syntax.Assign:
		// `??=` only evaluates its value when the target is unset or null.
		return x.Op.Kind != syntax.TCoalesceEqual && transactionClosedByExpr(ctx, x.Value, receiver)
	case *syntax.Binary:
		if transactionClosedByExpr(ctx, x.Left, receiver) {
			return true
		}
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
			return false
		}
		return transactionClosedByExpr(ctx, x.Right, receiver)
	case *syntax.Ternary:
		if transactionClosedByExpr(ctx, x.Cond, receiver) {
			return true
		}
		return x.Then != nil && transactionClosedByExpr(ctx, x.Then, receiver) && transactionClosedByExpr(ctx, x.Else, receiver)
	case *syntax.Match:
		if transactionClosedByExpr(ctx, x.Cond, receiver) {
			return true
		}
		// An unmatched subject throws, which is not a normal return.
		for _, arm := range x.Arms {
			if !transactionClosedByExpr(ctx, arm.Body, receiver) {
				return false
			}
		}
		return len(x.Arms) > 0
	}
	return false
}

// transactionGuardedFinally accepts the cleanup idiom of a surrounding
// finally block that closes the transaction only while one is still active:
// `if ($pdo->inTransaction()) { $pdo->rollBack(); }`. Other conditions do not
// prove cleanup, and a return inside the finally block itself is not covered.
func transactionGuardedFinally(ctx *analysis.Context, ret *syntax.Return, receiver syntax.Expr) bool {
	for p := ret.Parent(); p != nil; p = p.Parent() {
		tr, ok := p.(*syntax.Try)
		if !ok || tr.Finally == nil || ret.Span().Start >= tr.Finally.Span().Start {
			continue
		}
		for _, s := range tr.Finally.Body.Stmts {
			guard, ok := s.(*syntax.If)
			if !ok {
				continue
			}
			cond, ok := guard.Cond.(*syntax.MethodCall)
			if !ok || !semanticquery.NativeMethod(ctx, cond, "PDO", "inTransaction") || !semanticquery.NativeSameValue(ctx, receiver, cond.Var) {
				continue
			}
			stmts := []syntax.Stmt{guard.Body}
			if block, ok := guard.Body.(*syntax.Block); ok {
				stmts = block.Stmts
			}
			for _, inner := range stmts {
				if st, ok := inner.(*syntax.ExprStmt); ok && transactionCloseCall(ctx, st.Expr, receiver) {
					return true
				}
			}
		}
	}
	return false
}
