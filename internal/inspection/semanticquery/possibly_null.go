package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/types"
)

// Possibly-null values: a value native declarations prove may be null
// (see infer.Env.ProvenTypeOf), dereferenced or passed where null fails.
// The PossiblyNullAccess and PossiblyNullArgument rules report them, and
// NullPointerException leaves what they report to them when they run.

// NullAccess is a member or element access on a value that may be null.
type NullAccess struct {
	// Receiver is the value accessed: the object of `->`, the class
	// expression of `::`, the array of `[]`.
	Receiver syntax.Expr
	// Type is the receiver's proven type; it includes null.
	Type types.Type
}

// Message is the finding's text.
func (a NullAccess) Message() string {
	return "Check this value for null before the access: it is " + typeName(a.Type) + "."
}

type nullAccessKey struct{ access syntax.Node }

// PossiblyNullAccess reports access n (a property fetch, method call,
// static access through an expression, or array element read) when its
// receiver may be null (see possiblyNull) and the access fails or warns on
// null: not `?->`, not a read inside isset(), empty() or the left side of
// `??`, not an element write (which turns null into an array), not an
// element read before PHP 7.4 (which read null silently). Cached per node.
func PossiblyNullAccess(ctx *analysis.Context, n syntax.Node) (NullAccess, bool) {
	r := ctx.Shared(nullAccessKey{n}, func() any {
		a, ok := possiblyNullAccess(ctx, n)
		if !ok {
			return nil
		}
		return a
	})
	a, ok := r.(NullAccess)
	return a, ok
}

func possiblyNullAccess(ctx *analysis.Context, n syntax.Node) (NullAccess, bool) {
	var recv syntax.Expr
	switch a := n.(type) {
	case *syntax.PropertyFetch:
		if a.NullSafe || quietFetch(a) {
			return NullAccess{}, false
		}
		recv = a.Var
	case *syntax.MethodCall:
		if a.NullSafe {
			return NullAccess{}, false
		}
		recv = a.Var
	case *syntax.StaticCall:
		recv = a.Class
	case *syntax.ClassConstFetch:
		recv = a.Class
	case *syntax.StaticPropertyFetch:
		if quietFetch(a) {
			return NullAccess{}, false
		}
		recv = a.Class
	case *syntax.ArrayDimFetch:
		if ctx.PHP < phpversion.PHP74 || a.Dim == nil || quietFetch(a) || elementWrite(ctx, a) {
			return NullAccess{}, false
		}
		recv = a.Var
	default:
		return NullAccess{}, false
	}
	if _, ok := recv.(*syntax.Name); ok || recv == nil || inNullsafeChain(recv) {
		return NullAccess{}, false
	}
	t, ok := possiblyNull(ctx, recv)
	if !ok {
		return NullAccess{}, false
	}
	return NullAccess{Receiver: recv, Type: t}, true
}

type assertingEnvKey struct{}

// possiblyNull returns the proven type of x when it includes null and every
// other member is known (no mixed), unless the code rules null out in a way
// the proven type does not count: PHPDoc (an assertion method, an inline
// @var), an assert() call, trusted here whatever the configuration says
// since the code states that null is not expected, a check of a property
// that a later call may have undone, or a check of an identical call made
// earlier (see checkedCall). False for `$this` in a closure, which may run
// bound to another object.
func possiblyNull(ctx *analysis.Context, x syntax.Expr) (types.Type, bool) {
	if reboundThis(x) {
		return types.Unknown, false
	}
	if t := ctx.TypeOf(x); !t.Has("null") || t.Has("mixed") {
		return types.Unknown, false
	}
	t := ctx.Types().ProvenTypeOf(x)
	if !t.Has("null") || t.Has("mixed") {
		return types.Unknown, false
	}
	asserting := ctx.Shared(assertingEnvKey{}, func() any {
		return ctx.Types().WithOptions(infer.Options{Assertions: true, PropertyFactsSurviveCalls: true})
	}).(*infer.Env)
	if !asserting.TypeOf(x).Has("null") || checkedCall(ctx, x, t) {
		return types.Unknown, false
	}
	return t, true
}

