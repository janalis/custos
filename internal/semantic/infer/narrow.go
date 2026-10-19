package infer

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// NarrowMember narrows seed, a type assumed for the property, static
// property or literal-key element read x, by the conditions guarding x, as
// TypeOf narrows such a read whose declared type is known: what a guard
// proves of a value its declaration says nothing about (a result stored in
// an untyped property). Writes to x, its object or its array between a
// condition and x end what the condition proves; a write of an unknown
// value keeps the assumed type. Plain variables, whose
// narrowing starts after their reaching definitions, and other expressions
// return seed.
func (e *Env) NarrowMember(x syntax.Expr, seed types.Type) types.Type {
	x = syntax.UnwrapParens(x)
	key := narrowKey(x)
	if !isPropKey(key) && !isDimKey(key) {
		return seed
	}
	e.seedUse = x
	defer func() { e.seedUse = nil }()
	return e.narrowExpr(seed, x, key, syntax.EnclosingVariableScope(x))
}

// narrowKey identifies an expression whose type guards can refine: a
// variable by its name, `$this->prop` as "this->prop", and an element with
// a literal key of either (`$a['k']`, nested up to maxDimDepth levels:
// `$a['k']['l']`) as dimKey(base, key) ("" otherwise).
func narrowKey(x syntax.Expr) string {
	switch n := x.(type) {
	case *syntax.Variable:
		if n.NameExpr == nil {
			return n.Name
		}
	case *syntax.PropertyFetch:
		// `$this->p`, `$v->p` on a plain variable and chains of them
		// (`$param->var->name`); their facts also end when the variable or
		// a property along the chain changes (see propBase, condAt).
		id, ok := n.Name.(*syntax.Identifier)
		if !ok {
			return ""
		}
		switch b := syntax.UnwrapParens(n.Var).(type) {
		case *syntax.Variable:
			if b.Name != "" && b.NameExpr == nil {
				return b.Name + "->" + id.Value
			}
		case *syntax.PropertyFetch:
			if base := narrowKey(b); base != "" && !isDimKey(base) && !strings.Contains(base, "::") {
				return base + "->" + id.Value
			}
		}
	case *syntax.StaticPropertyFetch:
		// `self::$p`, `static::$p` (the same property in practice) and
		// `Name::$p`, by the class as written.
		nm, ok := n.Class.(*syntax.Name)
		v, ok2 := n.Name.(*syntax.Variable)
		if !ok || !ok2 || v.NameExpr != nil || v.Name == "" {
			return ""
		}
		cls := strings.ToLower(strings.TrimPrefix(nm.Value, `\`))
		if cls == "static" {
			cls = "self"
		}
		return cls + "::" + v.Name
	case *syntax.ArrayDimFetch:
		if n.Dim == nil {
			return ""
		}
		base := narrowKey(n.Var)
		if base == "" || strings.Count(base, dimSep) >= maxDimDepth-1 {
			return ""
		}
		if k, ok := literalKey(n.Dim); ok {
			return dimKey(base, k)
		}
	}
	return ""
}

// isPropKey reports the narrowing key of a property (`$this->p`, a static
// property): any non-builtin call may change it.
// callKeyPrefix starts the narrowing key of one call node (see callKey):
// the result of that very call, not of the same text elsewhere.
const callKeyPrefix = "\x03call "

// callKey is the narrowing key of call c, by its span.
func callKey(c *syntax.FuncCall) string {
	sp := c.Span()
	return callKeyPrefix + strconv.FormatUint(uint64(sp.Start), 10) + " " + strconv.FormatUint(uint64(sp.End), 10)
}

// callMatches reports whether x is the call whose key is name.
func callMatches(x syntax.Expr, name string) bool {
	rest, ok := strings.CutPrefix(name, callKeyPrefix)
	if !ok {
		return false
	}
	c, ok := unsilenced(x).(*syntax.FuncCall) // `@f()` is f()'s result
	if !ok {
		return false
	}
	start, end := callKeySpan(rest)
	sp := c.Span()
	return uint64(sp.Start) == start && uint64(sp.End) == end
}

// callKeySpan decodes the offsets of a call key stripped of its prefix.
func callKeySpan(rest string) (start, end uint64) {
	a, b, _ := strings.Cut(rest, " ")
	start, _ = strconv.ParseUint(a, 10, 32)
	end, _ = strconv.ParseUint(b, 10, 32)
	return start, end
}

// CallResultAt types the result of call c as known at expression at, a
// later place of its scope: the call's type narrowed by the conditions on
// that very call guarding at, written directly (`if (flock($h, LOCK_EX))`,
// `flock($h, LOCK_EX) or exit;`, `if (!@f()) { return; }`), assigned in
// place (`if ($ok = f())`) or through the variable it was stored in, while
// that assignment is the only one reaching at (`$ok = f(); if (!$ok) { … }`). A call that may run again before at
// (inside a loop around both) still counts by its last result.
func (e *Env) CallResultAt(c *syntax.FuncCall, at syntax.Expr) types.Type {
	return e.CallResultAsAt(c, at, e.TypeOf(c))
}

// CallResultAsAt is CallResultAt narrowing seed, a type the caller assumes
// for the result (one modelling values the declared type cannot tell
// apart), instead of the call's type.
func (e *Env) CallResultAsAt(c *syntax.FuncCall, at syntax.Expr, seed types.Type) types.Type {
	scope := syntax.EnclosingVariableScope(c)
	if syntax.EnclosingVariableScope(at) != scope || at.Span().Start < c.Span().End {
		return types.Unknown
	}
	t := e.narrowExprAfter(seed, at, callKey(c), scope, c.Span().End)
	if t.IsUnknown() {
		return t
	}
	for _, a := range e.holdersAt(c, at, scope) {
		// `$ok = f(); if ($ok != true) { return; }`: the tests of a variable
		// holding the result, narrowing the same seed, prove as much of it.
		vt := e.narrowExprAfter(seed, at, a.Var.(*syntax.Variable).Name, scope, a.Span().End)
		if vt.IsUnknown() {
			return vt // an impossible path
		}
		t = e.assertIs(t, vt).IntsAlso(vt)
	}
	return t
}

// maxHolderHops bounds the plain copies followed from a call result to a
// variable holding it (`$s = f(); $t = $s;` is two).
const maxHolderHops = 4

// holdersAt returns, for each local holding the result of call c at at, the
// plain assignment defining it there: its only reaching definition stores c
// (`$ok = f()`, `$ok = @f()`) or a copy of such a local (`$t = $ok`).
func (e *Env) holdersAt(c *syntax.FuncCall, at syntax.Expr, scope syntax.Node) []*syntax.Assign {
	sv := e.scopeVars(scope)
	// Candidates first: names assigned c, then names assigned a copy of a
	// candidate, before checking what reaches at.
	names := map[string]bool{}
	for hop := 0; hop <= maxHolderHops; hop++ {
		grew := false
		for name, defs := range sv.defs {
			if names[name] || len(defs) > maxVarDefs {
				continue
			}
			for _, d := range defs {
				if d.asg == nil {
					continue
				}
				v, copied := unsilenced(d.asg.Value).(*syntax.Variable)
				if unsilenced(d.asg.Value) == syntax.Expr(c) || copied && hop > 0 && names[v.Name] {
					names[name], grew = true, true
					break
				}
			}
		}
		if !grew {
			break
		}
	}
	var out []*syntax.Assign
	for name := range names {
		if a := e.onlyAssign(name, at, scope); a != nil && e.holds(a.Value, c, scope, 0) {
			out = append(out, a)
		}
	}
	slices.SortFunc(out, func(a, b *syntax.Assign) int { return int(a.Span().Start) - int(b.Span().Start) })
	return out
}

// unsilenced strips parentheses and `@` from x.
func unsilenced(x syntax.Expr) syntax.Expr {
	x = syntax.UnwrapParens(x)
	for {
		u, silenced := x.(*syntax.Unary)
		if !silenced || u.Op.Kind != syntax.TAt {
			return x
		}
		x = syntax.UnwrapParens(u.Expr)
	}
}

// onlyAssign returns the plain assignment that is the only definition of
// variable name reaching at (nil else).
func (e *Env) onlyAssign(name string, at, scope syntax.Node) *syntax.Assign {
	sv := e.scopeVars(scope)
	defs := sv.defs[name] // holdersAt only asks within maxVarDefs
	fwd, back, _ := e.reaching(defs, at, scope)
	if len(fwd) != 1 || len(back) != 0 || fwd[0].asg == nil || e.clobbered(sv, fwd, back, at, scope) {
		return nil
	}
	return fwd[0].asg
}

// holds reports whether value is the result of call c: the call itself
// (parenthesised or silenced) or a local holding it where it is read.
func (e *Env) holds(value syntax.Expr, c *syntax.FuncCall, scope syntax.Node, hops int) bool {
	switch v := unsilenced(value).(type) {
	case *syntax.FuncCall:
		return v == c
	case *syntax.Variable:
		if hops >= maxHolderHops || v.Name == "" || v.NameExpr != nil {
			return false
		}
		a := e.onlyAssign(v.Name, v, scope)
		return a != nil && e.holds(a.Value, c, scope, hops+1)
	}
	return false
}

func isPropKey(k string) bool { return strings.Contains(k, "->") || strings.Contains(k, "::") }

// propBase returns the variable holding the object of a `$v->p` key (also
// `this`, which never changes itself, for its chains); "" for other keys.
func propBase(k string) string {
	if isDimKey(k) {
		return ""
	}
	if b, _, ok := strings.Cut(k, "->"); ok {
		return b
	}
	return ""
}

// chainBroken reports whether the variable b holding the object of
// property key k, or one of the properties along the chain (`$a->b` for
// `$a->b->c`), was written between from and use.
func (e *Env) chainBroken(scope syntax.Node, k, b string, from uint32, use syntax.Node) bool {
	if e.nonEmptyBroken(scope, b, from, use) {
		return true
	}
	for i := strings.LastIndex(k, "->"); i > len(b); i = strings.LastIndex(k[:i], "->") {
		if e.brokenBy(scope, from, use, e.mutations(scope, k[:i])) {
			return true
		}
	}
	return false
}

// maxDimDepth bounds the literal keys of a narrowed element (`$a['k']['l']`
// is two).
const maxDimDepth = 4

// dimSep separates the base and the key of an element's narrowing key.
const dimSep = "\x01"

// dimKey is the narrowing key of element `key` of base (see narrowKey);
// dimKey(base, "*") collects the writes with a computed key.
func dimKey(base, key string) string { return base + dimSep + key }

func isDimKey(k string) bool { return strings.Contains(k, dimSep) }

// splitDimKey returns the base and key of an element's narrowing key: the
// variable or property, and the literal key path below it (still joined by
// dimSep for a nested element).
func splitDimKey(k string) (base, key string) {
	base, key, _ = strings.Cut(k, dimSep)
	return base, key
}

// narrowExpr refines the type t of occurrence x (identified by key, see
// narrowKey) by the conditions guarding it.
func (e *Env) narrowExpr(t types.Type, x syntax.Expr, key string, scope syntax.Node) types.Type {
	return e.narrowExprAfter(t, x, key, scope, 0)
}

// narrowExprAfter is narrowExpr ignoring the enclosing conditions that end
// before after (the end of the last definition reaching x): a value
// assigned inside `if (null === $x) { $x = f(); }` is not narrowed by that
// condition. scope is syntax.EnclosingVariableScope(x) (every caller passes it),
// so the walk up to it never crosses another function boundary.
func (e *Env) narrowExprAfter(t types.Type, x syntax.Expr, key string, scope syntax.Node, after uint32) types.Type {
	if t.IsUnknown() || key == "" {
		return t
	}
	cond := func(use syntax.Expr, scope syntax.Node, t types.Type, c syntax.Expr, key string, truthy bool) types.Type {
		if c.Span().End < after {
			return t
		}
		return e.cond(use, scope, t, c, key, truthy)
	}
	var child syntax.Node = x
	for p := x.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		if _, top := p.(*syntax.File); top {
			break // child is a top-level statement: see the file's guards below
		}
		switch n := p.(type) {
		case *syntax.Ternary:
			if n.Then != nil && child == syntax.Node(n.Then) {
				t = cond(x, scope, t, n.Cond, key, true)
			} else if child == syntax.Node(n.Else) {
				// Also `c ?: v`: v runs when c is falsy.
				t = cond(x, scope, t, n.Cond, key, false)
			}
		case *syntax.Binary:
			if child == syntax.Node(n.Right) {
				switch n.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					t = cond(x, scope, t, n.Left, key, true)
				case syntax.TBooleanOr, syntax.TOr:
					t = cond(x, scope, t, n.Left, key, false)
				}
			}
		case *syntax.If:
			if child == syntax.Node(n.Body) {
				t = cond(x, scope, t, n.Cond, key, true)
			} else if n.Else != nil && child == syntax.Node(n.Else) && len(n.ElseIfs) < maxBranchScan {
				// else: every condition of the chain was false.
				t = cond(x, scope, t, n.Cond, key, false)
				for _, ei := range n.ElseIfs {
					t = cond(x, scope, t, ei.Cond, key, false)
				}
			}
		case *syntax.While:
			if child == syntax.Node(n.Body) {
				t = cond(x, scope, t, n.Cond, key, true)
			}
		case *syntax.For:
			// The last condition expression decides each iteration.
			if child == syntax.Node(n.Body) && len(n.Cond) > 0 {
				t = cond(x, scope, t, n.Cond[len(n.Cond)-1], key, true)
			} else if len(n.Cond) > 0 && e.stepAfterCond(n, child, key, scope) {
				// A step runs after the body, which ran with the condition
				// true and left the variable alone.
				t = cond(x, scope, t, n.Cond[len(n.Cond)-1], key, true)
			}
		case *syntax.ElseIf:
			// Its condition, like its body, runs once the conditions before
			// it in the chain were false.
			ifn := n.Parent().(*syntax.If)
			if i := nodeIndex(ifn.ElseIfs, n); i < len(ifn.ElseIfs) && i < maxBranchScan {
				t = cond(x, scope, t, ifn.Cond, key, false)
				for _, ei := range ifn.ElseIfs[:i] {
					t = cond(x, scope, t, ei.Cond, key, false)
				}
			}
			if child == syntax.Node(n.Body) {
				t = cond(x, scope, t, n.Cond, key, true)
			}
		case *syntax.MatchArm:
			t = e.matchArm(x, scope, t, n, child, key, after)
		case *syntax.Block:
			t = e.guards(x, scope, t, n, n.Stmts, child, key)
		case *syntax.Case:
			t = e.guards(x, scope, t, n, n.Stmts, child, key)
			if child != syntax.Node(n.Cond) {
				t = e.switchCase(x, scope, t, n, key, after)
			}
		case *syntax.Namespace:
			t = e.guards(x, scope, t, n, n.Stmts, child, key)
		}
	}
	if scope == nil {
		t = e.guards(x, scope, t, nil, e.File.Stmts, child, key)
	}
	return t
}

// cond applies condition c (evaluated to truthy) to the type t of use. A
// non-empty-array fact it establishes is kept only when no mutation can have
// emptied the array between the condition and use (for `$this->prop`, no
// non-builtin call either).
func (e *Env) cond(use syntax.Expr, scope syntax.Node, t types.Type, c syntax.Expr, key string, truthy bool) types.Type {
	return e.condAt(use, scope, t, c, key, truthy, c.Span().End)
}

func (e *Env) condAt(use syntax.Expr, scope syntax.Node, t types.Type, c syntax.Expr, key string, truthy bool, from uint32) types.Type {
	r := e.applyCond(t, c, key, truthy)
	if !e.options.Assertions && !isDimKey(key) && !isPropKey(key) && r.ShapeString() != t.ShapeString() {
		for _, m := range e.mutations(scope, key) {
			if m.optional && e.brokenBy(scope, from, use, []mutation{m}) {
				return t
			}
		}
	}
	if b := propBase(key); b != "" && r.ShapeString() != t.ShapeString() && e.chainBroken(scope, key, b, from, use) {
		return t // `$v` (or a property along the chain) changed since
	}
	if isPropKey(key) && !isDimKey(key) && r.ShapeString() != t.ShapeString() && e.brokenBy(scope, from, use, e.mutations(scope, key)) {
		return t // the property itself was written since
	}
	if isPropKey(key) && !isDimKey(key) && r.ShapeString() != t.ShapeString() && e.native && e.memberBroken(scope, key, from, use) {
		return t
	}
	if !isPropKey(key) && !isDimKey(key) && r.ShapeString() != t.ShapeString() && e.scopeVars(scope).statics[key] && e.brokenBy(scope, from, use, e.mutations(scope, anyCall)) {
		return t // a call may enter the function again and assign its static local
	}
	if isDimKey(key) {
		// An element is narrowed only while neither it nor its array can
		// have changed since the condition.
		if r.ShapeString() != t.ShapeString() && e.dimBroken(scope, key, from, use) {
			return t
		}
		return r
	}
	if k, changed := e.applyKeyCond(r, c, key, truthy); changed && !e.nonEmptyBroken(scope, key, from, use) {
		r = k
	}
	if r.IsNonEmptyArray() && !t.IsNonEmptyArray() {
		if e.nonEmptyBroken(scope, key, from, use) {
			r = r.WithNonEmpty(false)
		}
	}
	return r
}

// dimBroken reports whether element key (a dimKey) may have changed
// between from and use: a mutation of its array (assignment, reference,
// unset, by-reference argument; for `$this->prop` any call), or a write to
// the same first-level literal key (anywhere below it, for a nested
// element) or to a computed key.
func (e *Env) dimBroken(scope syntax.Node, key string, from uint32, use syntax.Node) bool {
	base, path := splitDimKey(key)
	if first, _, nested := strings.Cut(path, dimSep); nested {
		key = dimKey(base, first) // element writes are recorded there
	}
	if e.nonEmptyBroken(scope, base, from, use) {
		return true
	}
	if b := propBase(base); b != "" && e.nonEmptyBroken(scope, b, from, use) {
		return true
	}
	return e.brokenBy(scope, from, use, e.mutations(scope, key), e.mutations(scope, dimKey(base, "*")))
}

// applyKeyCond narrows the array t of variable (or `$this->prop`) name by
// conditions on its elements: `isset($a['k'])`, `$a['k'] !== null`,
// `!empty($a['k'])`, `$a['k']` (truthy) make key k present with a non-null
// (non-false for the last two) value; `array_key_exists('k', $a)` makes it
// present. The array then is neither null nor false, and non-empty.
// changed reports whether a condition applied.
func (e *Env) applyKeyCond(t types.Type, cond syntax.Expr, name string, truthy bool) (types.Type, bool) {
	elem := func(x syntax.Expr) (string, bool) {
		d, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
		if !ok || d.Dim == nil || narrowKey(d.Var) != name {
			return "", false
		}
		return literalKey(d.Dim)
	}
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		if c.Op.Kind == syntax.TExclaim {
			return e.applyKeyCond(t, c.Expr, name, !truthy)
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			if truthy {
				l, cl := e.applyKeyCond(t, c.Left, name, true)
				r, cr := e.applyKeyCond(l, c.Right, name, true)
				return r, cl || cr
			}
		case syntax.TBooleanOr, syntax.TOr:
			if !truthy {
				l, cl := e.applyKeyCond(t, c.Left, name, false)
				r, cr := e.applyKeyCond(l, c.Right, name, false)
				return r, cl || cr
			}
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
			notNull := c.Op.Kind == syntax.TIsNotIdentical || c.Op.Kind == syntax.TIsNotEqual
			if notNull != truthy {
				break
			}
			k, ok := elem(c.Left)
			other := c.Right
			if !ok {
				k, ok = elem(c.Right)
				other = c.Left
			}
			if ok && syntax.IsNullConst(syntax.UnwrapParens(other)) {
				return keyPresent(t, k, "null"), true
			}
		}
	case *syntax.Isset:
		if !truthy {
			break
		}
		changed := false
		for _, x := range c.Vars {
			if k, ok := elem(x); ok {
				t, changed = keyPresent(t, k, "null"), true
			}
		}
		return t, changed
	case *syntax.Empty:
		if k, ok := elem(c.Expr); ok && !truthy {
			return keyPresent(t, k, "null", "false"), true
		}
	case *syntax.ArrayDimFetch:
		if k, ok := elem(c); ok && truthy {
			return keyPresent(t, k, "null", "false"), true
		}
	case *syntax.FuncCall:
		nm, ok := c.Name.(*syntax.Name)
		if !ok || !truthy || c.Args == nil || len(c.Args.Args) != 2 {
			break
		}
		switch strings.ToLower(strings.TrimPrefix(nm.Value, `\`)) {
		case "array_key_exists", "key_exists":
		default:
			return t, false
		}
		a0, ok0 := c.Args.Args[0].(*syntax.Arg)
		a1, ok1 := c.Args.Args[1].(*syntax.Arg)
		if !ok0 || !ok1 || a0.Name != nil || a1.Name != nil || a0.Unpack || a1.Unpack || narrowKey(syntax.UnwrapParens(a1.Value)) != name {
			break
		}
		if k, ok := literalKey(a0.Value); ok {
			return keyPresent(t, k), true
		}
	}
	return t, false
}

// keyPresent marks key k of the array members of t as present (a shape
// key becomes required, its value loses the drop atoms), the array as
// non-empty, and drops null and false from t.
func keyPresent(t types.Type, k string, drop ...string) types.Type {
	if nt := t.Without("null", "false"); len(nt.Atoms()) > 0 {
		t = nt
	}
	if t.HasShape() {
		t = t.MapShape(func(sk types.ShapeKey) types.ShapeKey {
			if sk.Name != k {
				return sk
			}
			sk.Optional = false
			if nv := sk.Type.Without(drop...); len(drop) > 0 && len(nv.Atoms()) > 0 {
				sk.Type = nv
			}
			return sk
		})
	}
	return t.WithNonEmpty(true)
}

// aliasCond returns the condition a boolean variable v stands for when it
// narrows variable name: v's only reaching definition is `$v = cond;` with
// cond a boolean expression, and name is not written between that
// assignment and v. Only direct aliases (an alias of an alias is not
// followed) and plain variable names; nil otherwise.
func (e *Env) aliasCond(v *syntax.Variable, name string) syntax.Expr {
	if v.NameExpr != nil || v.Name == "" || v.Name == "this" || strings.ContainsAny(name, ">:"+dimSep) {
		return nil
	}
	scope := syntax.EnclosingVariableScope(v)
	defs := e.scopeVars(scope).defs[v.Name]
	if len(defs) == 0 || len(defs) > maxVarDefs {
		return nil
	}
	fwd, back, _ := e.reaching(defs, v, scope)
	if len(fwd) != 1 || len(back) != 0 || fwd[0].asg == nil {
		return nil
	}
	a := fwd[0].asg
	cond := syntax.UnwrapParens(a.Value)
	if !isBoolExpr(cond) {
		return nil
	}
	end, at := a.Span().End, v.Span().Start
	for _, m := range e.mutations(scope, name) {
		if m.span.Start >= end && m.span.Start < at {
			return nil // name changed since the alias was computed
		}
	}
	return cond
}

// assertCall returns the condition of a call statement `assert(cond)` to
// the global assert() (nil otherwise).
func (e *Env) assertCall(x syntax.Expr) syntax.Expr {
	if !e.options.Assertions {
		return nil
	}
	if call, ok := syntax.UnwrapParens(x).(*syntax.FuncCall); ok {
		return e.AssertGuard(call)
	}
	return nil
}

// memberBroken reports, for the native environment, which proves rather
// than assumes, whether a fact a condition established at from about
// property key may no longer hold at use: a call that may run code of the
// project lies between them, a property of that name is written through
// another receiver (which may hold the same object), or reading use may run
// code itself (see plainMember).
func (e *Env) memberBroken(scope syntax.Node, key string, from uint32, use syntax.Expr) bool {
	if e.brokenBy(scope, from, use, e.mutations(scope, anyCall)) {
		return true
	}
	pf, ok := syntax.UnwrapParens(use).(*syntax.PropertyFetch)
	if !ok {
		return false // a static property: no other receiver, no hook
	}
	id := pf.Name.(*syntax.Identifier) // a narrowing key names the property
	return e.brokenBy(scope, from, use, e.mutations(scope, propNameKey(id.Value)), e.mutations(scope, propNameKey("*"))) || !e.plainMember(pf, id.Value)
}

// plainMember reports a property read that runs no code: every class its
// receiver may be of declares the property, without a hook, and has no
// __get method. A class extending it that declares the property again with
// a hook is not considered.
func (e *Env) plainMember(pf *syntax.PropertyFetch, name string) bool {
	cs := e.TypeOf(pf.Var).Classes()
	for _, c := range cs {
		c = strings.TrimPrefix(c, `\`)
		p := e.Index.FindProperty(c, name, e.PHP)
		if p == nil || p.Magic || p.Static || p.Hooked || p.ReadsRunCode || e.Index.FindMethod(c, "__get", e.PHP) != nil {
			return false
		}
	}
	return len(cs) > 0
}

// nativeProp reports whether property fetch x names a property with a
// native type on every class of its receiver (PHP enforces it on write).
func (e *Env) nativeProp(x syntax.Expr) bool {
	pf, ok := x.(*syntax.PropertyFetch)
	if !ok {
		return false
	}
	id := pf.Name.(*syntax.Identifier) // a narrowing key names the property
	cs := e.TypeOf(pf.Var).Classes()
	for _, c := range cs {
		p := e.Index.FindProperty(strings.TrimPrefix(c, `\`), id.Value, e.PHP)
		if p == nil || p.Type == "" {
			return false
		}
	}
	return len(cs) > 0
}

// stepAfterCond reports whether child is a step expression of for loop n
// that only runs with n's condition still true for variable key: neither
// the body (a continue reaches the steps without testing the condition
// again) nor an earlier step assigns it.
func (e *Env) stepAfterCond(n *syntax.For, child syntax.Node, key string, scope syntax.Node) bool {
	i := slices.IndexFunc(n.Loop, func(x syntax.Expr) bool { return syntax.Node(x) == child })
	if i < 0 || isPropKey(key) || isDimKey(key) || n.Body == nil {
		return false
	}
	body := n.Body.Span()
	for _, d := range e.scopeVars(scope).defs[key] {
		if body.Contains(syntax.Span{Start: d.pos, End: d.pos}) || d.pos >= n.Loop[0].Span().Start && d.pos < n.Loop[i].Span().Start {
			return false
		}
	}
	for _, m := range e.mutations(scope, key) {
		if body.Contains(m.span) {
			return false
		}
	}
	return true
}
