package semanticquery

import "custos/internal/php/syntax"

// expansionDFailureBranch requires a falsy condition and rejection on every
// structured path through its body. Unknown control flow establishes no proof.
func expansionDFailureBranch(e syntax.Expr) *syntax.If {
	odd := false
	for budget := 256; budget > 0; budget-- {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			if p.Op.Kind != syntax.TEqual || p.ByRef || p.Value != e {
				return nil
			}
			e = p
		case *syntax.Unary:
			switch p.Op.Kind {
			case syntax.TExclaim:
				odd = !odd
			case syntax.TAt:
			default:
				return nil
			}
			e = p
		case *syntax.Binary:
			if p.Op.Kind != syntax.TIsEqual || !(p.Left == e && expansionDFalse(p.Right) || p.Right == e && expansionDFalse(p.Left)) {
				return nil
			}
			odd = !odd
			e = p
		case *syntax.If:
			if !odd || p.Cond != e {
				return nil
			}
			remaining := 256
			if failureBranchStructured(p.Body, &remaining) && failureBranchOutcomes(p.Body, &remaining) == failureReject {
				return p
			}
			return nil
		default:
			return nil
		}
	}
	return nil
}

type failureOutcome uint8

const (
	failureFallthrough failureOutcome = 1 << iota
	failureReject
	failureUnknown
)

func failureBranchStructured(body syntax.Stmt, remaining *int) bool {
	known := true
	syntax.Inspect(body, func(node syntax.Node) bool {
		if !known {
			return false
		}
		*remaining--
		if *remaining < 0 {
			known = false
			return false
		}
		if syntax.IsVariableScope(node) {
			return false
		}
		switch node.(type) {
		case *syntax.Yield, *syntax.YieldFrom, *syntax.Goto, *syntax.Label:
			known = false
		}
		return known
	})
	return known
}

func failureBranchOutcomes(statement syntax.Stmt, remaining *int) failureOutcome {
	*remaining--
	if *remaining < 0 {
		return failureUnknown
	}
	switch s := statement.(type) {
	case *syntax.Block:
		out := failureFallthrough
		for _, child := range s.Stmts {
			if out&failureFallthrough == 0 {
				break
			}
			out = out&^failureFallthrough | failureBranchOutcomes(child, remaining)
		}
		return out
	case *syntax.Return:
		if expansionDFalse(s.Expr) {
			return failureReject
		}
		return failureUnknown
	case *syntax.ExprStmt:
		if _, ok := syntax.UnwrapParens(s.Expr).(*syntax.Throw); ok {
			return failureReject
		}
		if failureExpressionContinues(s.Expr, remaining) {
			return failureFallthrough
		}
	case *syntax.If:
		if !failureExpressionContinues(s.Cond, remaining) {
			return failureUnknown
		}
		out := failureBranchOutcomes(s.Body, remaining)
		for _, branch := range s.ElseIfs {
			if !failureExpressionContinues(branch.Cond, remaining) {
				return failureUnknown
			}
			out |= failureBranchOutcomes(branch.Body, remaining)
		}
		if s.Else == nil {
			return out | failureFallthrough
		}
		return out | failureBranchOutcomes(s.Else.Body, remaining)
	case *syntax.Nop:
		return failureFallthrough
	}
	return failureUnknown
}

// Suspension and embedded abrupt exits cannot be treated as plain statements.
// Nested function bodies execute in a different scope and do not affect this one.
func failureExpressionContinues(expression syntax.Expr, remaining *int) bool {
	continues := true
	syntax.Inspect(expression, func(node syntax.Node) bool {
		if !continues {
			return false
		}
		*remaining--
		if *remaining < 0 {
			continues = false
			return false
		}
		if syntax.IsVariableScope(node) {
			return false
		}
		switch node.(type) {
		case *syntax.Throw:
			continues = false
		default:
			if expr, ok := node.(syntax.Expr); ok && syntax.ExitInvocation(expr) {
				continues = false
			}
		}
		return continues
	})
	return continues
}
