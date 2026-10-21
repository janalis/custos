package flow

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

// intInterval is immutable. Literal singleton bounds need no allocation.
type intInterval struct{ lo, hi int64 }

// IntRange returns integer bounds that include every value reaching this read.
// Float overflow and paths with unknown or noninteger values are not bounded.
func (e *Env) IntRange(x syntax.Expr) (lo, hi int64, known bool) {
	if x == nil {
		return 0, 0, false
	}
	for _, p := range syntax.VariableScopeParams(syntax.EnclosingVariableScope(x)) {
		if p.ByRef {
			return 0, 0, false
		}
	}
	v := e.Value(x)
	if !v.Complete {
		return 0, 0, false
	}
	return integerBounds(v)
}

func integerBounds(v Value) (int64, int64, bool) {
	if v.interval != nil {
		return v.interval.lo, v.interval.hi, true
	}
	n, ok := literalInteger(v.Expr)
	return n, n, ok
}

func literalInteger(x syntax.Expr) (int64, bool) { return integerLiteralDepth(x, MaxDepth*8) }

func integerLiteralDepth(x syntax.Expr, budget int) (int64, bool) {
	if budget == 0 {
		return 0, false
	}
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal:
		if n.LitKind == syntax.LitInt {
			raw := strings.ReplaceAll(n.Raw, "_", "")
			base := 0
			if strings.HasPrefix(raw, "0o") || strings.HasPrefix(raw, "0O") {
				raw = raw[2:]
				base = 8
			}
			v, err := strconv.ParseInt(raw, base, 64)
			return v, err == nil
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TPlus || n.Op.Kind == syntax.TMinus {
			v, ok := integerLiteralDepth(n.Expr, budget-1)
			if n.Op.Kind == syntax.TMinus {
				return -v, ok && v != math.MinInt64
			}
			return v, ok
		}
	}
	return 0, false
}

func integerValue(n int64) Value      { return Value{Complete: true, interval: &intInterval{n, n}} }
func sameInts(a, b *intInterval) bool { return a == b || a != nil && b != nil && *a == *b }
func joinInts(a, b Value) *intInterval {
	if a.Expr == b.Expr && a.interval == nil && b.interval == nil {
		return nil
	}
	lo, hi, ok := integerBounds(a)
	bl, bh, bok := integerBounds(b)
	if !ok || !bok {
		return nil
	}
	lo, hi = min(lo, bl), max(hi, bh)
	if a.interval != nil && a.interval.lo == lo && a.interval.hi == hi {
		return a.interval
	}
	if b.interval != nil && b.interval.lo == lo && b.interval.hi == hi {
		return b.interval
	}
	return &intInterval{lo, hi}
}

func addInt(a, b int64) (int64, bool) {
	n := a + b
	return n, !(b > 0 && n < a || b < 0 && n > a)
}

func subInt(a, b int64) (int64, bool) {
	n := a - b
	return n, !(b < 0 && n < a || b > 0 && n > a)
}

func arithmetic(op syntax.TokenKind, a, b Value) *intInterval {
	lo, hi, ok := integerBounds(a)
	bl, bh, bok := integerBounds(b)
	if !ok || !bok {
		return nil
	}
	var low, high int64
	var lok, hok bool
	switch op {
	case syntax.TPlus, syntax.TPlusEqual:
		low, lok = addInt(lo, bl)
		high, hok = addInt(hi, bh)
	case syntax.TMinus, syntax.TMinusEqual:
		low, lok = subInt(lo, bh)
		high, hok = subInt(hi, bl)
	default:
		return nil
	}
	if !lok || !hok {
		return nil
	}
	return &intInterval{low, high}
}

func unaryInts(op syntax.TokenKind, v Value) *intInterval {
	lo, hi, ok := integerBounds(v)
	if !ok {
		return nil
	}
	switch op {
	case syntax.TPlus:
		if v.interval != nil {
			return v.interval
		}
		return &intInterval{lo, hi}
	case syntax.TMinus:
		if lo != math.MinInt64 {
			return &intInterval{-hi, -lo}
		}
	}
	return nil
}

