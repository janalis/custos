package infer

import (
	"custos/internal/syntax"
	"custos/internal/types"
)

// nullsafeFact separates a skipped chain from the value produced when its
// members execute. A genuine nullable member return must remain in evaluated.
type nullsafeFact struct {
	evaluated types.Type
	skipped   bool
	halted    bool
}

func chainParent(x syntax.Expr) (syntax.Expr, bool) {
	switch n := x.(type) {
	case *syntax.PropertyFetch:
		return n.Var, n.NullSafe
	case *syntax.MethodCall:
		return n.Var, n.NullSafe
	case *syntax.ArrayDimFetch:
		return n.Var, false
	case *syntax.StaticCall:
		return n.Class, false
	case *syntax.StaticPropertyFetch:
		return n.Class, false
	case *syntax.ClassConstFetch:
		return n.Class, false
	case *syntax.Paren:
		return n.Expr, false
	}
	return nil, false
}

// inferNullsafe processes only the direct receiver edge; memoized facts keep
// a chain linear in its length. Arguments and computed names are independent.
func (e *Env) inferNullsafe(x syntax.Expr) types.Type {
	receiver, safe := chainParent(x)
	if receiver == nil {
		return e.infer(x)
	}
	if !safe {
		// A terminal receiver cannot carry a skipped chain. In particular,
		// do not prime variable reads: dimType intentionally asks baseType
		// for the value before element-level narrowing and writes.
		terminal := receiver
		if _, dim := x.(*syntax.ArrayDimFetch); dim {
			terminal = syntax.UnwrapParens(receiver)
		}
		if parent, _ := chainParent(terminal); parent == nil {
			return e.infer(x)
		}
	}
	rt := e.TypeOf(receiver)
	previous := e.chains[receiver]
	halted := previous.halted
	skipped := previous.skipped
	if safe && !rt.IsUnknown() {
		if previous.skipped {
			rt = previous.evaluated
		}
		skipped = skipped || rt.IsNullable()
		halted = halted || rt.OnlyOf("null")
	}
	if !skipped {
		return e.infer(x)
	}
	if e.chains == nil {
		e.chains = map[syntax.Expr]nullsafeFact{}
	}
	t := types.Unknown
	if !halted {
		t = e.infer(x)
	}
	e.chains[x] = nullsafeFact{evaluated: t, skipped: true, halted: halted}
	if halted {
		return types.Null
	}
	t = types.Union(t, types.Null)
	if key := narrowKey(x); key != "" && !t.IsUnknown() {
		t = e.narrowExpr(t, x, key, syntax.EnclosingVariableScope(x))
	}
	return t
}

func (e *Env) chainReceiver(x syntax.Expr, safe bool) types.Type {
	t := e.TypeOf(x)
	if fact, ok := e.chains[x]; ok {
		t = fact.evaluated
	}
	if safe {
		t = t.Without("null")
	}
	return t
}

// ChainTypeOf returns the type produced when x executes, excluding null
// contributed by a skipped nullsafe chain. A completely skipped expression
// has no evaluated value and returns Unknown, false. Nullable member returns remain.
func (e *Env) ChainTypeOf(x syntax.Expr) (types.Type, bool) {
	t := e.TypeOf(x)
	if fact, ok := e.chains[x]; ok {
		return fact.evaluated, !fact.halted
	}
	return t, true
}
