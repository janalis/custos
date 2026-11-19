package flow

import (
	"math"
	"strconv"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
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

func integerBounds(v Value) (int64, int64, bool) { return bounds(v.interval(), v.Expr) }

// bounds returns the bounds of interval, or else those of a literal expr.
func bounds(interval *intInterval, expr syntax.Expr) (int64, int64, bool) {
	if interval != nil {
		return interval.lo, interval.hi, true
	}
	n, ok := literalInteger(expr)
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

func integerValue(n int64) Value {
	return Value{Complete: true, ext: rare(valueExt{interval: &intInterval{n, n}})}
}
func sameInts(a, b *intInterval) bool { return a == b || a != nil && b != nil && *a == *b }
func joinInts(a, b Value) *intInterval {
	if a.Expr == b.Expr && a.interval() == nil && b.interval() == nil {
		return nil
	}
	lo, hi, ok := integerBounds(a)
	bl, bh, bok := integerBounds(b)
	if !ok || !bok {
		return nil
	}
	lo, hi = min(lo, bl), max(hi, bh)
	if a.interval() != nil && a.interval().lo == lo && a.interval().hi == hi {
		return a.interval()
	}
	if b.interval() != nil && b.interval().lo == lo && b.interval().hi == hi {
		return b.interval()
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
		if v.interval() != nil {
			return v.interval()
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
// interval is that of the join, which holds expr; each bound past the one
// of old widens to the extremum.
func widenInterval(old Value, interval *intInterval, expr syntax.Expr) *intInterval {
	lo, hi, known := bounds(interval, expr)
	beforeLo, beforeHi, beforeKnown := integerBounds(old)
	if !known || !beforeKnown || lo == beforeLo && hi == beforeHi {
		return interval
	}
	if lo < beforeLo {
		lo = math.MinInt64
	}
	if hi > beforeHi {
		hi = math.MaxInt64
	}
	if interval != nil && interval.lo == lo && interval.hi == hi {
		return interval
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
	v := f.vars.value(variable.Name)
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
		v.setInterval(&intInterval{lo, hi})
		f.setVar(variable.Name, v)
	}
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