// widenInterval widens the second visit to a loop head. Comparison guards
// restore finite body bounds. Arithmetic must still reject extrema overflow.
func widenInterval(old, joined Value) *intInterval {
	lo, hi, known := integerBounds(joined)
	beforeLo, beforeHi, beforeKnown := integerBounds(old)
	if !known || !beforeKnown || lo == beforeLo && hi == beforeHi {
		return joined.interval
	}
	if lo < beforeLo {
		lo = math.MinInt64
	}
	if hi > beforeHi {
		hi = math.MaxInt64
	}
	if joined.interval != nil && joined.interval.lo == lo && joined.interval.hi == hi {
		return joined.interval
	}
	return &intInterval{lo, hi}
}

func (e *Env) refineInts(n *syntax.Binary, f *frame, truth bool) {
	variable, ok := syntax.UnwrapParens(n.Left).(*syntax.Variable)
	bound, known := literalInteger(n.Right)
	op := n.Op.Kind
	if !ok {
		variable, ok = syntax.UnwrapParens(n.Right).(*syntax.Variable)
		bound, known = literalInteger(n.Left)
		switch op {
		case syntax.TLess:
			op = syntax.TGreater
		case syntax.TGreater:
			op = syntax.TLess
		case syntax.TIsSmallerOrEqual:
			op = syntax.TIsGreaterOrEqual
		case syntax.TIsGreaterOrEqual:
			op = syntax.TIsSmallerOrEqual
		}
	}
	if !ok || !known || variable.NameExpr != nil {
		return
	}
	v := f.vars[variable.Name]
	lo, hi, known := integerBounds(v)
	if !known {
		return
	}
	beforeLo, beforeHi := lo, hi
	if !truth {
		switch op {
		case syntax.TLess:
			op = syntax.TIsGreaterOrEqual
		case syntax.TGreater:
			op = syntax.TIsSmallerOrEqual
		case syntax.TIsSmallerOrEqual:
			op = syntax.TGreater
		case syntax.TIsGreaterOrEqual:
			op = syntax.TLess
		default:
			return
		}
	}
	switch op {
	case syntax.TLess:
		hi = min(hi, bound-1)
	case syntax.TGreater:
		if bound == math.MaxInt64 {
			return
		}
		lo = max(lo, bound+1)
	case syntax.TIsSmallerOrEqual:
		hi = min(hi, bound)
	case syntax.TIsGreaterOrEqual:
		lo = max(lo, bound)
	default:
		return
	}
	if lo == beforeLo && hi == beforeHi {
		return
	}
	if lo <= hi {
		v.interval = &intInterval{lo, hi}
		f.setVar(variable.Name, v)
	}
}

// IntEntry returns an exact integer at a read on a concrete, acyclic entry
// prefix. Unknown branches and side effects end this proof. Its environment
// and expression values are separate from the fixed-point analysis.
func (e *Env) IntEntry(x syntax.Expr) (int64, bool) {
	if x == nil {
		return 0, false
	}
	if plainVariable(x) == "" {
		return 0, false
	}
	scope := syntax.EnclosingVariableScope(x)
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.ByRef {
			return 0, false
		}
	}
	normal := e.scope(scope)
	if value := normal.values[x]; !normal.complete || value.Expr == nil && value.interval == nil && value.origins == nil {
		return 0, false
	}
	twin := *e
	twin.scopes = make(map[syntax.Node]*scopeResult, len(e.scopes))
	for key, value := range e.scopes {
		twin.scopes[key] = value
	}
	// The replay visits an acyclic prefix of the already complete scope.
	// Its cloned memo retains normal callee results, and it cannot require
	// more transfer work than that completed analysis.
	r := twin.followMode(syntax.EnclosingVariableScope(x), nil, true, x)
	value := r.values[x]
	lo, hi, known := integerBounds(value)
	return lo, value.Complete && known && lo == hi
}

func (g *graph) entryPrefix(at int, f frame, r *scopeResult, target syntax.Expr) {
	seen := make([]bool, len(g.blocks))
	for steps := 0; at > 0 && steps < MaxDepth*32; steps++ {
		if seen[at] {
			return
		}
		seen[at] = true
		b := g.blocks[at]
		nodes := b.nodes
		if b.node != nil {
			nodes = []syntax.Node{b.node}
		}
		for j := len(nodes) - 1; j >= 0; j-- {
			node := nodes[j]
			contains := false
			syntax.Inspect(node, func(n syntax.Node) bool {
				if n == target {
					contains = true
				}
				return !contains
			})
			if !entrySafe(node, target, contains, f) {
				return
			}
			g.env.transfer(node, &f, r)
			if contains {
				return
			}
		}
		if len(b.next) == 1 {
			at = b.next[0]
			continue
		}
		condition, ok := b.node.(syntax.Expr)
		truth, known := entryTruth(condition, r)
		if !ok || !known || len(b.next) != 2 {
			return
		}
		branch := 1
		if truth {
			branch = 0
		}
		g.env.refine(condition, &f, truth)
		at = b.next[branch]
	}
}