// checkedCall reports whether x is a call and the conditions guarding it
// rule null out of the result of an identical call made earlier in its
// scope (`if ($p->date() !== null) { $p->date()->format('c'); }`): the
// code expects the method to answer the same again. A local counts when
// every value reaching it is such a call or never null (see checkedLocal).
func checkedCall(ctx *analysis.Context, x syntax.Expr, t types.Type) bool {
	switch v := x.(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
	case *syntax.Variable:
		return checkedLocal(ctx, v, t)
	default:
		return false
	}
	return NativeEarlierCallProves(ctx, x, func(n syntax.Expr) (types.Type, func(types.Type) bool) {
		if !astquery.Equivalent(ctx.File, n, x) {
			return types.Unknown, nil
		}
		return t, func(r types.Type) bool { return !r.Has("null") }
	})
}

// checkedLocal reports whether every value reaching local v is a checked
// call (see checkedCall) or proven never null, at least one being such a
// call (`while ($e->getPrevious()) { $e = $e->getPrevious(); }`). A value
// from the function's start must be a parameter whose type excludes null.
func checkedLocal(ctx *analysis.Context, v *syntax.Variable, t types.Type) bool {
	scope := syntax.EnclosingVariableScope(v)
	defs, entry := flowquery.ReachingAssignmentsIn(ctx.File, scope, v, v.Name)
	if entry && !nonNullParam(ctx, scope, v.Name) {
		return false
	}
	calls := 0
	for _, d := range defs {
		switch c := d.Value.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			if checkedCall(ctx, c, t) {
				calls++
				continue
			}
		}
		if dt := ctx.Types().ProvenTypeOf(d.Value); dt.IsUnknown() || MayHoldAtom(dt, "null") {
			return false
		}
	}
	return calls > 0
}

// nonNullParam reports whether scope declares parameter name with a type
// proven never null.
func nonNullParam(ctx *analysis.Context, scope syntax.Node, name string) bool {
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name == name {
			pt := ctx.Types().ProvenTypeOf(p.Var)
			return !p.Variadic && !pt.IsUnknown() && !MayHoldAtom(pt, "null")
		}
	}
	return false
}

// inNullsafeChain reports whether x is an access chain holding a nullsafe
// access: when that one meets null, the rest of the chain, the access on x
// included, is skipped, so null in x may come from there.
func inNullsafeChain(x syntax.Expr) bool {
	for {
		switch a := x.(type) {
		case *syntax.PropertyFetch:
			if a.NullSafe {
				return true
			}
			x = a.Var
		case *syntax.MethodCall:
			if a.NullSafe {
				return true
			}
			x = a.Var
		case *syntax.ArrayDimFetch:
			x = a.Var
		case *syntax.StaticPropertyFetch:
			x = a.Class
		default:
			return false
		}
	}
}

// fetchTop climbs from fetch x through the property and element fetches
// it is the base of (`$x->a['k']->b`): a method call ends the climb.
func fetchTop(x syntax.Expr) syntax.Expr {
	for {
		switch p := x.Parent().(type) {
		case *syntax.PropertyFetch:
			if p.Var != x {
				return x
			}
			x = p
		case *syntax.ArrayDimFetch:
			if p.Var != x {
				return x
			}
			x = p
		default:
			return x
		}
	}
}

// quietFetch reports whether fetch x is read where null raises nothing:
// inside isset() or empty(), or on the left of `??` or `??=`.
func quietFetch(x syntax.Expr) bool {
	top := fetchTop(x)
	switch p := top.Parent().(type) {
	case *syntax.Isset:
		return true
	case *syntax.Empty:
		return true
	case *syntax.Binary:
		return p.Left == top && p.Op.Kind == syntax.TCoalesce
	case *syntax.Assign:
		return p.Var == top && p.Op.Kind == syntax.TCoalesceEqual
	}
	return false
}

