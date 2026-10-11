// Package signalwaitfailureacceptedassignal implements the native SignalWaitFailureAcceptedAsSignal inspection.
package signalwaitfailureacceptedassignal

import (
	"math"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject signal-wait failure before handling the result."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SignalWaitFailureAcceptedAsSignal" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.If)
	e := syntax.UnwrapParens(b.Cond)
	var assigned syntax.Expr
	if a, ok := e.(*syntax.Assign); ok {
		assigned = a.Var
		e = a.Value
	}
	c, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "pcntl_sigwaitinfo") {
		return
	}
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
	if len(calls) > 256 {
		return
	}
	// An assigned variable stays the wait result up to its first rebinding,
	// alias or dynamic write in the branch; other targets need the first call.
	var limit uint32
	if v, ok := assigned.(*syntax.Variable); ok && v.NameExpr == nil {
		limit = stableUntil(ctx, b.Body, v)
	}
	for _, record := range calls {
		call, ok := record.Node.(*syntax.FuncCall)
		if !ok || !b.Body.Span().Contains(call.Span()) || call.Args == nil {
			continue
		}
		for _, node := range call.Args.Args {
			a, ok := node.(*syntax.Arg)
			if !ok || a.Unpack {
				continue
			}
			same := semanticquery.NativeLocalValue(ctx, a.Value) == c
			if !same && assigned != nil && astquery.Equivalent(ctx.File, assigned, a.Value) {
				same = a.Value.Span().Start < limit || (limit == 0 && directFirst(b.Body, call))
			}
			if same && !failureRejected(ctx, c, a.Value) {
				ctx.ReportNode(b.Cond, message)
				return
			}
		}
	}
}

func directFirst(body syntax.Stmt, call *syntax.FuncCall) bool {
	if block, ok := body.(*syntax.Block); ok {
		body = block.Stmts[0]
	}
	st, ok := body.(*syntax.ExprStmt)
	return ok && st.Expr == call
}

// stableUntil returns the offset in body from which v may no longer hold the
// value assigned by the condition.
func stableUntil(ctx *analysis.Context, body syntax.Stmt, v *syntax.Variable) uint32 {
	limit := uint32(math.MaxUint32)
	stop := func(n syntax.Node) {
		limit = min(limit, n.Span().Start)
	}
	for _, access := range flowquery.VarAccesses(ctx.File, syntax.EnclosingVariableScope(v), v.Name) {
		if !body.Span().Contains(access.Var.Span()) {
			continue
		}
		use, captured := access.By.(*syntax.ClosureUse)
		if access.Write || access.ElemWrite || (captured && use.ByRef) || aliased(access.Var) {
			stop(access.Var)
		}
	}
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Variable:
			if x.NameExpr != nil {
				stop(x)
			}
		case *syntax.Eval, *syntax.Include:
			stop(n)
		case *syntax.FuncCall:
			if name := semanticquery.NativeBuiltinName(ctx, x); name == "extract" || name == "parse_str" {
				stop(x)
			}
		}
		return true
	})
	return limit
}

// aliased reports a read that may bind v by reference: a reference
// assignment or array item, or an argument of anything but a plain function
// call (which is itself a consuming call).
func aliased(v *syntax.Variable) bool {
	switch p := v.Parent().(type) {
	case *syntax.Assign:
		return p.ByRef
	case *syntax.ArrayItem:
		return p.ByRef
	case *syntax.Arg:
		_, plain := p.Parent().Parent().(*syntax.FuncCall)
		return !plain
	}
	return false
}

// failureRejected reports whether every path to value passed a comparison
// that excludes -1, made after the wait call c returned: an enclosing
// condition, or an earlier sibling `if` whose body leaves the block.
func failureRejected(ctx *analysis.Context, c *syntax.FuncCall, value syntax.Expr) bool {
	excludes := func(cond syntax.Expr, truth bool) bool {
		return cond.Span().Start >= c.Span().End && rejects(ctx, c, value, cond, truth)
	}
	for child := syntax.Node(value); child != nil && !syntax.IsVariableScope(child); child = child.Parent() {
		parent := child.Parent()
		switch p := parent.(type) {
		case *syntax.If:
			if child == syntax.Node(p.Body) {
				if excludes(p.Cond, true) {
					return true
				}
			} else if child != syntax.Node(p.Cond) {
				if excludes(p.Cond, false) {
					return true
				}
				for _, branch := range p.ElseIfs {
					if syntax.Node(branch) == child {
						break
					}
					if excludes(branch.Cond, false) {
						return true
					}
				}
			}
		case *syntax.ElseIf:
			if child == syntax.Node(p.Body) && excludes(p.Cond, true) {
				return true
			}
		case *syntax.Ternary:
			if child != syntax.Node(p.Cond) && excludes(p.Cond, child != syntax.Node(p.Else)) {
				return true
			}
		case *syntax.Binary:
			if child == syntax.Node(p.Right) {
				switch p.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					if excludes(p.Left, true) {
						return true
					}
				case syntax.TBooleanOr, syntax.TOr:
					if excludes(p.Left, false) {
						return true
					}
				}
			}
		}
		list, _ := syntax.StmtListOf(parent)
		if parent == nil {
			list = ctx.File.Stmts
		}
		for _, st := range list {
			if st.Span().Start >= child.Span().Start {
				break
			}
			if guard, ok := st.(*syntax.If); ok && syntax.Terminates(guard.Body) && excludes(guard.Cond, false) {
				return true
			}
		}
	}
	return false
}

// rejects reports whether cond evaluating to truth proves the tracked value
// is not -1, from integer comparisons joined by !, && and ||.
func rejects(ctx *analysis.Context, c *syntax.FuncCall, value, cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && rejects(ctx, c, value, x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return truth && (rejects(ctx, c, value, x.Left, true) || rejects(ctx, c, value, x.Right, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !truth && (rejects(ctx, c, value, x.Left, false) || rejects(ctx, c, value, x.Right, false))
		}
		for i, pair := range [][2]syntax.Expr{{x.Left, x.Right}, {x.Right, x.Left}} {
			bound, ok := semanticquery.NativeInt(ctx, pair[1])
			if !ok || !(astquery.Equivalent(ctx.File, value, pair[0]) || semanticquery.NativeLocalValue(ctx, pair[0]) == c) {
				continue
			}
			left, right := int64(-1), bound
			if i == 1 {
				left, right = bound, -1
			}
			var holds bool
			switch x.Op.Kind {
			case syntax.TIsEqual, syntax.TIsIdentical:
				holds = left == right
			case syntax.TIsNotEqual, syntax.TIsNotIdentical:
				holds = left != right
			case syntax.TLess:
				holds = left < right
			case syntax.TIsSmallerOrEqual:
				holds = left <= right
			case syntax.TGreater:
				holds = left > right
			case syntax.TIsGreaterOrEqual:
				holds = left >= right
			default:
				return false
			}
			return holds != truth
		}
	}
	return false
}
