package infer

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
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
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		if c.Op.Kind == syntax.TExclaim {
			return e.applyCondB(t, c.Expr, name, !truthy, steps)
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
				return r
			}
		}
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.ArrayDimFetch, *syntax.MethodCall:
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
	case *syntax.StaticCall:
		if r, ok := e.condAsserts(c, name, truthy, t); ok {
			return r
		}
	case *syntax.FuncCall:
		if r, ok := e.condAsserts(c, name, truthy, t); ok {
			return r
		}
		nm, ok := c.Name.(*syntax.Name)
		if !ok || c.Args == nil || len(c.Args.Args) == 0 {
			return t
		}
		a, ok := c.Args.Args[0].(*syntax.Arg)
		if !ok || !isVar(a.Value, name) {
			return t
		}
		fn := strings.ToLower(strings.TrimPrefix(nm.Value, `\`))
		switch fn {
		case "count", "sizeof":
			return nonEmptyIf(t, truthy)
		case "is_a", "is_subclass_of":
			return e.isACond(t, c, fn, truthy)
		case "in_array":
			return e.inArrayCond(t, c, truthy)
		}
		atoms, ok := typeChecks[fn]
		if !ok {
			return t
		}
		if fn == "is_numeric" && !truthy {
			// A string failing is_numeric() is a non-numeric string: only
			// int and float are excluded.
			atoms = atoms[:2]
		}
		return narrowAtoms(t, atoms, truthy)
	}
	return t
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
		k, ok := literalKey(x)
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
