package infer

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/index"
	"custos/internal/syntax"
	"custos/internal/types"
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
		if v, ok := n.Var.(*syntax.Variable); ok && v.Name == "this" && v.NameExpr == nil {
			if id, ok := n.Name.(*syntax.Identifier); ok {
				return "this->" + id.Value
			}
		}
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
// condition. scope is syntax.EnclosingFuncLike(x) (every caller passes it),
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

// guards applies early-exit guards among the statements preceding child
// in stmts (the statements of owner; nil: the file). Only the statements
// that can matter for name are visited (see guardIndex).
func (e *Env) guards(use syntax.Expr, scope syntax.Node, t types.Type, owner syntax.Node, stmts []syntax.Stmt, child syntax.Node, name string) types.Type {
	gi := e.guardIndexOf(owner, stmts)
	end, ok := gi.pos[child]
	if !ok {
		return t
	}
	cands := gi.candidates(name)
	// Only the statements after the last one assigning name (or its array)
	// matter: an assignment resets the guards seen before it. No statement
	// scanned below assigns name (or, for an element key, its array or an
	// element of it): gi.resets records every such top-level assignment.
	from := gi.lastReset(name, end)
	if from > 0 && strings.HasPrefix(name, "this->") && !isDimKey(name) {
		t = e.afterPropertyWrite(use, scope, t, stmts[from-1], name)
	}
	lo, _ := slices.BinarySearch(cands, from)
	hi, _ := slices.BinarySearch(cands, end)
	if hi-lo > maxGuardScan {
		return types.Unknown // hostile statement lists: see maxGuardScan
	}
	for _, i := range cands[lo:hi] {
		s := stmts[i]
		if es, ok := s.(*syntax.ExprStmt); ok {
			// `Assert::string($x);` (@phpstan-assert on the callee).
			if ca := e.assertsOf(es.Expr); ca != nil {
				t, _ = e.applyAsserts(ca, index.AssertAlways, name, t, false)
			}
			continue
		}
		if g, ok := s.(*syntax.If); ok && g.Else == nil && len(g.ElseIfs) == 0 {
			if terminates(g.Body) {
				t = e.condAt(use, scope, t, g.Cond, name, false, g.Span().End)
			} else if val := e.overwrites(g.Body, name); val != nil && !(isDimKey(name) && e.dimBroken(scope, name, g.Span().End, use)) &&
				!(strings.HasPrefix(name, "this->") && e.nonEmptyBroken(scope, name, g.Span().End, use)) {
				// `if (false === $x) { $x = $default; }`: past the if, the
				// condition no longer holds (unless the new value matches it).
				vt := e.TypeOf(val)
				if !vt.IsUnknown() && e.applyCond(vt, g.Cond, name, false).String() == vt.String() {
					t = e.applyCond(t, g.Cond, name, false)
				}
			}
		}
	}
	return t
}

// afterPropertyWrite drops null from the type t of `$this->prop` read use
// when s, a statement preceding it in an enclosing block (so it ran on
// every path to use), is `$this->prop = v;` or `$this->prop ??= v;` with a
// non-null v, and nothing that may reset the property (an assignment, a
// reference, a non-builtin call) lies between them (nonEmptyBroken).
func (e *Env) afterPropertyWrite(use syntax.Expr, scope syntax.Node, t types.Type, s syntax.Stmt, name string) types.Type {
	a := s.(*syntax.ExprStmt).Expr.(*syntax.Assign) // gi.resets lists only such statements
	if a.ByRef || (a.Op.Kind != syntax.TEqual && a.Op.Kind != syntax.TCoalesceEqual) || !t.IsNullable() {
		return t
	}
	vt := e.TypeOf(a.Value)
	if vt.IsUnknown() || vt.IsNullable() || vt.Has("mixed") || e.nonEmptyBroken(scope, name, a.Span().End, use) {
		return t
	}
	if nt := t.Without("null"); len(nt.Atoms()) > 0 {
		return nt
	}
	return t
}

