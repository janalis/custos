package infer

import (
	"cmp"
	"slices"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
)

// ---- array shapes ---------------------------------------------------------------------

// shapeDim types `$a['key']` from the shape of $a when the key is a literal
// listed by the shape; ok is false otherwise (callers fall back to the
// element type).
func (e *Env) shapeDim(ct types.Type, n *syntax.ArrayDimFetch) (types.Type, bool) {
	if !ct.HasShape() || n.Dim == nil || !ct.Without("null").IsArrayLike() {
		return types.Unknown, false
	}
	key, ok := syntax.LiteralArrayKey(n.Dim)
	if !ok {
		return types.Unknown, false
	}
	v, isVar := n.Var.(*syntax.Variable)
	if !isVar || v.Name == "" || v.Name == "this" {
		v = nil
	}
	return e.shapeKeyOf(ct, key, v)
}

// shapeKeyOf types key of a value of type ct from its shape, widened by
// the element writes into variable v (nil: not a variable) reaching it; ok
// is false when the shape does not list the key and no write adds it.
func (e *Env) shapeKeyOf(ct types.Type, key string, v *syntax.Variable) (types.Type, bool) {
	isVar := v != nil
	kt, ok := ct.ShapeKey(key)
	if !ok {
		if !ct.IsSealedShape() {
			return types.Unknown, false
		}
		// A key missing from a literal array may be added by a write
		// (`$a = []; $a['k'] = 1;`).
		// Absent from the literal: only the writes that may store into the
		// key give it a type (the same literal key, or a computed one, and
		// the element a nested write creates: `$rows = ['total' => 1];
		// $rows['meta']['s'] = 1;` makes `$rows['meta']` an int[], not an
		// int); otherwise it is unknown, never the type of the other
		// elements.
		// On an empty literal (`$a = []; $a[] = $x;`) an integer key may
		// also hold an appended value (or be missing, read as for `T[]`).
		appends := len(ct.ShapeKeys()) == 0 && types.IsIntKey(key)
		if isVar && e.mayWriteKey(v, key, appends) {
			return e.widenKey(types.Of("never"), v, key, appends), true
		}
		return types.Unknown, true
	}
	if isVar {
		kt = e.widenKey(kt, v, key, false)
	}
	return kt, true
}

// mayWriteKey reports whether a write reaching variable v may store a
// known value into its element `key`: a direct write with that literal key
// or with a computed key, or an append when appends is set, and no nested,
// unknown or (for an integer key, without appends) appending write that
// would leave its value unknown.
func (e *Env) mayWriteKey(v *syntax.Variable, key string, appends bool) bool {
	ws, _ := e.reachingWrites(v)
	may := false
	for _, w := range ws {
		if w.nested {
			if k, ok := syntax.LiteralArrayKey(w.key); !ok || k == key {
				may = true // it may create the element
			}
			continue
		}
		d := w.dim()
		switch k, ok := syntax.LiteralArrayKey(d); {
		case !w.known():
			return false
		case d == nil && appends:
			may = true
		case d == nil && types.IsIntKey(key):
			return false
		case d == nil || (ok && k != key):
		default:
			may = true
		}
	}
	return may
}

// widenKey unions the type kt of key `key` of variable v with the values
// written into that key by the writes reaching v: writes to the same
// literal key or to a computed key, and appends when appends is set. A
// nested write into that key (`$a['key'][…] = …`) gives the element it
// creates when the key had none, else makes it unknown, as a destructuring
// write does.
func (e *Env) widenKey(kt types.Type, v *syntax.Variable, key string, appends bool) types.Type {
	ws, back := e.reachingWrites(v)
	if len(ws) == 0 {
		return kt
	}
	ts := []types.Type{kt}
	var nested []types.Type
	for i, w := range ws {
		if w.nested {
			if k, ok := syntax.LiteralArrayKey(w.key); !ok || k == key {
				nested = append(nested, e.nestedElementType(w))
			}
			continue
		}
		if !w.known() {
			return types.Unknown
		}
		d := w.dim()
		if d == nil && !appends {
			continue // `$a[] = v` appends a new key
		}
		if k, ok := syntax.LiteralArrayKey(d); d != nil && ok && k != key {
			continue
		}
		if w.a != nil && w.a.ByRef {
			return types.Unknown
		}
		t := e.elementStoredType(w)
		if t.IsUnknown() && back[i] {
			continue // a back-edge cycle adds nothing (as for variables)
		}
		ts = append(ts, t)
	}
	if len(nested) > 0 {
		if created, ok := createdByNested(ts, nested); ok {
			return created
		}
		return types.Unknown
	}
	return types.Union(ts...)
}

