package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// htmlQuoteSafety keeps PHP's two independently controlled quote conversions.
// Omitted flags changed from ENT_COMPAT to ENT_QUOTES in PHP 8.1. Absence
// is supplied by the validated argument binding, not inferred from Expr=nil.
func (e *Env) htmlQuoteSafety(flags Value, omitted bool, r *scopeResult) Context {
	if !omitted && !flags.Complete {
		return 0
	}
	bits := int64(3)
	if omitted {
		if e.types.PHP != 0 && e.types.PHP < phpversion.PHP81 {
			bits = 2
		}
	} else {
		var known bool
		bits, known = e.htmlFlagBits(flags.Expr, r, 16)
		if !known {
			return 0
		}
	}
	var safe Context
	if bits&1 != 0 {
		safe |= HTMLSingleQuoted
	}
	if bits&2 != 0 {
		safe |= HTMLDoubleQuoted
	}
	return safe
}

func (e *Env) htmlFlagBits(x syntax.Expr, r *scopeResult, budget int) (int64, bool) {
	if budget == 0 {
		r.gap("flow-budget", x.Span())
		return 0, false
	}
	if v, ok := r.values[x]; ok && v.Complete && v.Expr != nil && v.Expr != x {
		return e.htmlFlagBits(v.Expr, r, budget-1)
	}
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Binary:
		if n.Op.Kind == syntax.TBar {
			a, ak := e.htmlFlagBits(n.Left, r, budget-1)
			b, bk := e.htmlFlagBits(n.Right, r, budget-1)
			return a | b, ak && bk
		}
	case *syntax.ConstFetch:
		fqn, fallback := e.types.Names.Const(n.Name.Value, n.Name.Span().Start)
		k := e.types.Index.Constant(fqn, e.types.PHP)
		if k == nil && fallback != "" {
			k = e.types.Index.Constant(fallback, e.types.PHP)
		}
		if k != nil {
			v, err := strconv.ParseInt(k.Value, 0, 64)
			return v, err == nil
		}
	default:
		return literalInteger(x)
	}
	return 0, false
}
