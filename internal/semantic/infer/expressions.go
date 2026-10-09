package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// TypeOf returns the inferred type of x (types.Unknown when it cannot be determined).
func (e *Env) TypeOf(x syntax.Expr) types.Type {
	if x == nil {
		return types.Unknown
	}
	if t, ok := e.cache[x]; ok {
		return t
	}
	if e.busy[x] {
		return types.Unknown // recursion (e.g. $a = $a + 1)
	}
	e.busy[x] = true
	t := e.inferNullsafe(x)
	delete(e.busy, x)
	e.cache[x] = t
	return t
}

// isFirstClassCallable reports `f(...)`-style argument lists.
func isFirstClassCallable(a *syntax.ArgList) bool {
	if a == nil || len(a.Args) != 1 {
		return false
	}
	_, ok := a.Args[0].(*syntax.VariadicPlaceholder)
	return ok
}

func (e *Env) infer(x syntax.Expr) types.Type {
	switch n := x.(type) {
	case *syntax.FuncCall:
		if isFirstClassCallable(n.Args) {
			return e.firstClassCallableType(n)
		}
	case *syntax.MethodCall:
		if isFirstClassCallable(n.Args) {
			return e.firstClassCallableType(n)
		}
	case *syntax.StaticCall:
		if isFirstClassCallable(n.Args) {
			return e.firstClassCallableType(n)
		}
	}
	switch n := x.(type) {
	case *syntax.Literal:
		switch n.LitKind {
		case syntax.LitInt:
			return types.Int
		case syntax.LitFloat:
			return types.Float
		}
		return types.String
	case *syntax.InterpolatedString:
		if n.Backtick {
			return types.Of("string", "null", "false")
		}
		return types.String
	case *syntax.MagicConst:
		if n.Token.Kind == syntax.TLine {
			return types.Int
		}
		return types.String
	case *syntax.Paren:
		return e.chainReceiver(n.Expr, false)
	case *syntax.ConstFetch:
		return e.constType(n)
	case *syntax.Array:
		return e.arrayType(n)
	case *syntax.Unary:
		return e.unaryType(n)
	case *syntax.Binary:
		return e.binaryType(n)
	case *syntax.Assign:
		if n.Op.Kind == syntax.TEqual {
			return e.TypeOf(n.Value)
		}
		return e.compoundType(n)
	case *syntax.Ternary:
		then := n.Then
		if then == nil {
			return types.Union(e.TypeOf(n.Cond).Without("null", "false"), e.TypeOf(n.Else))
		}
		return types.Union(e.TypeOf(then), e.TypeOf(n.Else))
	case *syntax.Instanceof, *syntax.Isset, *syntax.Empty:
		return types.Bool
	case *syntax.Print:
		return types.Int
	case *syntax.Yield:
		return e.yieldType(n)
	case *syntax.YieldFrom:
		return e.yieldFromType(n)
	case *syntax.Exit:
		return e.exitType(n)
	case *syntax.Throw:
		return types.Of("never")
	case *syntax.Clone:
		return e.cloneType(n)
	case *syntax.Closure, *syntax.ArrowFunction:
		return e.closureType(n)
	case *syntax.IncDec:
		if n.Prefix {
			return e.incDecStored(n)
		}
		return e.TypeOf(n.Var)
	case *syntax.Match:
		var ts []types.Type
		for _, a := range n.Arms {
			ts = append(ts, e.TypeOf(a.Body))
		}
		if len(ts) == 0 {
			return types.Unknown
		}
		return types.Union(ts...)
	case *syntax.New:
		return e.newType(n)
	case *syntax.Variable:
		return e.variableType(n)
	case *syntax.FuncCall:
		return e.funcCallType(n)
	case *syntax.MethodCall:
		return e.methodCallType(n)
	case *syntax.StaticCall:
		return e.staticCallType(n)
	case *syntax.PropertyFetch:
		t := e.propertyType(e.chainReceiver(n.Var, n.NullSafe), n.Name, false)
		if key := narrowKey(n); key != "" {
			t = e.narrowExpr(t, n, key, syntax.EnclosingVariableScope(n))
		}
		return t
	case *syntax.StaticPropertyFetch:
		cls := e.classRef(n.Class)
		if cls == "" {
			return types.Unknown
		}
		v, ok := n.Name.(*syntax.Variable)
		if !ok || v.Name == "" {
			return types.Unknown
		}
		if p := e.Index.FindProperty(cls, v.Name, e.PHP); p != nil {
			t := e.propType(p, cls)
			if key := narrowKey(n); key != "" {
				t = e.narrowExpr(t, n, key, syntax.EnclosingVariableScope(n))
			}
			return t
		}
		return types.Unknown
	case *syntax.ClassConstFetch:
		return e.classConstType(n)
	case *syntax.ArrayDimFetch:
		t := e.dimType(n)
		if key := narrowKey(n); key != "" && !t.IsUnknown() {
			t = e.narrowExpr(t, n, key, syntax.EnclosingVariableScope(n))
		}
		return t
	}
	return types.Unknown
}