// createdByNested is the element nested writes create (see
// nestedElementType) when they are its only source: ts holds the element
// type before the writes, which "never" says is absent, and nothing else
// stores into it. A nested write into an element that may exist changes it
// in place, which no union of types describes.
func createdByNested(ts, nested []types.Type) (types.Type, bool) {
	if len(ts) != 1 || !ts[0].OnlyOf("never") {
		return types.Type{}, false
	}
	return types.Union(nested...), true
}

// shapeElem is the union of the values of a sealed shape (unknown when the
// type has no sealed shape), widened by the element writes into variable x
// reaching it (when x is one). A sealed empty shape (`[]`) gives the union
// of those writes, unknown without any.
func (e *Env) shapeElem(t types.Type, x syntax.Expr) types.Type {
	if !t.IsSealedShape() {
		return types.Unknown
	}
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	isVar := ok && v.Name != "" && v.Name != "this"
	keys := t.ShapeKeys()
	if len(keys) == 0 {
		if !isVar {
			return types.Unknown
		}
		if ws, _ := e.reachingWrites(v); len(ws) == 0 {
			return types.Unknown
		}
		el := e.widenVarElem(types.Of("never"), v)
		if el.OnlyOf("never") {
			return types.Unknown
		}
		return el
	}
	ts := make([]types.Type, 0, len(keys))
	for _, k := range keys {
		ts = append(ts, k.Type)
	}
	el := types.Union(ts...)
	if isVar {
		el = e.widenVarElem(el, v)
	}
	return el
}

// foreachElem is the value type of `foreach (x as $v)`: the element type of
// a `T[]`-only collection, else the union of a sealed shape's values; both
// widened by the element writes into variable x reaching the loop.
func (e *Env) foreachElem(x syntax.Expr) types.Type {
	t := e.baseType(x)
	if el := iterElem(t); !el.IsUnknown() {
		if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
			return e.widenVarElem(el, v)
		}
		return el
	}
	if !t.Without("null", "false").IsArrayLike() {
		// Traversable<K, V>, iterable<V>, possibly with `T[]` members.
		_, v := e.iterTypes(t)
		return v
	}
	return e.shapeElem(t, x)
}

// foreachKey is the key type of `foreach (x as $k => …)`: int|string, or
// only the kinds of keys a sealed shape lists plus those the element writes
// into variable x reaching the loop may add.
func (e *Env) foreachKey(x syntax.Expr) types.Type {
	anyKey := types.Of("int", "string")
	t := e.baseType(x)
	if !t.Without("null", "false").IsArrayLike() {
		if k, _ := e.iterTypes(t); !k.IsUnknown() {
			return k
		}
		return anyKey
	}
	key := t.ArrayKey()
	if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
		key = e.arrayKeyAt(t, v.Name, v, syntax.EnclosingVariableScope(v))
	}
	if key.IsUnknown() {
		return anyKey
	}
	return key
}

// arrayKeyAt widens an array key domain by the element writes reaching use.
func (e *Env) arrayKeyAt(t types.Type, name string, use, scope syntax.Node) types.Type {
	key := t.ArrayKey()
	if t.IsSealedShape() && len(t.ShapeKeys()) == 0 {
		key = types.Of("never")
	}
	if key.IsUnknown() || e.shapeClobbered(scope, name) {
		return types.Unknown
	}
	ws, _ := e.reachingWritesAt(name, use, scope)
	for _, w := range ws {
		var d syntax.Expr
		switch {
		case w.nested:
			d = w.key
		case !w.known():
			return types.Unknown
		default:
			d = w.dim()
			if d == nil {
				key = types.Union(key, types.Int)
				continue
			}
		}
		k, ok := syntax.LiteralArrayKey(d)
		if !ok {
			return types.Unknown
		}
		kt := types.String
		if types.IsIntKey(k) {
			kt = types.Int
		}
		key = types.Union(key, kt)
	}
	if key.OnlyOf("never") {
		return types.Unknown
	}
	return key
}

