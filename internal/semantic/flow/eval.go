package flow

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func (e *Env) eval(x syntax.Expr, f *frame, r *scopeResult, depth int) Value {
	if x == nil {
		return Value{Complete: true}
	}
	if !r.complete {
		return Value{}
	}
	r.operations++
	if r.operations > MaxTransfers || r.copies > MaxCopied || depth >= MaxDepth*8 {
		r.complete = false
		return Value{}
	}
	remember(r, x, *f)
	if len(r.statics) > 0 && !calm(x) {
		// It may also throw, which ends the function here.
		leaving(f, r)
	}
	var v Value
	switch n := x.(type) {
	case *syntax.Variable:
		v = e.evalVariable(n, f, r, depth)
	case *syntax.Paren:
		v = e.eval(n.Expr, f, r, depth+1)
	case *syntax.Assign:
		v = e.evalAssign(n, f, r, depth)
	case *syntax.FuncCall:
		v = e.evalFuncCall(n, f, r, depth)
	case *syntax.MethodCall:
		v = e.evalMethodCall(n, f, r, depth)
	case *syntax.StaticCall:
		v = e.evalStaticCall(n, f, r, depth)
	case *syntax.New:
		v = e.evalNew(n, f, r, depth)
	case *syntax.Arg:
		v = e.evalArg(n, f, r, depth)
	case *syntax.Binary:
		v = e.evalBinary(n, f, r, depth)
	case *syntax.Unary:
		v = e.evalUnary(n, f, r, depth)
	case *syntax.IncDec:
		v = e.evalIncDec(n, f, r, depth)
	case *syntax.Ternary:
		v = e.evalTernary(n, f, r, depth)
	case *syntax.Match:
		v = e.evalMatch(n, f, r, depth)
	case *syntax.ArrayDimFetch:
		v = e.evalArrayDimFetch(n, f, r, depth)
	case *syntax.Array:
		v = e.evalArray(n, f, r, depth)
	case *syntax.PropertyFetch:
		v = e.evalPropertyFetch(n, f, r, depth)
	case *syntax.InterpolatedString:
		v = e.evalInterpolatedString(n, f, r, depth)
	case *syntax.Closure:
		v = e.evalClosure(n, f, r, depth)
	case *syntax.ArrowFunction:
		v = e.evalArrowFunction(n, f, r)
	case *syntax.Yield, *syntax.YieldFrom:
		e.suspend(x, f, r, depth)
		// What the consumer sends back is not known.
		v = Value{}
	case *syntax.Eval, *syntax.Include:
		v = evalDynamicCode(x, f, r)
	default:
		v = e.evalOther(x, f, r, depth)
	}
	// What the expression yields is no object when it computes a number, a
	// boolean, text, an array or a function; one that hands a value on
	// yields what it read.
	switch n := x.(type) {
	case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.ArrayDimFetch, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall,
		*syntax.PropertyFetch:
	case *syntax.Binary:
		// `+` may unite two arrays, with what they hold.
		if n.Op.Kind != syntax.TCoalesce {
			v.plain, v.deep = true, v.deep || n.Op.Kind != syntax.TPlus
		}
	case *syntax.Unary:
		// An array cast keeps what its operand holds, a text cast its text.
		switch n.Op.Kind {
		case syntax.TAt:
		case syntax.TObjectCast:
			v.plain, v.deep = false, false
		default:
			v.plain, v.deep = true, v.deep || n.Op.Kind != syntax.TArrayCast
			v.matched = v.matched && n.Op.Kind == syntax.TStringCast
		}
	case *syntax.Array:
		v.plain = true
	case *syntax.Literal, *syntax.StringPart, *syntax.InterpolatedString, *syntax.MagicConst:
		v.plain, v.deep = true, true
	case *syntax.ConstFetch:
		v.plain, v.deep = e.fixed(n), e.fixed(n)
	case *syntax.Isset, *syntax.Empty, *syntax.Instanceof, *syntax.Closure, *syntax.ArrowFunction, *syntax.Print:
		v.plain, v.deep, v.matched = true, true, false
	case *syntax.IncDec:
		// Since PHP 8.3 only a number or text can be stepped.
		strict := e.types.PHP.AtLeast(phpversion.PHP83)
		v.plain, v.deep, v.matched = v.plain || strict, v.deep || strict, false
	case *syntax.ClassConstFetch:
		v.plain, v.deep = className(n), className(n)
	default:
		v.plain, v.deep = false, false
	}
	// Shape completeness describes an array value, not a scalar coercion or
	// predicate computed from that array. Only value-preserving forms retain it.
	if v.literal != nil || v.truth != 0 {
		switch x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.ArrayDimFetch, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		default:
			v.literal = nil
			v.truth = 0
		}
	}
	switch x.(type) {
	case *syntax.Binary, *syntax.Unary:
		v = evaluatedScalar(x, v, r)
	}
	if v.elements != nil {
		switch node := x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.Array, *syntax.ArrayDimFetch, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		case *syntax.Unary:
			if node.Op.Kind != syntax.TArrayCast && node.Op.Kind != syntax.TAt {
				v.elements = nil
			}
		case *syntax.Binary:
			if node.Op.Kind != syntax.TPlus && node.Op.Kind != syntax.TCoalesce {
				v.elements = nil
			}
		default:
			v.elements = nil
		}
	}
	if v.targets != nil || v.returned != nil {
		switch x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.Closure, *syntax.ArrowFunction, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		default:
			v.targets = nil
			v.returned = nil
		}
	}
	// Only an expression that hands its operand on unchanged yields the
	// uploads table, one of its entries or an array of numbers.
	if v.upload != 0 || v.scalars {
		switch n := x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.ArrayDimFetch, *syntax.Unary, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.Arg:
		case *syntax.Binary:
			if n.Op.Kind != syntax.TCoalesce {
				v.upload, v.scalars = 0, false
			}
		default:
			v.upload, v.scalars = 0, false
		}
	}
	switch n := x.(type) {
	case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.IncDec, *syntax.Binary, *syntax.Unary:
	case *syntax.ConstFetch:
		v.interval = e.constantInt(n)
	case *syntax.ClassConstFetch:
		v.interval = e.classConstantInt(n)
	default:
		v.interval = nil
	}
	if old, ok := r.values.get(x); ok {
		seen := merge(old, v)
		if _, read := x.(*syntax.Variable); read {
			// Every visit of the read adds the producers that reach it.
			seen.origins = origins(old, v, r)
		}
		r.values.set(x, seen)
	} else {
		r.values.set(x, v)
	}
	return v
}

