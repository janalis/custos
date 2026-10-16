package flow

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

// constantScalar deliberately omits PHP's version-dependent mixed coercions.
// k is i(integer), f(float), s(string), b(boolean), or n(null).
type constantScalar struct {
	k byte
	i int64
	f float64
	s string
	b bool
}

func constantTruth(x syntax.Expr) (bool, bool) {
	fuel := 256
	v, known := constantValue(x, MaxDepth*8, &fuel)
	return v.truth(), known
}

func (v constantScalar) truth() bool {
	switch v.k {
	case 'i':
		return v.i != 0
	case 'f':
		return v.f != 0
	case 's':
		return v.s != "" && v.s != "0"
	case 'b':
		return v.b
	}
	return false
}

func constantValue(x syntax.Expr, budget int, fuel *int) (constantScalar, bool) {
	if budget == 0 || *fuel == 0 {
		return constantScalar{}, false
	}
	(*fuel)--
	switch n := x.(type) {
	case *syntax.Paren:
		return constantValue(n.Expr, budget-1, fuel)
	case *syntax.Literal:
		if len(n.Raw) > 4096 {
			return constantScalar{}, false
		}
		switch n.LitKind {
		case syntax.LitInt:
			v, ok := literalInteger(n)
			return constantScalar{k: 'i', i: v}, ok
		case syntax.LitFloat:
			v, err := strconv.ParseFloat(strings.ReplaceAll(n.Raw, "_", ""), 64)
			return constantScalar{k: 'f', f: v}, err == nil && !math.IsInf(v, 0) && !math.IsNaN(v)
		case syntax.LitString:
			v, ok := constantString(n)
			return constantScalar{k: 's', s: v}, ok
		}
	case *syntax.ConstFetch:
		switch constantName(n) {
		case "true":
			return constantScalar{k: 'b', b: true}, true
		case "false":
			return constantScalar{k: 'b'}, true
		case "null":
			return constantScalar{k: 'n'}, true
		}
	case *syntax.Unary:
		v, known := constantValue(n.Expr, budget-1, fuel)
		if !known {
			return constantScalar{}, false
		}
		switch n.Op.Kind {
		case syntax.TExclaim:
			return constantScalar{k: 'b', b: !v.truth()}, true
		case syntax.TPlus:
			return v, v.k == 'i' || v.k == 'f'
		case syntax.TMinus:
			if v.k == 'i' && v.i != math.MinInt64 {
				v.i = -v.i
				return v, true
			}
			if v.k == 'f' {
				v.f = -v.f
				return v, true
			}
		}
	case *syntax.Binary:
		a, ak := constantValue(n.Left, budget-1, fuel)
		b, bk := constantValue(n.Right, budget-1, fuel)
		if !ak || !bk {
			return constantScalar{}, false
		}
		return constantBinary(n.Op.Kind, a, b)
	}
	return constantScalar{}, false
}

// strconv.Unquote accepts Go-only escapes (notably \a, \b and \uNNNN).
// Reject those before using the shared literal decoder. Unsupported PHP
// escapes stay unknown; a restricted decoder must never change their value.
func constantString(n *syntax.Literal) (string, bool) {
	if n.Raw != "" && n.Raw[0] == '"' {
		for i := 1; i < len(n.Raw)-1; i++ {
			if n.Raw[i] != '\\' {
				continue
			}
			i++
			switch n.Raw[i] {
			case '\\', '"', 'n', 'r', 't', 'v', 'f', 'x', '0', '1', '2', '3', '4', '5', '6', '7':
			default:
				return "", false
			}
		}
	}
	return literalText(n)
}

func constantBinary(op syntax.TokenKind, a, b constantScalar) (constantScalar, bool) {
	boolean := func(v bool) (constantScalar, bool) { return constantScalar{k: 'b', b: v}, true }
	switch op {
	case syntax.TBooleanAnd, syntax.TAnd:
		return boolean(a.truth() && b.truth())
	case syntax.TBooleanOr, syntax.TOr:
		return boolean(a.truth() || b.truth())
	case syntax.TXor:
		return boolean(a.truth() != b.truth())
	case syntax.TIsIdentical, syntax.TIsNotIdentical:
		same := a == b
		return boolean(same == (op == syntax.TIsIdentical))
	case syntax.TDot:
		if a.k == 's' && b.k == 's' && len(a.s)+len(b.s) <= 4096 {
			return constantScalar{k: 's', s: a.s + b.s}, true
		}
	}
	if a.k != b.k {
		return constantScalar{}, false
	}
	if op == syntax.TIsEqual || op == syntax.TIsNotEqual {
		if a.k == 's' {
			return constantScalar{}, false
		}
		return boolean((a == b) == (op == syntax.TIsEqual))
	}
	var cmp int
	switch a.k {
	case 'i':
		switch op {
		case syntax.TPlus, syntax.TMinus, syntax.TMul:
			var v int64
			var ok bool
			switch op {
			case syntax.TPlus:
				v, ok = addInt(a.i, b.i)
			case syntax.TMinus:
				v, ok = subInt(a.i, b.i)
			case syntax.TMul:
				v = a.i * b.i
				ok = a.i == 0 || b.i == 0 || !(a.i == math.MinInt64 && b.i == -1 || b.i == math.MinInt64 && a.i == -1) && v/a.i == b.i
			}
			return constantScalar{k: 'i', i: v}, ok
		}
		if a.i < b.i {
			cmp = -1
		} else if a.i > b.i {
			cmp = 1
		}
	case 'f':
		if a.f < b.f {
			cmp = -1
		} else if a.f > b.f {
			cmp = 1
		}
	default:
		return constantScalar{}, false
	}
	switch op {
	case syntax.TLess:
		return boolean(cmp < 0)
	case syntax.TGreater:
		return boolean(cmp > 0)
	case syntax.TIsSmallerOrEqual:
		return boolean(cmp <= 0)
	case syntax.TIsGreaterOrEqual:
		return boolean(cmp >= 0)
	}
	return constantScalar{}, false
}

// generatorScopesKey caches syntax only; no semantic option or index state.
type generatorScopesKey struct{}

func (e *Env) generator(scope syntax.Node) bool {
	scopes := e.file.Memo(generatorScopesKey{}, func() any {
		scopes := map[syntax.Node]bool{}
		present := false
		for _, token := range e.file.Tokens {
			if token.Kind == syntax.TYield || token.Kind == syntax.TYieldFrom {
				present = true
				break
			}
		}
		if present {
			syntax.InspectFile(e.file, func(n syntax.Node) bool {
				switch n.(type) {
				case *syntax.Yield, *syntax.YieldFrom:
					scopes[syntax.EnclosingVariableScope(n)] = true
				}
				return true
			})
		}
		return scopes
	}).(map[syntax.Node]bool)
	return scopes[scope]
}