// shapeTarget types a destructuring target item from the shape of the
// destructured value.
func shapeTarget(src func() types.Type, key string) func() types.Type {
	return func() types.Type {
		st := src()
		if !st.Without("null").IsArrayLike() {
			return types.Unknown
		}
		if kt, ok := st.ShapeKey(key); ok {
			return kt
		}
		if !st.HasShape() && !st.Has("array") {
			return st.Elem() // `[$a, $b] = $list` with `T[]` elements
		}
		return types.Unknown
	}
}

// ---- mutations (for shape and emptiness facts) ---------------------------------------

type mutKind uint8

const (
	mutAssign  mutKind = iota // the variable is (re)assigned or a foreach/list target
	mutRef                    // bound by reference (`&$x`, foreach by ref, closure use by ref, global/static)
	mutCallArg                // passed as a bare argument: a mutation when the parameter is by-reference
	mutUnset                  // unset($x) / unset($x[…])
	mutPointer                // passed to an internal-pointer function (next, prev, end, each, reset)
	mutCall                   // any call (recorded under anyCall): may change properties of $this
)

// anyCall is the mutations key listing every call of the scope.
const anyCall = "\x00call"

// argKey keys the calls a local is passed to as a bare argument.
func argKey(name string) string { return "\x00arg " + name }

// anyWrite keys every write of a scope (assignment, increment, unset, loop
// target, argument), whatever it changes.
const anyWrite = "\x00write"

// propNameKey is the mutations key listing the writes of property name
// through any receiver ("*" for a name computed at run time).
func propNameKey(name string) string { return "\x00prop " + name }

type mutation struct {
	span syntax.Span
	kind mutKind
	call syntax.Expr // mutCallArg: the call
	arg  *syntax.Arg // mutCallArg: the argument
	idx  int         // mutCallArg: argument position
}

// mutations returns the operations that may change variable name in scope
// (collected once per scope, lazily).
func (e *Env) mutations(scope syntax.Node, name string) []mutation {
	sv := e.scopeVars(scope)
	if sv.muts == nil {
		sv.muts = e.collectMutations(scope)
	}
	return sv.muts[name]
}

