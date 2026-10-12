package posixaccountlookupfailuredereferenced

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// guard proves lexically that a failed lookup cannot reach a read of target.
// Only tests of the value produced by call are considered.
type guard struct {
	ctx    *analysis.Context
	target syntax.Expr
	call   *syntax.FuncCall
}

// covers walks the enclosing conditions and the earlier aborting statements
// of every enclosing statement list, up to the variable scope.
func (g guard) covers() bool {
	if _, ok := syntax.UnwrapParens(g.target).(*syntax.Variable); !ok {
		return false
	}
	var child syntax.Node = g.target
	for p := g.target.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Ternary:
			if (child == syntax.Node(x.Then) && g.proves(x.Cond, true)) || (child == syntax.Node(x.Else) && g.proves(x.Cond, false)) {
				return true
			}
		case *syntax.Binary:
			if child == syntax.Node(x.Right) && shortCircuit(x) && g.proves(x.Left, runsRightWhen(x)) {
				return true
			}
		case *syntax.If:
			if g.branch(x, g.target.Span()) {
				return true
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, listed := astquery.StmtList(g.ctx.File, st)
		if !listed {
			continue
		}
		for _, prev := range list[:i] {
			if g.rejects(prev) {
				return true
			}
		}
	}
	return false
}

// runsRightWhen gives the left outcome that lets a short-circuit operator
// evaluate its right operand; callers check shortCircuit first.
func runsRightWhen(b *syntax.Binary) bool {
	return b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TAnd
}

// rejects reports whether a statement leaves the list whenever the lookup
// failed: an early-exit branch, or a short-circuit abort such as `or throw`.
func (g guard) rejects(s syntax.Stmt) bool {
	switch x := s.(type) {
	case *syntax.If:
		return syntax.Terminates(x.Body) && !jumps(x.Body) && g.proves(x.Cond, false)
	case *syntax.ExprStmt:
		b, ok := syntax.UnwrapParens(x.Expr).(*syntax.Binary)
		if !ok || !shortCircuit(b) {
			return false
		}
		if _, thrown := syntax.UnwrapParens(b.Right).(*syntax.Throw); !thrown && !syntax.ExitInvocation(b.Right) {
			return false
		}
		return g.proves(b.Left, !runsRightWhen(b))
	}
	return false
}

// jumps reports whether a branch contains a goto, which may land on a label
// later in the same statement list instead of leaving it.
func jumps(body syntax.Stmt) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		_, isGoto := n.(*syntax.Goto)
		found = found || isGoto
		return !found
	})
	return found
}

func shortCircuit(b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
		return true
	}
	return false
}

// branch resolves which outcome of each condition leads to the branch at at.
func (g guard) branch(x *syntax.If, at syntax.Span) bool {
	if x.Body.Span().Contains(at) {
		return g.proves(x.Cond, true)
	}
	if x.Cond.Span().Contains(at) {
		return false
	}
	if g.proves(x.Cond, false) {
		return true
	}
	for _, e := range x.ElseIfs {
		if e.Span().Contains(at) {
			return e.Body.Span().Contains(at) && g.proves(e.Cond, true)
		}
		if g.proves(e.Cond, false) {
			return true
		}
	}
	return false
}

// proves reports whether cond evaluating to truth excludes the failure value.
func (g guard) proves(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Variable:
		return truth && g.same(x)
	case *syntax.Assign:
		// `$info = posix_getpwnam(...)` used as a condition tests the result.
		v, plain := x.Var.(*syntax.Variable)
		return truth && plain && semanticquery.NativeUnsilenced(x.Value) == syntax.Expr(g.call) && astquery.Equivalent(g.ctx.File, v, g.target)
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.proves(x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			if truth {
				return g.proves(x.Left, true) || g.proves(x.Right, true)
			}
			return g.proves(x.Left, false) && g.proves(x.Right, false)
		case syntax.TBooleanOr, syntax.TOr:
			if truth {
				return g.proves(x.Left, true) && g.proves(x.Right, true)
			}
			return g.proves(x.Left, false) || g.proves(x.Right, false)
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && g.comparesFalse(x)
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && g.comparesFalse(x)
		}
	case *syntax.Isset:
		// isset($info) accepts false; only a key probe implies an array.
		for _, v := range x.Vars {
			if _, keyed := syntax.UnwrapParens(v).(*syntax.ArrayDimFetch); keyed && truth && g.same(root(v)) {
				return true
			}
		}
	case *syntax.Empty:
		return !truth && g.same(root(x.Expr))
	case *syntax.FuncCall:
		return truth && semanticquery.NativeBuiltin(g.ctx, x, "is_array") && g.same(semanticquery.CallArgument(x.Args, 0, "value"))
	}
	return false
}

func (g guard) comparesFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		c, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same ties a tested expression to the value that reaches the read. A test
// before the lookup, or of a reassigned variable, proves nothing.
func (g guard) same(e syntax.Expr) bool {
	if e == nil || e.Span().Start < g.call.Span().End || !astquery.Equivalent(g.ctx.File, e, g.target) {
		return false
	}
	tested, read := g.ctx.Flow().Value(e), g.ctx.Flow().Value(g.target)
	return tested.Expr == read.Expr && tested.Invalidated == read.Invalidated
}

// root strips array accesses down to the indexed expression.
func root(e syntax.Expr) syntax.Expr {
	for {
		e = syntax.UnwrapParens(e)
		fetch, ok := e.(*syntax.ArrayDimFetch)
		if !ok {
			return e
		}
		e = fetch.Var
	}
}