// guardIndex lists, for one statement list, the statements guards may use
// for each narrowing key: expression statements (assignments, assertion
// calls) and else-less ifs, by the keys (narrowKey) occurring in them.
// guards used to test every preceding statement for each read, which is
// quadratic in long statement lists.
type guardIndex struct {
	pos  map[syntax.Node]int // statement -> position
	keys map[string][]int    // key -> positions, ascending
	// resets lists, per key, the positions of the statements assigning it
	// (assigns) and, under baseKey(k), those assigning an element of it
	// or it.
	resets map[string][]int
}

// maxGuardScan caps the statements guards examines for one read (those
// after the last assignment of the name that mention it). Beyond it the
// read is unknown: real code stays far below; hostile code with thousands
// of guards or calls on one variable would make each read linear.
const maxGuardScan = 512

func baseKey(k string) string { return "\x02" + k }

// lastReset returns the position just after the last statement before end
// that assigns name (for an element key, also its array): guards before
// it no longer apply (0 when none).
func (gi *guardIndex) lastReset(name string, end int) int {
	from := 0
	last := func(l []int) {
		if i, _ := slices.BinarySearch(l, end); i > 0 && l[i-1]+1 > from {
			from = l[i-1] + 1
		}
	}
	last(gi.resets[name])
	if isDimKey(name) {
		base, _ := splitDimKey(name)
		last(gi.resets[baseKey(base)])
	}
	return from
}

// thisCallKey lists the statements with a call on `$this` (assertions on
// `$this->prop` targets).
const thisCallKey = "this->*"

func (e *Env) guardIndexOf(owner syntax.Node, stmts []syntax.Stmt) *guardIndex {
	if gi, ok := e.guardIdx[owner]; ok {
		return gi
	}
	gi := &guardIndex{pos: make(map[syntax.Node]int, len(stmts)), keys: map[string][]int{}, resets: map[string][]int{}}
	for i, st := range stmts {
		gi.pos[st] = i
		seen := map[string]bool{}
		add := func(k string) {
			if k != "" && !seen[k] {
				seen[k] = true
				gi.keys[k] = append(gi.keys[k], i)
			}
		}
		collect := func(x syntax.Node) {
			syntax.Inspect(x, func(n syntax.Node) bool {
				switch n := n.(type) {
				case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
					return false
				case *syntax.Variable, *syntax.PropertyFetch, *syntax.ArrayDimFetch:
					add(narrowKey(n.(syntax.Expr)))
				case *syntax.MethodCall:
					if narrowKey(syntax.UnwrapParens(n.Var)) == "this" {
						add(thisCallKey)
					}
				case *syntax.StaticCall:
					if nm, ok := n.Class.(*syntax.Name); ok {
						switch strings.ToLower(nm.Value) {
						case "self", "static", "parent":
							add(thisCallKey)
						}
					}
				}
				return true
			})
		}
		switch st := st.(type) {
		case *syntax.ExprStmt:
			collect(st.Expr)
			if a, ok := st.Expr.(*syntax.Assign); ok {
				// The writes assigns matches, and under baseKey those of
				// the array or one of its elements.
				if k := narrowKey(a.Var); k != "" {
					gi.resets[k] = append(gi.resets[k], i)
				}
				x := a.Var
				for {
					d, ok := x.(*syntax.ArrayDimFetch)
					if !ok {
						break
					}
					x = d.Var
				}
				if k := narrowKey(x); k != "" {
					bk := baseKey(k)
					if l := gi.resets[bk]; len(l) == 0 || l[len(l)-1] != i {
						gi.resets[bk] = append(gi.resets[bk], i)
					}
				}
			}
		case *syntax.If:
			if st.Else == nil && len(st.ElseIfs) == 0 {
				collect(st.Cond)
				last := st.Body
				if b, ok := last.(*syntax.Block); ok && len(b.Stmts) > 0 {
					last = b.Stmts[len(b.Stmts)-1]
				}
				if es, ok := last.(*syntax.ExprStmt); ok {
					if a, ok := es.Expr.(*syntax.Assign); ok {
						add(narrowKey(a.Var))
					}
				}
			}
		}
	}
	if e.guardIdx == nil {
		e.guardIdx = map[syntax.Node]*guardIndex{}
	}
	e.guardIdx[owner] = gi
	return gi
}

