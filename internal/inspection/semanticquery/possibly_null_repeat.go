package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// Repeated accesses: once an access on a possibly-null value is reported,
// a later access on that same value that only runs after it repeats the
// finding (see repeatedAccess).

// maxRepeatScan bounds the earlier accesses repeatedAccess compares an
// access with.
const maxRepeatScan = 64

// repeatedAccess reports whether access n, whose receiver recv may be
// null, repeats an earlier access on the same value (a local or a property
// of `$this`, see valueKey) that may be null too and runs before n on every
// path (see runsBefore), with no other use of the value between them: no
// write, no comparison, no argument. In a loop around n that does not hold
// the earlier access, no such use anywhere in the loop either.
func repeatedAccess(ctx *analysis.Context, n syntax.Node, recv syntax.Expr) bool {
	value := valueKey(recv)
	if value == "" {
		return false
	}
	u := valueAccesses(ctx, syntax.EnclosingVariableScope(n), value)
	start := recv.Span().Start
	scanned := 0
	for i := len(u.accesses) - 1; i >= 0 && scanned < maxRepeatScan; i-- {
		c := u.accesses[i]
		if c.Span().End > start {
			continue
		}
		scanned++
		if !runsBefore(c, n) {
			continue
		}
		if usedBetween(u.others, c, n, start) {
			return false // earlier accesses have this use in between too
		}
		if _, ok := possiblyNullAccess(ctx, c); ok {
			return true
		}
	}
	return false
}

// usedBetween reports whether one of the occurrences others lies after
// access c and before start (n's receiver), or in a loop around n that
// does not hold c.
func usedBetween(others []syntax.Expr, c, n syntax.Node, start uint32) bool {
	from, end := c.Span().End, start
	for p := n.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if sp := p.Span(); sp.Start >= from {
				end = max(end, sp.End)
			}
		}
	}
	for _, o := range others {
		if s := o.Span().Start; s >= from && s < end {
			return true
		}
	}
	return false
}

type valueUsesKey struct {
	scope syntax.Node
	value string
}

type valueUses struct {
	accesses []syntax.Node // accesses whose receiver holds the value, in source order
	others   []syntax.Expr // the other occurrences of the value
}

// valueAccesses lists the accesses (as PossiblyNullAccess takes them) in
// scope whose receiver holds value (see valueKey), and the occurrences of
// value that are no such receiver or are written through it (`$x['k'] =
// 1`). Nested functions are another scope. Cached per scope and value.
func valueAccesses(ctx *analysis.Context, scope syntax.Node, value string) *valueUses {
	return ctx.Shared(valueUsesKey{scope, value}, func() any {
		u := &valueUses{}
		root := scope
		if scope == nil {
			root = ctx.File
		}
		syntax.Inspect(root, func(n syntax.Node) bool {
			if n != root && syntax.IsVariableScope(n) {
				return false
			}
			x, ok := n.(syntax.Expr)
			if !ok || valueKey(x) != value {
				return true
			}
			if a := accessOf(ctx, x); a != nil {
				u.accesses = append(u.accesses, a)
			} else {
				u.others = append(u.others, x)
			}
			return true
		})
		return u
	}).(*valueUses)
}

// valueKey names the value x holds when repeated accesses on it can be
// told apart: a local (`$x`) or a property of `$this` (`$this->x`); ""
// otherwise.
func valueKey(x syntax.Expr) string {
	switch v := x.(type) {
	case *syntax.Variable:
		if v.NameExpr == nil && v.Name != "" && v.Name != "this" {
			return "$" + v.Name
		}
	case *syntax.PropertyFetch:
		this, ok := v.Var.(*syntax.Variable)
		id, named := v.Name.(*syntax.Identifier)
		if ok && named && !v.NullSafe && this.NameExpr == nil && this.Name == "this" {
			return "->" + id.Value
		}
	}
	return ""
}

// accessOf returns the access x is the receiver of, nil when x is no
// receiver of an access PossiblyNullAccess checks or is the array of an
// element written (which makes null an array).
func accessOf(ctx *analysis.Context, x syntax.Expr) syntax.Node {
	switch p := x.Parent().(type) {
	case *syntax.PropertyFetch:
		if p.Var == x {
			return p
		}
	case *syntax.MethodCall:
		if p.Var == x {
			return p
		}
	case *syntax.StaticCall:
		if p.Class == x {
			return p
		}
	case *syntax.ClassConstFetch:
		if p.Class == x {
			return p
		}
	case *syntax.StaticPropertyFetch:
		if p.Class == x {
			return p
		}
	case *syntax.ArrayDimFetch:
		if p.Var == x && !elementWrite(ctx, p) {
			return p
		}
	}
	return nil
}