// elementWrite reports whether element access x is, or is the base of, an
// element written to (assigned, destructured into, iterated into,
// incremented, unset, taken by reference): writing an element of null
// makes it an array.
func elementWrite(ctx *analysis.Context, x *syntax.ArrayDimFetch) bool {
	var top syntax.Expr = x
	for {
		p, ok := top.Parent().(*syntax.ArrayDimFetch)
		if !ok || p.Var != top {
			break
		}
		top = p
	}
	switch p := top.Parent().(type) {
	case *syntax.Assign:
		return p.Var == top
	case *syntax.IncDec:
		return true
	case *syntax.ArrayItem:
		return p.ByRef || inDestructuring(p)
	case *syntax.Foreach:
		return p.Expr != top
	case *syntax.Unset:
		return true
	case *syntax.Arg:
		return mayBindByRef(ctx, p)
	}
	return false
}

// inDestructuring reports whether item belongs to a list() or `[...]`
// assignment target, at any depth.
func inDestructuring(item *syntax.ArrayItem) bool {
	var x syntax.Node = item
	for {
		switch p := x.Parent().(type) {
		case *syntax.List:
			return true
		case *syntax.Array:
			x = p
		case *syntax.ArrayItem:
			x = p
		case *syntax.Assign:
			return p.Var == x
		case *syntax.Foreach:
			return p.Expr != x
		default:
			return false
		}
	}
}

// mayBindByRef reports whether argument arg may bind by reference: its
// call cannot be resolved, or binds it (or may, behind an unpacked
// argument) to a by-reference parameter.
func mayBindByRef(ctx *analysis.Context, arg *syntax.Arg) bool {
	list := arg.Parent().(*syntax.ArgList) // an argument's parent is its list
	callee, ok := ResolveCallee(ctx, list.Parent())
	if !ok {
		return true
	}
	byRef, known := ArgumentBindsByRef(list, arg, callee.Params, true)
	return byRef || !known
}

// NullArgument is a possibly-null argument passed to a parameter that
// rejects null.
type NullArgument struct {
	Arg *syntax.Arg
	// Callee names the nominal declaration (`f`, `Class::method`).
	Callee string
	// Param is the name of the parameter of the nominal declaration.
	Param string
	// Type is the argument's proven type; it includes null.
	Type types.Type
}

// Message is the finding's text.
func (a NullArgument) Message() string {
	return "Check this value for null before passing it to $" + a.Param + " of " + a.Callee + "(): it is " + typeName(a.Type) + "."
}

type nullArgumentsKey struct{ call syntax.Node }

// PossiblyNullArguments lists the arguments of call that may be null (see
// possiblyNull), bound (as ArgumentTypeMismatches binds them) to a
// parameter whose native type rejects null in every declaration the call
// may bind to (see CallTargets). An argument the parameter rejects
// whatever its value (types.AllRejected, an ArgumentTypeMismatches case)
// is left out, and so is null passed to a scalar parameter of a builtin
// outside strict mode, which PHP only deprecates. Cached per call.
func PossiblyNullArguments(ctx *analysis.Context, call syntax.Node) []NullArgument {
	return ctx.Shared(nullArgumentsKey{call}, func() any { return possiblyNullArguments(ctx, call) }).([]NullArgument)
}

func possiblyNullArguments(ctx *analysis.Context, call syntax.Node) []NullArgument {
	if reboundReceiver(call) {
		return nil
	}
	type candidate struct {
		b boundArg
		t types.Type
	}
	var cands []candidate
	for _, b := range bindableArgs(call) {
		if t, ok := possiblyNull(ctx, b.arg.Value); ok {
			cands = append(cands, candidate{b, t})
		}
	}
	if len(cands) == 0 {
		return nil
	}
	callee, ok := ResolveCallee(ctx, call)
	if !ok {
		return nil
	}
	label, targets, ok := CallTargets(ctx, call, callee, func(file, name string) bool {
		return TrustedBuiltinTypes(file, name, ctx.PHP)
	})
	if !ok {
		return nil
	}
	check := fileTypeCheck(ctx)
	var out []NullArgument
	for _, c := range cands {
		params, ok := bindAll(targets, c.b)
		if !ok || !check.nativeRejects(targets, params, types.Of("null")) || check.nativeRejects(targets, params, c.t) {
			continue
		}
		out = append(out, NullArgument{Arg: c.b.arg, Callee: label, Param: params[0].Name, Type: c.t})
	}
	return out
}