// chainDepth charges nothing for a fluent link: each `->m()` of a chain wraps
// the one before it, so a long builder chain is deep without being nested
// code. Parsing bounds the tree depth, and operations still bound the work.
func chainDepth(receiver syntax.Expr) int {
	switch receiver.(type) {
	case *syntax.MethodCall, *syntax.PropertyFetch:
		return 0
	}
	return 1
}

// plainVariable names a target that is a variable written as such.
func plainVariable(x syntax.Expr) string {
	if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.NameExpr == nil {
		return v.Name
	}
	return ""
}

// conditional reports a member access PHP may skip: a null-safe link, or
// anything chained after one, is not evaluated when the receiver is null.
func conditional(x syntax.Expr) bool {
	for {
		switch n := syntax.UnwrapParens(x).(type) {
		case *syntax.MethodCall:
			if n.NullSafe {
				return true
			}
			x = n.Var
		case *syntax.PropertyFetch:
			if n.NullSafe {
				return true
			}
			x = n.Var
		case *syntax.ArrayDimFetch:
			x = n.Var
		case *syntax.StaticCall:
			x = n.Class
		case *syntax.StaticPropertyFetch:
			x = n.Class
		case *syntax.ClassConstFetch:
			x = n.Class
		default:
			return false
		}
	}
}

// request constructs the same provenance for every spelling of a request table.
func (e *Env) request(name string, at syntax.Expr, f *frame) Value {
	held, written := f.vars.lookup(name)
	v := Value{
		Complete: true, Expr: at, plain: true, deep: !written || held.deep,
		Sources: []Source{{Kind: name, Path: e.file.Path, Span: at.Span(), Parameter: -1}},
	}
	if name == "_FILES" {
		v.upload = uploadTable
	}
	return v
}

// searchedOnly reports whether argument x, the i-th of a call to builtin
// name, is the text or pattern searched for by a replacing function: none
// of it reaches the result.
func searchedOnly(name string, x syntax.Node, i int) bool {
	var param string
	switch name {
	case "str_replace", "str_ireplace":
		param = "search"
	case "preg_replace":
		param = "pattern"
	default:
		return false
	}
	a, ok := x.(*syntax.Arg)
	if !ok || a.Unpack {
		return false
	}
	if a.Name != nil {
		return a.Name.Value == param
	}
	return i == 0
}

func external(s string) bool {
	switch s {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_FILES", "_SERVER":
		return true
	}
	return false
}
