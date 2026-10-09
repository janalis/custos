package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// narrow refines the type of variable occurrence v using the conditions that
// guard it: ternary branches, if/elseif/else bodies, the right operand of
// && / ||, and early-exit guards (`if (cond) { return; }`) earlier in the
// same block. Only checks on the same variable name are understood.
func (e *Env) narrow(t types.Type, v *syntax.Variable, scope syntax.Node, after uint32) types.Type {
	return e.narrowExprAfter(t, v, v.Name, scope, after)
}

// narrowKey identifies an expression whose type guards can refine: a
// variable by its name, `$this->prop` as "this->prop", and an element with
// a literal key of either (`$a['k']`) as dimKey(base, key) ("" otherwise).
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
		if base == "" || isDimKey(base) {
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

// dimSep separates the base and the key of an element's narrowing key.
const dimSep = "\x01"

// dimKey is the narrowing key of element `key` of base (see narrowKey);
// dimKey(base, "*") collects the writes with a computed key.
func dimKey(base, key string) string { return base + dimSep + key }

func isDimKey(k string) bool { return strings.Contains(k, dimSep) }

// splitDimKey returns the base and key of an element's narrowing key.
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
		switch n := p.(type) {
		case *syntax.Ternary:
			if n.Then != nil && child == syntax.Node(n.Then) {
				t = cond(x, scope, t, n.Cond, key, true)
			} else if child == syntax.Node(n.Else) && n.Then != nil {
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
			}
		case *syntax.ElseIf:
			if child == syntax.Node(n.Body) {
				// The conditions before it in the chain were false.
				ifn := n.Parent().(*syntax.If)
				if i := nodeIndex(ifn.ElseIfs, n); i < len(ifn.ElseIfs) && i < maxBranchScan {
					t = cond(x, scope, t, ifn.Cond, key, false)
					for _, ei := range ifn.ElseIfs[:i] {
						t = cond(x, scope, t, ei.Cond, key, false)
					}
				}
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
	if b := propBase(key); b != "" && r.ShapeString() != t.ShapeString() && e.chainBroken(scope, key, b, from, use) {
		return t // `$v` (or a property along the chain) changed since
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
// the same literal key or to a computed key.
func (e *Env) dimBroken(scope syntax.Node, key string, from uint32, use syntax.Node) bool {
	base, _ := splitDimKey(key)
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
	c, ok := x.(*syntax.FuncCall)
	if !ok || len(c.Args.Args) == 0 {
		return nil
	}
	nm, ok := c.Name.(*syntax.Name)
	if !ok {
		return nil
	}
	fqn, fb := e.Names.Function(nm.Value, nm.Span().Start)
	if !strings.EqualFold(fqn, "assert") && !(strings.EqualFold(fb, "assert") && e.Index.Function(fqn, e.PHP) == nil) {
		return nil
	}
	a, ok := c.Args.Args[0].(*syntax.Arg)
	if !ok || a.Unpack || a.Name != nil {
		return nil
	}
	return a.Value
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