// Only scalar local operations are replayed before the read. A call containing
// the requested argument is terminal, and no other call may occur in it.
func entrySafe(node syntax.Node, target syntax.Expr, contains bool, f frame) bool {
	safe := true
	syntax.Inspect(node, func(n syntax.Node) bool {
		if !safe {
			return false
		}
		switch n := n.(type) {
		case *syntax.ExprStmt, *syntax.Paren, *syntax.Literal, *syntax.Variable, *syntax.Arg, *syntax.ArgList, *syntax.Name:
			if v, ok := n.(*syntax.Variable); ok && (v.NameExpr != nil || external(v.Name)) {
				safe = false
			}
		case *syntax.Return:
			safe = contains
		case *syntax.Echo:
			// Earlier echo operands can invoke an output handler before this read.
			first := false
			syntax.Inspect(n.Exprs[0], func(child syntax.Node) bool {
				if child == target {
					first = true
				}
				return !first
			})
			safe = first
		case *syntax.Assign:
			name := plainVariable(n.Var)
			old, held := f.vars[name]
			safe = name != "" && !n.ByRef && f.aliased[name] == 0 && (held && old.plain && old.deep || !held && f.definite)
		case *syntax.IncDec:
			safe = plainVariable(n.Var) != "" && f.aliased[plainVariable(n.Var)] == 0
		case *syntax.Binary:
			switch n.Op.Kind {
			case syntax.TPlus, syntax.TMinus, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
			default:
				safe = false
			}
		case *syntax.Unary:
			safe = n.Op.Kind == syntax.TPlus || n.Op.Kind == syntax.TMinus || n.Op.Kind == syntax.TExclaim
		case *syntax.ConstFetch:
			safe = constantName(n) != ""
		case *syntax.FuncCall:
			// The target must be a direct argument expression of this call.
			found := false
			if contains && n.Args != nil {
				for _, arg := range n.Args.Args {
					value := arg
					if named, ok := arg.(*syntax.Arg); ok {
						value = named.Value
					}
					if value == target {
						found = true
					}
				}
			}
			safe = found
		default:
			safe = false
		}
		return safe
	})
	return safe
}

func entryTruth(x syntax.Expr, r *scopeResult) (bool, bool) {
	if n, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch); ok {
		switch constantName(n) {
		case "true":
			return true, true
		case "false", "null":
			return false, true
		}
	}
	if n, ok := syntax.UnwrapParens(x).(*syntax.Binary); ok {
		lo, hi, a := integerBounds(r.values[n.Left])
		bl, bh, b := integerBounds(r.values[n.Right])
		if a && b && lo == hi && bl == bh {
			switch n.Op.Kind {
			case syntax.TLess:
				return lo < bl, true
			case syntax.TGreater:
				return lo > bl, true
			case syntax.TIsSmallerOrEqual:
				return lo <= bl, true
			case syntax.TIsGreaterOrEqual:
				return lo >= bl, true
			}
		}
	}
	lo, hi, known := integerBounds(r.values[x])
	return lo != 0, known && lo == hi
}

func (e *Env) constantInt(n *syntax.ConstFetch) *intInterval {
	if n.Name == nil || constantName(n) != "" {
		return nil
	}
	fqn, fallback := e.types.Names.Const(n.Name.Value, n.Span().Start)
	c := e.types.Index.Constant(fqn, e.types.PHP)
	if c == nil && fallback != "" {
		c = e.types.Index.Constant(fallback, e.types.PHP)
	}
	if c == nil || !c.Avail.In(e.types.PHP) {
		return nil
	}
	value, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		return nil
	}
	return &intInterval{value, value}
}

func (e *Env) classConstantInt(n *syntax.ClassConstFetch) *intInterval {
	class, ok := n.Class.(*syntax.Name)
	name, named := n.Name.(*syntax.Identifier)
	if !ok || !named || strings.EqualFold(class.Value, "static") {
		return nil
	}
	c := e.types.Index.FindConst(e.named(class), name.Value, e.types.PHP)
	if c == nil || c.Case {
		return nil
	}
	value, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil {
		return nil
	}
	return &intInterval{value, value}
}