// candidates returns the positions of the statements that may concern
// name, ascending: those mentioning it, its array for an element key, and
// calls on `$this` for a `$this->prop` key.
func (gi *guardIndex) candidates(name string) []int {
	lists := [][]int{gi.keys[name]}
	if isDimKey(name) {
		base, _ := splitDimKey(name)
		lists = append(lists, gi.keys[base])
		if strings.HasPrefix(base, "this->") {
			lists = append(lists, gi.keys[thisCallKey])
		}
	} else if strings.HasPrefix(name, "this->") {
		lists = append(lists, gi.keys[thisCallKey])
	}
	n := 0
	var only []int
	for _, l := range lists {
		if len(l) > 0 {
			n++
			only = l
		}
	}
	if n <= 1 {
		return only
	}
	var out []int
	for _, l := range lists {
		out = append(out, l...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// assigns reports whether statement s writes variable name at its top level.
func assigns(s syntax.Stmt, name string) bool {
	es, ok := s.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	a, ok := es.Expr.(*syntax.Assign)
	if !ok {
		return false
	}
	return narrowKey(a.Var) == name
}

// overwrites returns the value of a plain `$name = value;` that is the last
// statement of body, or nil.
func (e *Env) overwrites(body syntax.Stmt, name string) syntax.Expr {
	last := body
	if b, ok := body.(*syntax.Block); ok {
		if len(b.Stmts) == 0 {
			return nil
		}
		last = b.Stmts[len(b.Stmts)-1]
	}
	if !assigns(last, name) {
		return nil
	}
	a := last.(*syntax.ExprStmt).Expr.(*syntax.Assign)
	if a.Op.Kind != syntax.TEqual || a.ByRef {
		return nil
	}
	return a.Value
}

// terminates reports whether a statement always leaves the current block.
// Deliberately shallower than syntax.Terminates (no if/else, try or switch
// analysis): switching would make more guards count as early exits, which
// changes narrowing and therefore findings.
func terminates(s syntax.Stmt) bool {
	switch n := s.(type) {
	case *syntax.Return, *syntax.Break, *syntax.Continue, *syntax.Goto:
		return true
	case *syntax.ExprStmt:
		switch n.Expr.(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
	case *syntax.Block:
		if len(n.Stmts) > 0 {
			return terminates(n.Stmts[len(n.Stmts)-1])
		}
	}
	return false
}

func isVar(x syntax.Expr, name string) bool {
	x = syntax.UnwrapParens(x)
	if a, ok := x.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
		x = a.Var // `false === ($x = f())` tests $x
	}
	return narrowKey(x) == name
}

// constLiteral returns "null", "true" or "false" for those constants.
func constLiteral(x syntax.Expr) string {
	c, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch)
	if !ok {
		return ""
	}
	switch v := strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)); v {
	case "null", "true", "false":
		return v
	}
	return ""
}

var typeChecks = map[string][]string{
	"is_array": {"array"}, "is_string": {"string"}, "is_int": {"int"}, "is_integer": {"int"},
	"is_long": {"int"}, "is_float": {"float"}, "is_double": {"float"}, "is_bool": {"bool", "true", "false"},
	"is_null": {"null"}, "is_object": {"object"}, "is_callable": {"callable"}, "is_iterable": {"iterable", "array"},
	"is_numeric": {"int", "float", "string"}, "is_scalar": {"int", "float", "string", "bool", "true", "false"},
	"is_resource": {"resource"},
}

// maxCondSteps caps the condition nodes one applyCond call examines. The
// negation of `A && B` (and the truth of `A || B`) narrows A both ways, so
// alternating nestings would otherwise cost 3^depth; beyond the budget the
// remaining sub-conditions narrow nothing (sound: the type is kept).
const maxCondSteps = 1024

// applyCond narrows t assuming cond evaluates to `truthy`.
func (e *Env) applyCond(t types.Type, cond syntax.Expr, name string, truthy bool) types.Type {
	steps := maxCondSteps
	return e.applyCondB(t, cond, name, truthy, &steps)
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

func (e *Env) applyCondB(t types.Type, cond syntax.Expr, name string, truthy bool, steps *int) types.Type {
	if t.IsUnknown() || *steps <= 0 {
		return t
	}
	*steps--
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
			return instanceType(t, cls)
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
		return t
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