func (e *Env) collectMutations(scope syntax.Node) map[string][]mutation {
	out := map[string][]mutation{}
	add := func(x syntax.Expr, m mutation) {
		out[anyWrite] = append(out[anyWrite], m)
		x = syntax.UnwrapParens(x)
		if k := writtenPropName(x); k != "" {
			out[k] = append(out[k], m)
		}
		if d, ok := x.(*syntax.ArrayDimFetch); ok {
			// An element write: recorded under the first-level element's
			// key (dimKey), or dimKey(base, "*") for a computed key.
			for {
				inner, ok := syntax.UnwrapParens(d.Var).(*syntax.ArrayDimFetch)
				if !ok {
					break
				}
				d = inner
			}
			b := narrowKey(syntax.UnwrapParens(d.Var))
			if b != "" && !isDimKey(b) {
				out[dimKey(b, anyElem)] = append(out[dimKey(b, anyElem)], m)
			}
			if k := narrowKey(d); k != "" && len(dimVars(k)) == 0 {
				out[k] = append(out[k], m)
			} else if b != "" && !isDimKey(b) {
				// A computed key (a local one included) may be any element.
				out[dimKey(b, "*")] = append(out[dimKey(b, "*")], m)
			}
			return
		}
		if k := narrowKey(x); k != "" {
			out[k] = append(out[k], m)
		}
	}
	var targets func(x syntax.Expr, span syntax.Span, kind mutKind)
	targets = func(x syntax.Expr, span syntax.Span, kind mutKind) {
		switch t := x.(type) {
		case *syntax.List:
			for _, it := range t.Items {
				if it != nil && it.Value != nil {
					k := kind
					if it.ByRef {
						k = mutRef
					}
					targets(it.Value, span, k)
				}
			}
		case *syntax.Array:
			for _, it := range t.Items {
				if it != nil && it.Value != nil {
					k := kind
					if it.ByRef {
						k = mutRef
					}
					targets(it.Value, span, k)
				}
			}
		default:
			add(x, mutation{span: span, kind: kind})
		}
	}
	args := func(call syntax.Expr, al *syntax.ArgList, pointer bool) {
		if al == nil {
			return
		}
		for i, x := range al.Args {
			a, ok := x.(*syntax.Arg)
			if !ok || a.Unpack {
				continue
			}
			kind := mutCallArg
			if a.ByRef {
				kind = mutRef
			} else if pointer && i == 0 {
				kind = mutPointer
			}
			m := mutation{span: call.Span(), kind: kind, call: call, arg: a, idx: i}
			add(a.Value, m)
			if v, ok := syntax.UnwrapParens(a.Value).(*syntax.Variable); ok && kind == mutCallArg && v.Name != "" && v.NameExpr == nil {
				out[argKey(v.Name)] = append(out[argKey(v.Name)], m)
			}
		}
	}
	var body []syntax.Node
	if scope == nil {
		for _, st := range e.File.Stmts {
			body = append(body, st)
		}
	} else if b := syntax.VariableScopeBody(scope); b != nil {
		body = []syntax.Node{b}
	}
	for _, b := range body {
		syntax.Inspect(b, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Closure:
				if n != scope {
					for _, u := range n.Uses {
						if u.ByRef && u.Var != nil {
							add(u.Var, mutation{span: n.Span(), kind: mutRef})
						}
					}
				}
				return n == scope
			case *syntax.Function, *syntax.Method, *syntax.ArrowFunction, *syntax.PropertyHook, *syntax.ClassLike:
				return n == scope
			case *syntax.Assign:
				kind := mutAssign
				if n.ByRef {
					kind = mutRef
					add(n.Value, mutation{span: n.Span(), kind: mutRef})
				}
				targets(n.Var, n.Span(), kind)
			case *syntax.IncDec:
				add(n.Var, mutation{span: n.Span(), kind: mutAssign})
			case *syntax.Foreach:
				if n.ByRef {
					add(n.Expr, mutation{span: n.Span(), kind: mutRef})
				}
				if n.Key != nil {
					targets(n.Key, n.Span(), mutAssign)
				}
				kind := mutAssign
				if n.ByRef {
					kind = mutRef
				}
				targets(n.Value, n.Span(), kind)
			case *syntax.Unset:
				for _, x := range n.Vars {
					for {
						d, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
						if !ok {
							break
						}
						x = d.Var
					}
					add(x, mutation{span: n.Span(), kind: mutUnset})
				}
			case *syntax.Global:
				for _, x := range n.Vars {
					add(x, mutation{span: n.Span(), kind: mutRef})
				}
			case *syntax.StaticStmt:
				for _, sv := range n.Vars {
					if sv.Var != nil {
						add(sv.Var, mutation{span: n.Span(), kind: mutRef})
					}
				}
			case *syntax.FuncCall:
				pointer := false
				if nm, ok := n.Name.(*syntax.Name); ok {
					switch strings.ToLower(strings.TrimPrefix(nm.Value, `\`)) {
					case "next", "prev", "end", "each", "reset":
						pointer = true
					}
				}
				args(n, n.Args, pointer)
				out[anyCall] = append(out[anyCall], mutation{span: n.Span(), kind: mutCall, call: n})
			case *syntax.MethodCall:
				args(n, n.Args, false)
				out[anyCall] = append(out[anyCall], mutation{span: n.Span(), kind: mutCall, call: n})
			case *syntax.StaticCall:
				args(n, n.Args, false)
				out[anyCall] = append(out[anyCall], mutation{span: n.Span(), kind: mutCall, call: n})
			case *syntax.New:
				args(n, n.Args, false)
				out[anyCall] = append(out[anyCall], mutation{span: n.Span(), kind: mutCall, call: n})
			}
			return true
		})
	}
	return out
}

// writtenPropName is the propNameKey of the property a write to x changes
// (`$q->p`, `$q->p['k']`), or "".
func writtenPropName(x syntax.Expr) string {
	for {
		d, ok := x.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		x = syntax.UnwrapParens(d.Var)
	}
	pf, ok := x.(*syntax.PropertyFetch)
	if !ok {
		return ""
	}
	if id, named := pf.Name.(*syntax.Identifier); named {
		return propNameKey(id.Value)
	}
	return propNameKey("*")
}

// byRefArg reports whether the argument of mutation m binds a by-reference
// parameter of a resolvable callee.
func (e *Env) byRefArg(m mutation) bool {
	var params [][]index.Param
	// unresolved: some callee the call may reach is unknown. Where
	// declarations are proofs, it may take the argument by reference.
	unresolved := false
	switch c := m.call.(type) {
	case *syntax.FuncCall:
		if f := e.ResolveFunction(c); f != nil {
			params = append(params, f.Params)
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return e.native
		}
		recv := e.TypeOf(c.Var)
		unresolved = recv.HasAny("mixed", "object")
		for _, cls := range recv.Classes() {
			if mm := e.Index.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, e.PHP); mm != nil {
				params = append(params, mm.Params)
			} else {
				unresolved = true
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return e.native
		}
		if cls := e.classRef(c.Class); cls != "" {
			if mm := e.Index.FindMethod(cls, id.Value, e.PHP); mm != nil {
				params = append(params, mm.Params)
			}
		}
	case *syntax.New:
		cls := e.classRef(c.Class)
		if cls == "" {
			return e.native
		}
		if mm := e.Index.FindMethod(cls, "__construct", e.PHP); mm != nil {
			params = append(params, mm.Params)
		} else if e.Index.Class(cls, e.PHP) != nil {
			return false // no constructor: the arguments are not bound
		}
	}
	if e.native && (unresolved || len(params) == 0) {
		return true
	}
	for _, ps := range params {
		if p, ok := paramFor(ps, m.arg, m.idx); ok && p.ByRef {
			return true
		}
	}
	return false
}

func paramFor(ps []index.Param, a *syntax.Arg, idx int) (index.Param, bool) {
	if a.Name != nil {
		for _, p := range ps {
			if strings.TrimPrefix(p.Name, "$") == a.Name.Value {
				return p, true
			}
		}
		return index.Param{}, false
	}
	if idx >= 0 && idx < len(ps) {
		return ps[idx], true
	}
	if idx >= 0 && len(ps) > 0 && ps[len(ps)-1].Variadic {
		return ps[len(ps)-1], true
	}
	return index.Param{}, false
}

// applies reports whether m may change the variable's array value.
func (e *Env) applies(m mutation) bool {
	switch m.kind {
	case mutCallArg:
		return e.byRefArg(m)
	case mutCall:
		// Builtin functions do not reach $this (by-reference arguments are
		// recorded separately).
		if c, ok := m.call.(*syntax.FuncCall); ok {
			return !e.isBuiltinCall(c)
		}
	}
	return true
}

// isBuiltinCall reports a call resolving to a builtin function (cached:
// brokenBy asks for the same calls on every read of a property).
func (e *Env) isBuiltinCall(c *syntax.FuncCall) bool {
	if b, ok := e.builtinCalls[c]; ok {
		return b
	}
	f := e.ResolveFunction(c)
	b := f != nil && stubs.Index().Function(f.FQN, e.PHP) == f
	if e.builtinCalls == nil {
		e.builtinCalls = map[*syntax.FuncCall]bool{}
	}
	e.builtinCalls[c] = b
	return b
}

// shapeClobbered reports whether variable name may be modified in scope in
// a way element-write tracking does not see: bound by reference or passed
// to a by-reference parameter (sort(), array_shift(), …).
func (e *Env) shapeClobbered(scope syntax.Node, name string) bool {
	for _, m := range e.mutations(scope, name) {
		switch m.kind {
		case mutRef:
			return true
		case mutCallArg:
			if e.byRefArg(m) {
				return true
			}
		}
	}
	return false
}

// shapeClobberedAt is shapeClobbered for array facts established at from
// and read at use: a by-reference argument only clobbers them between the
// two (or in a loop around use), a reference binding anywhere does (writes
// through the reference may happen at any point).
func (e *Env) shapeClobberedAt(scope syntax.Node, name string, from uint32, use syntax.Node) bool {
	var calls []mutation
	for _, m := range e.mutations(scope, name) {
		switch m.kind {
		case mutRef:
			return true
		case mutCallArg:
			calls = append(calls, m)
		}
	}
	return e.brokenBy(scope, from, use, calls)
}

// nonEmptyBroken reports whether a non-empty fact established at position
// from (a definition or a condition) may no longer hold at use: a mutation
// of the variable lies between them (outside branches exclusive with use),
// or the use sits in a loop (entered after from) that mutates it. Moving
// the internal pointer does not count.
func (e *Env) nonEmptyBroken(scope syntax.Node, name string, from uint32, use syntax.Node) bool {
	if strings.Contains(name, "->") || strings.Contains(name, "::") {
		// A property may also be changed by any (non-builtin) call.
		return e.brokenBy(scope, from, use, e.mutations(scope, name), e.mutations(scope, anyCall))
	}
	return e.brokenBy(scope, from, use, e.mutations(scope, name))
}

// maxMutScan caps the mutations brokenBy examines for one fact (those
// between the fact and its use, and in loops around the use): beyond it
// the fact is considered broken. A property read after thousands of calls
// made each read linear (quadratic bodies); real code stays far below.
const maxMutScan = 512

// brokenBy reports whether one of the mutations in lists (each in source
// order, as collectMutations records them) may change a fact established
// at position from before use (see nonEmptyBroken).
func (e *Env) brokenBy(scope syntax.Node, from uint32, use syntax.Node, lists ...[]mutation) bool {
	empty := true
	for _, l := range lists {
		empty = empty && len(l) == 0
	}
	if empty {
		return false
	}
	at := use.Span().Start
	var exclusive []syntax.Span // branches that cannot run on the way to use
	var loops []syntax.Span     // loops around use entered after from
	child := use
	for p := use.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		switch n := p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if n.Span().Start >= from {
				loops = append(loops, n.Span())
			}
		case *syntax.If:
			branches := []syntax.Node{n.Body}
			for _, ei := range n.ElseIfs {
				branches = append(branches, ei)
			}
			if n.Else != nil {
				branches = append(branches, n.Else)
			}
			for _, b := range branches {
				if b != nil && b != child {
					exclusive = append(exclusive, b.Span())
				}
			}
		case *syntax.Ternary:
			if n.Then != nil && child == syntax.Node(n.Else) {
				exclusive = append(exclusive, n.Then.Span())
			}
		}
	}
	// Only mutations starting in [from, limit) can matter: before use, or
	// later in a loop around use entered after from.
	limit := at
	for _, l := range loops {
		limit = max(limit, l.End)
	}
	scanned := 0
	for _, muts := range lists {
		lo, _ := slices.BinarySearchFunc(muts, from, func(m mutation, from uint32) int { return cmp.Compare(m.span.Start, from) })
		for _, m := range muts[lo:] {
			if m.span.Start >= limit {
				break
			}
			if scanned++; scanned > maxMutScan {
				return true
			}
			if e.mutBreaks(m, at, loops, exclusive) {
				return true
			}
		}
	}
	return false
}

// mutBreaks reports whether mutation m, after a fact was established and
// before its use at (see brokenBy), may change the fact.
func (e *Env) mutBreaks(m mutation, at uint32, loops, exclusive []syntax.Span) bool {
	if m.kind == mutPointer {
		return false
	}
	// brokenBy passes mutations starting in [from, limit): those at or
	// after use lie in a loop around it.
	inLoop := inSpans(m.span, loops)
	if !inLoop && m.span.End > at && m.kind != mutUnset {
		return false // the use is inside the mutating expression (`array_shift($x)`, `$x = f($x)`)
	}
	if inSpans(m.span, exclusive) {
		return false
	}
	return e.applies(m)
}

// inSpans reports whether s starts inside one of spans.
func inSpans(s syntax.Span, spans []syntax.Span) bool {
	for _, x := range spans {
		if s.Start >= x.Start && s.Start < x.End {
			return true
		}
	}
	return false
}

// pointerMoved reports whether the internal pointer of variable name may
// have been moved in scope (next/prev/end/each, or by-reference binding).
func (e *Env) pointerMoved(scope syntax.Node, name string) bool {
	for _, m := range e.mutations(scope, name) {
		if m.kind == mutPointer || m.kind == mutRef {
			return true
		}
	}
	return false
}