// runsBefore reports whether access c, written before access n in the same
// scope, has run whenever n runs: from c up to the first node also
// enclosing n, every step is evaluated unconditionally (see evaluated),
// and there c's part is evaluated before n's (see evaluatedBefore).
func runsBefore(c, n syntax.Node) bool {
	sides := map[syntax.Node]syntax.Node{} // n's ancestors: the child on n's side
	child := n
	for p := n.Parent(); p != nil; child, p = p, p.Parent() {
		sides[p] = child
		if syntax.IsVariableScope(p) {
			break
		}
	}
	// Both lie in one scope: climbing from c meets an ancestor of n.
	child, p := c, c.Parent()
	for sides[p] == nil {
		if !evaluated(p, child) {
			return false
		}
		child, p = p, p.Parent()
	}
	return evaluatedBefore(p, child, sides[p])
}

// evaluated reports whether child is evaluated whenever its parent p is:
// not a branch, a short-circuited operand, a loop body or step, a `try`
// block, an argument after a nullsafe access, nor in another function.
func evaluated(p, child syntax.Node) bool {
	switch p := p.(type) {
	case *syntax.Binary:
		if child == syntax.Node(p.Right) {
			switch p.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
				return false
			}
		}
		return true
	case *syntax.Ternary:
		return child == syntax.Node(p.Cond)
	case *syntax.Assign:
		return !(p.Op.Kind == syntax.TCoalesceEqual && child == syntax.Node(p.Value))
	case *syntax.MethodCall:
		return child == syntax.Node(p.Var) || !p.NullSafe && !inNullsafeChain(p.Var)
	case *syntax.PropertyFetch:
		return child == syntax.Node(p.Var) || !p.NullSafe && !inNullsafeChain(p.Var)
	case *syntax.ArrayDimFetch:
		return child == syntax.Node(p.Var) || !inNullsafeChain(p.Var)
	case *syntax.Isset, *syntax.Match, *syntax.MatchArm:
		return false // isset() stops at the first unset argument
	case *syntax.ExprStmt, *syntax.Echo, *syntax.Return, *syntax.Block:
		return true
	case *syntax.If:
		return child == syntax.Node(p.Cond)
	case *syntax.While:
		return child == syntax.Node(p.Cond)
	case *syntax.Foreach:
		return child == syntax.Node(p.Expr)
	case *syntax.Switch:
		return child == syntax.Node(p.Cond)
	case *syntax.For:
		for _, x := range p.Init {
			if child == syntax.Node(x) {
				return true
			}
		}
		return false
	case syntax.Stmt:
		return false
	}
	return !syntax.IsVariableScope(p)
}

// evaluatedBefore reports whether, in p, child cc (holding the earlier
// access) is evaluated before child nc (holding the later one) on every
// path reaching nc.
func evaluatedBefore(p, cc, nc syntax.Node) bool {
	switch p := p.(type) {
	case *syntax.Block, *syntax.Case, *syntax.File, *syntax.Namespace, *syntax.Echo, *syntax.ArgList, *syntax.Array, *syntax.ArrayItem, *syntax.Binary:
		return true // statements, items and operands in source order
	case *syntax.If:
		return cc == syntax.Node(p.Cond)
	case *syntax.ElseIf:
		return cc == syntax.Node(p.Cond)
	case *syntax.While:
		return cc == syntax.Node(p.Cond)
	case *syntax.Foreach:
		return cc == syntax.Node(p.Expr)
	case *syntax.Switch:
		return cc == syntax.Node(p.Cond)
	case *syntax.Match:
		return cc == syntax.Node(p.Cond)
	case *syntax.Ternary:
		return cc == syntax.Node(p.Cond)
	case *syntax.For:
		for _, x := range p.Init {
			if cc == syntax.Node(x) {
				return true
			}
		}
	case *syntax.MethodCall:
		return cc == syntax.Node(p.Var)
	case *syntax.StaticCall:
		return cc == syntax.Node(p.Class)
	case *syntax.New:
		return cc == syntax.Node(p.Class)
	case *syntax.FuncCall:
		return cc == syntax.Node(p.Name)
	case *syntax.PropertyFetch:
		return cc == syntax.Node(p.Var)
	case *syntax.ArrayDimFetch:
		return cc == syntax.Node(p.Var)
	}
	return false
}
