package infer

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// applyCond narrows t assuming cond evaluates to `truthy`.
func (e *Env) applyCond(t types.Type, cond syntax.Expr, name string, truthy bool) types.Type {
	return e.applyCondB(t, cond, name, truthy, &condBudget{n: maxCondSteps})
}

// condBudget bounds one applyCond evaluation: n condition nodes at most
// (see maxCondSteps); inAlias is set while a boolean alias's condition is
// applied, so aliases of aliases are not followed.
type condBudget struct {
	n       int
	inAlias bool
}

// unionNarrowed is the union of two alternative narrowings of one type; an
// emptied (unknown) side is an impossible path and adds nothing.
func unionNarrowed(a, b types.Type) types.Type {
	switch {
	case a.IsUnknown():
		return b
	case b.IsUnknown():
		return a
	}
	return types.Union(a, b)
}

func (e *Env) applyCondB(t types.Type, cond syntax.Expr, name string, truthy bool, steps *condBudget) types.Type {
	if t.IsUnknown() || steps.n <= 0 {
		return t
	}
	steps.n--
	if callMatches(cond, name) {
		// The result of this very call or comparison (see callKey).
		if truthy {
			return truthyType(t)
		}
		return falsyType(t)
	}
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		switch c.Op.Kind {
		case syntax.TExclaim:
			return e.applyCondB(t, c.Expr, name, !truthy, steps)
		case syntax.TAt:
			return e.applyCondB(t, c.Expr, name, truthy, steps) // silencing keeps the value
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			lt := e.applyCondB(t, c.Left, name, true, steps)
			if truthy {
				return e.applyCondB(lt, c.Right, name, true, steps)
			}
			// !(A && B): A is false, or A is true and B false.
			return unionNarrowed(e.applyCondB(t, c.Left, name, false, steps), e.applyCondB(lt, c.Right, name, false, steps))
		case syntax.TBooleanOr, syntax.TOr:
			lf := e.applyCondB(t, c.Left, name, false, steps)
			if !truthy {
				return e.applyCondB(lf, c.Right, name, false, steps)
			}
			// A || B: A is true, or A is false and B true.
			return unionNarrowed(e.applyCondB(t, c.Left, name, true, steps), e.applyCondB(lf, c.Right, name, true, steps))
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			return e.eqCond(t, c, name, truthy, steps)
		case syntax.TLess, syntax.TIsSmallerOrEqual, syntax.TGreater, syntax.TIsGreaterOrEqual:
			if n, ok := countComparison(c, name); ok {
				return nonEmptyIf(t, countNonEmpty(c.Op.Kind, n, truthy))
			}
			if e.lengthOperand(c.Left, name) {
				if v, ok := e.knownInt(c.Right); ok {
					return lengthCmp(t, c.Op.Kind, v, truthy)
				}
			}
			if e.lengthOperand(c.Right, name) {
				if v, ok := e.knownInt(c.Left); ok {
					return lengthCmp(t, mirrorOrder[c.Op.Kind], v, truthy)
				}
			}
			if _, known := t.IntRanges(); known {
				if v, ok := e.knownInt(c.Right); isVar(c.Left, name) && ok {
					return intOrder(t, c.Op.Kind, v, truthy)
				}
				if v, ok := e.knownInt(c.Left); isVar(c.Right, name) && ok {
					return intOrder(t, mirrorOrder[c.Op.Kind], v, truthy) // `0 < $x` reads `$x > 0`
				}
			}
		}
	case *syntax.Empty:
		if isVar(c.Expr, name) {
			if truthy {
				return falsyType(t)
			}
			return truthyType(t)
		}
		if !truthy {
			return chainBaseNonNull(t, c.Expr, name)
		}
	case *syntax.Instanceof:
		if !isVar(c.Expr, name) {
			if truthy {
				return chainBaseNonNull(t, c.Expr, name)
			}
			return t
		}
		cls := e.classRef(c.Class)
		if cls == "" && truthy {
			// `$x instanceof $cls` with `$cls` a class-string<C>: a C.
			cls = e.classStringRef(c.Class)
		}
		if cls == "" {
			return t
		}
		if truthy {
			return e.instanceOf(t, cls)
		}
		return e.notInstance(t, cls, true)
	case *syntax.Isset:
		if !truthy {
			return t
		}
		for _, x := range c.Vars {
			if isVar(x, name) {
				return t.Without("null")
			}
		}
		for _, x := range c.Vars {
			// `isset($x->p)`, `isset($x['a']['b'])`: the base is set too.
			if r, ok := chainBase(t, x, name); ok {
				return issetOffset(r, x, name)
			}
		}
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch, *syntax.MethodCall:
		if narrowKey(c) == name {
			if truthy {
				return truthyType(t)
			}
			return falsyType(t)
		}
		if v, ok := c.(*syntax.Variable); ok {
			// `$isObject = is_object($r); if ($isObject) { … }`
			if !steps.inAlias {
				if ac := e.aliasCond(v, name); ac != nil {
					steps.inAlias = true // one level: the alias's condition is read as is
					r := e.applyCondB(t, ac, name, truthy, steps)
					steps.inAlias = false
					return r
				}
			}
			return t
		}
		if mc, ok := c.(*syntax.MethodCall); ok {
			if r, ok := e.condAsserts(mc, name, truthy, t); ok {
				return r
			}
		}
		if truthy {
			// `$x?->isReady()`, `$x->items[0]`: a null (or, for an element,
			// false) base would make the whole chain null.
			return chainBaseNonNull(t, c, name)
		}
	case *syntax.Assign:
		// `while ($job = $q->next())`, `if (!$x = f())`: the assigned
		// variable's truthiness.
		if c.Op.Kind == syntax.TEqual && !c.ByRef && narrowKey(c.Var) == name {
			return e.applyCondB(t, c.Var, name, truthy, steps)
		}
		if c.Op.Kind == syntax.TEqual && !c.ByRef && callMatches(c.Value, name) {
			return e.applyCondB(t, c.Value, name, truthy, steps)
		}
		if v, ok := c.Var.(*syntax.Variable); ok && c.Op.Kind == syntax.TEqual && !c.ByRef && v.NameExpr == nil && !keyUnder(name, v.Name) {
			// `if ($neg = str_starts_with($line, '!'))` tests what it stores.
			return e.applyCondB(t, c.Value, name, truthy, steps)
		}
	case *syntax.StaticCall:
		if r, ok := e.condAsserts(c, name, truthy, t); ok {
			return r
		}
	case *syntax.FuncCall:
		if truthy && e.needleFound(c, name) {
			return t.Without("null", "false")
		}
		if r, ok := e.inListCond(t, c, name, truthy); ok {
			return r
		}
		if r, ok := e.condAsserts(c, name, truthy, t); ok {
			return r
		}
		nm, ok := c.Name.(*syntax.Name)
		if !ok || c.Args == nil || len(c.Args.Args) == 0 {
			return t
		}
		if r, ok := e.keyExistsCond(t, c, nm, name, truthy); ok {
			return r
		}
		a, ok := c.Args.Args[0].(*syntax.Arg)
		fn := strings.ToLower(strings.TrimPrefix(nm.Value, `\`))
		if !ok {
			return t
		}
		if x := unlessNull(a.Value); !isVar(a.Value, name) && (fn == "is_null" || x == nil || !isVar(x, name)) {
			return t
		}
		switch _, check := typeChecks[fn]; {
		case !check && fn != "count" && fn != "sizeof" && fn != "is_a" && fn != "is_subclass_of" && fn != "in_array":
			return t
		case !e.isBuiltinCall(c):
			return t // a namespaced function shadows the builtin
		}
		switch fn {
		case "count", "sizeof":
			return nonEmptyIf(t, truthy)
		case "is_a", "is_subclass_of":
			return e.isACond(t, c, fn, truthy)
		case "in_array":
			return e.inArrayCond(t, c, truthy)
		}
		atoms := typeChecks[fn]
		if fn == "is_numeric" && !truthy {
			// A string failing is_numeric() is a non-numeric string: only
			// int and float are excluded.
			atoms = atoms[:2]
		}
		return narrowAtoms(t, atoms, truthy)
	}
	return t
}

// unlessNull returns x of `x ?? null`, else nil: a type check other than
// is_null() is true of it exactly when it is true of x, which is then set.
func unlessNull(v syntax.Expr) syntax.Expr {
	b, ok := syntax.UnwrapParens(v).(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TCoalesce {
		return nil
	}
	if c, ok := syntax.UnwrapParens(b.Right).(*syntax.ConstFetch); ok && strings.EqualFold(c.Name.Value, "null") {
		return b.Left
	}
	return nil
}

// needleSearches are the builtins that find a needle in a haystack: a true
// (truthy) result means it was found, which an empty haystack never allows.
var needleSearches = map[string]bool{
	"str_contains": true, "str_starts_with": true, "str_ends_with": true,
	"strpos": true, "stripos": true, "strrpos": true, "strripos": true,
	"strstr": true, "stristr": true, "strrchr": true,
}

// needleFound reports whether c searches haystack name for a non-empty
// literal needle (`str_contains($x, 'a')`, `strpos($x, 'a')`): where it
// succeeds, name was neither null nor false, which read as `”`.
func (e *Env) needleFound(c *syntax.FuncCall, name string) bool {
	nm, ok := c.Name.(*syntax.Name)
	if !ok || c.Args == nil || !needleSearches[strings.ToLower(strings.TrimPrefix(nm.Value, `\`))] || !e.isBuiltinCall(c) {
		return false
	}
	// The haystack and the needle; any other argument (an offset) keeps the
	// proof.
	args, ok := bindArgs(c, "haystack", "needle")
	if !ok {
		return false
	}
	hay, needle := args[0], args[1]
	if hay == nil || needle == nil || !isVar(hay, name) {
		return false
	}
	text, ok := stringLiteral(needle)
	return ok && text != ""
}

// issetOffset narrows r, the type of name, where `isset($x[k])` holds with
// k a non-negative integer and the string lengths of name known: a string
// with a byte at offset k is at least k+1 bytes long.
func issetOffset(r types.Type, x syntax.Expr, name string) types.Type {
	d, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
	if _, known := r.StringLengthRanges(); !ok || !known || !isVar(d.Var, name) {
		return r
	}
	if k, ok := integerLiteral(d.Dim); ok && k >= 0 && k < math.MaxInt64 {
		return r.LengthsWithin(k+1, math.MaxInt64)
	}
	return r
}

// keyExistsCond narrows name, the array of a true `array_key_exists($k, $a)`
// whatever its key: from PHP 8.0 a value that is no array throws instead, and
// before it false and null make the call fail. The array holds a key.
func (e *Env) keyExistsCond(t types.Type, c *syntax.FuncCall, nm *syntax.Name, name string, truthy bool) (types.Type, bool) {
	switch strings.ToLower(strings.TrimPrefix(nm.Value, `\`)) {
	case "array_key_exists", "key_exists":
	default:
		return t, false
	}
	if !truthy || len(c.Args.Args) != 2 || !e.isBuiltinCall(c) {
		return t, false
	}
	a, ok := c.Args.Args[1].(*syntax.Arg)
	if !ok || a.Name != nil || a.Unpack || !isVar(a.Value, name) {
		return t, false
	}
	if e.PHP.AtLeast(phpversion.PHP80) {
		return narrowAtoms(t, typeChecks["is_array"], true).WithNonEmpty(true), true
	}
	return t.Without("null", "false").WithNonEmpty(true), true
}

// condAsserts applies the -assert-if-true (truthy) or -assert-if-false
// annotations of call c to the type t of name.
func (e *Env) condAsserts(c syntax.Expr, name string, truthy bool, t types.Type) (types.Type, bool) {
	ca := e.assertsOf(c)
	if ca == nil {
		return t, false
	}
	// Where the call is true, -if-true assertions hold and -if-false ones
	// are negated (`is_wp_error($x)` false: $x is not a WP_Error); and
	// conversely where it is false.
	kind, other := index.AssertIfTrue, index.AssertIfFalse
	if !truthy {
		kind, other = other, kind
	}
	t, ok := e.applyAsserts(ca, kind, name, t, false)
	t, ok2 := e.applyAsserts(ca, other, name, t, true)
	return t, ok || ok2
}

// narrowAtoms keeps (truthy) or removes (falsy) the atoms matching a type check.
func narrowAtoms(t types.Type, atoms []string, truthy bool) types.Type {
	// iterable and callable may each hold an array of unknown elements
	// (`[$object, 'method']` for a callable).
	if truthy && slices.Contains(atoms, "array") && (t.Has("iterable") || t.Has("callable")) {
		rest := t.Without("iterable", "callable")
		if len(rest.Atoms()) == 0 {
			return types.Array
		}
		return types.Union(narrowAtoms(rest, atoms, true), types.Array)
	}
	match := func(a string) bool {
		for _, x := range atoms {
			if a == x || (x == "array" && strings.HasSuffix(a, "[]")) || (x == "object" && strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]")) ||
				(x == "bool" && (a == "true" || a == "false")) {
				return true
			}
		}
		return false
	}
	var drop []string
	for _, a := range t.Atoms() {
		if match(a) != truthy {
			drop = append(drop, a)
		}
	}
	if truthy && t.Has("mixed") && len(drop) < len(t.Atoms()) {
		// mixed may hold any of the checked types, not only the matching
		// members (`string|mixed` passing is_scalar() may be an int).
		kept := slices.DeleteFunc(slices.Clone(t.Atoms()), func(a string) bool { return slices.Contains(drop, a) })
		if atoms[0] == "bool" {
			return types.Of(append(kept, "bool")...)
		}
		return types.Of(append(kept, atoms...)...)
	}
	if len(drop) == len(t.Atoms()) {
		if truthy {
			// Nothing known matches (e.g. mixed): the checked type, all of
			// it (is_numeric() is int|float|string, not int).
			if atoms[0] == "bool" {
				return types.Bool
			}
			return types.Of(atoms...)
		}
		// Every known member fails: a branch the types say is impossible
		// (often a PHPDoc type that does not hold, `!is_object($x)` on a
		// `@var Foo`): unknown rather than the refuted type.
		return types.Unknown
	}
	return t.Without(drop...) // keeps array facts (shapes) of the remaining members
}

// isEmptyArrayComparison reports `$x OP []` / `[] OP $x` on variable name.
func isEmptyArrayComparison(c *syntax.Binary, name string) bool {
	isEmpty := func(x syntax.Expr) bool {
		a, ok := syntax.UnwrapParens(x).(*syntax.Array)
		return ok && len(a.Items) == 0
	}
	return (isVar(c.Left, name) && isEmpty(c.Right)) || (isVar(c.Right, name) && isEmpty(c.Left))
}

// countComparison matches `count($x) OP n` or `n OP count($x)` on variable
// name with an integer literal n.
func countComparison(c *syntax.Binary, name string) (countCmp, bool) {
	isCount := func(x syntax.Expr) bool {
		f, ok := syntax.UnwrapParens(x).(*syntax.FuncCall)
		if !ok || f.Args == nil || len(f.Args.Args) != 1 {
			return false
		}
		nm, ok := f.Name.(*syntax.Name)
		if !ok {
			return false
		}
		switch strings.ToLower(strings.TrimPrefix(nm.Value, `\`)) {
		case "count", "sizeof":
		default:
			return false
		}
		a, ok := f.Args.Args[0].(*syntax.Arg)
		return ok && !a.Unpack && isVar(a.Value, name)
	}
	intLit := func(x syntax.Expr) (int64, bool) {
		k, ok := syntax.LiteralArrayKey(x)
		if !ok {
			return 0, false
		}
		l, isLit := syntax.UnwrapParens(x).(*syntax.Literal)
		if isLit && l.LitKind != syntax.LitInt {
			return 0, false
		}
		n, err := strconv.ParseInt(k, 10, 64)
		return n, err == nil
	}
	if isCount(c.Left) {
		if n, ok := intLit(c.Right); ok {
			return countCmp{n: n}, true
		}
	}
	if isCount(c.Right) {
		if n, ok := intLit(c.Left); ok {
			return countCmp{n: n, flipped: true}, true
		}
	}
	return countCmp{}, false
}

type countCmp struct {
	n       int64
	flipped bool // `n OP count($x)`
}

// countNonEmpty reports whether `count($x) OP n` evaluating to truthy
// implies count($x) >= 1.
func countNonEmpty(op syntax.TokenKind, c countCmp, truthy bool) bool {
	if c.flipped {
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
	if !truthy {
		// Negate the comparison.
		switch op {
		case syntax.TLess:
			op = syntax.TIsGreaterOrEqual
		case syntax.TGreater:
			op = syntax.TIsSmallerOrEqual
		case syntax.TIsSmallerOrEqual:
			op = syntax.TGreater
		case syntax.TIsGreaterOrEqual:
			op = syntax.TLess
		case syntax.TIsIdentical, syntax.TIsEqual:
			op = syntax.TIsNotIdentical
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			op = syntax.TIsIdentical
		}
	}
	n := c.n
	switch op {
	case syntax.TGreater:
		return n >= 0
	case syntax.TIsGreaterOrEqual:
		return n >= 1
	case syntax.TIsIdentical, syntax.TIsEqual:
		return n >= 1
	case syntax.TIsNotIdentical, syntax.TIsNotEqual:
		return n == 0
	}
	return false
}

// nonEmptyIf marks t's array members non-empty when cond (count($x) >= 1)
// holds; null is gone too (count(null) is 0 before PHP 8, a TypeError
// since).
func nonEmptyIf(t types.Type, cond bool) types.Type {
	if !cond {
		return t
	}
	if nt := t.Without("null"); len(nt.Atoms()) > 0 {
		t = nt
	}
	return t.WithNonEmpty(true)
}

// inListCond narrows the tracked string values of name by `in_array($name,
// ['a', 'b'], true)`: one of them when found, none of them when not. A loose
// search counts when no value is numeric (strings then compare by identity).
func (e *Env) inListCond(t types.Type, c *syntax.FuncCall, name string, truthy bool) (types.Type, bool) {
	if !t.StringsTracked() || c.Args == nil || len(c.Args.Args) < 2 || len(c.Args.Args) > 3 {
		return t, false
	}
	nm, ok := c.Name.(*syntax.Name)
	if !ok || !strings.EqualFold(strings.TrimPrefix(nm.Value, `\`), "in_array") {
		return t, false
	}
	args, ok := bindArgs(c, "needle", "haystack", "strict")
	if !ok {
		return t, false
	}
	list, ok := syntax.UnwrapParens(args[1]).(*syntax.Array)
	if !isVar(args[0], name) || !ok || len(list.Items) == 0 {
		return t, false
	}
	strict := args[2] != nil && constLiteral(args[2]) == "true"
	if args[2] != nil && !strict && constLiteral(args[2]) != "false" {
		return t, false
	}
	var values []string
	for _, it := range list.Items {
		if it == nil || it.ByRef || it.Unpack {
			return t, false
		}
		v, ok := stringLiteral(it.Value)
		if !ok || !strict && numericText(v) {
			return t, false
		}
		values = append(values, v)
	}
	if !e.isBuiltinCall(c) {
		return t, false
	}
	if truthy {
		return t.StringsOnly(values...), true
	}
	for _, v := range values {
		t = t.StringsWithout(v)
	}
	return t, true
}

// bindArgs binds the arguments of c to the parameters named params, by
// position or by name (matched exactly, as PHP does); a missing one is nil.
// Arguments past params and other names are left aside. An unpacked
// argument, or one given twice (the call then throws), binds nothing.
func bindArgs(c *syntax.FuncCall, params ...string) ([]syntax.Expr, bool) {
	out := make([]syntax.Expr, len(params))
	for i, x := range c.Args.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack {
			return nil, false
		}
		if a.Name != nil {
			i = slices.Index(params, a.Name.Value)
		}
		if i < 0 || i >= len(params) {
			continue
		}
		if out[i] != nil {
			return nil, false
		}
		out[i] = a.Value
	}
	return out, true
}

// keyUnder reports whether narrowing key name is variable v or a member or
// element of it.
func keyUnder(name, v string) bool {
	rest, ok := strings.CutPrefix(name, v)
	return ok && (rest == "" || !(rest[0] == '_' || '0' <= rest[0] && rest[0] <= '9' || 'a' <= rest[0]|0x20 && rest[0]|0x20 <= 'z' || rest[0] >= 0x80))
}
