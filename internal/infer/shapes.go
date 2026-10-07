package infer

import (
	"strconv"
	"strings"

	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
	"custos/internal/types"
)

// ---- array shapes ---------------------------------------------------------------------

// literalKey returns the canonical array key of a literal key expression
// (integer or plain string literal, optionally negated): "0", "-1", "name".
func literalKey(x syntax.Expr) (string, bool) {
	switch n := unparen(x).(type) {
	case *syntax.Literal:
		switch n.LitKind {
		case syntax.LitInt:
			v, err := strconv.ParseInt(strings.ReplaceAll(n.Raw, "_", ""), 0, 64)
			if err != nil {
				return "", false
			}
			return strconv.FormatInt(v, 10), true
		case syntax.LitString:
			s, ok := plainString(n.Raw)
			if !ok {
				return "", false
			}
			return s, true // "12" is the integer key 12, same canonical text
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TMinus {
			if l, ok := unparen(n.Expr).(*syntax.Literal); ok && l.LitKind == syntax.LitInt {
				if k, ok := literalKey(l); ok && k != "0" {
					return "-" + k, true
				}
			}
		}
	}
	return "", false
}

// plainString decodes a quoted string literal without escapes that need
// interpretation beyond `\\` and `\'` in single quotes.
func plainString(raw string) (string, bool) {
	if len(raw) > 0 && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if len(raw) < 2 || raw[len(raw)-1] != raw[0] {
		return "", false
	}
	body := raw[1 : len(raw)-1]
	switch raw[0] {
	case '\'':
		if strings.IndexByte(body, '\\') < 0 {
			return body, true
		}
		var b strings.Builder
		for i := 0; i < len(body); i++ {
			if body[i] == '\\' && i+1 < len(body) && (body[i+1] == '\\' || body[i+1] == '\'') {
				i++
			}
			b.WriteByte(body[i])
		}
		return b.String(), true
	case '"':
		if strings.IndexByte(body, '\\') < 0 && strings.IndexByte(body, '$') < 0 {
			return body, true
		}
	}
	return "", false
}

// arrayType types an array literal: the atoms are those of its values (a
// single `T[]` when uniform, `array` otherwise), plus a shape when every key
// is a literal (or implicit) and non-emptiness when it has an element.
func (e *Env) arrayType(n *syntax.Array) types.Type {
	if len(n.Items) == 0 {
		return types.Array.WithShape(nil, true)
	}
	elems := make([]types.Type, 0, len(n.Items))
	nonEmpty, plain := false, true
	for _, it := range n.Items {
		if it == nil || it.Value == nil || it.Unpack {
			plain = false
			continue
		}
		nonEmpty = true
		elems = append(elems, e.TypeOf(it.Value))
	}
	if !plain {
		return types.Array.WithNonEmpty(nonEmpty)
	}
	var t types.Type
	u := types.Union(elems...)
	if u.IsUnknown() || len(u.Atoms()) != 1 || strings.HasSuffix(u.Atoms()[0], "[]") {
		t = types.Array
	} else {
		t = types.Of(u.Atoms()[0] + "[]").WithElem(u)
	}
	if len(n.Items) > types.MaxShapeKeys {
		return t.WithNonEmpty(true)
	}
	keys := make([]types.ShapeKey, 0, len(n.Items))
	next, hasInt := int64(0), false
	for i, it := range n.Items {
		name := ""
		if it.Key == nil {
			name = strconv.FormatInt(next, 10)
		} else {
			k, ok := literalKey(it.Key)
			if !ok {
				return t.WithNonEmpty(true)
			}
			name = k
		}
		if v, err := strconv.ParseInt(name, 10, 64); err == nil && strconv.FormatInt(v, 10) == name {
			switch {
			case !hasInt && v < 0 && e.PHP.Below(phpver.PHP83):
				next = 0 // before 8.3 a negative first key is followed by 0
			case !hasInt || v >= next:
				next = v + 1
			}
			hasInt = true
		}
		ty := elems[i]
		replaced := false
		for j := range keys {
			if keys[j].Name == name {
				keys[j].Type, replaced = ty, true
			}
		}
		if !replaced {
			keys = append(keys, types.ShapeKey{Name: name, Type: ty})
		}
	}
	return t.WithShape(keys, true)
}

// shapeDim types `$a['key']` from the shape of $a when the key is a literal
// listed by the shape; ok is false otherwise (callers fall back to the
// element type).
func (e *Env) shapeDim(ct types.Type, n *syntax.ArrayDimFetch) (types.Type, bool) {
	if !ct.HasShape() || n.Dim == nil || !ct.Without("null").IsArrayLike() {
		return types.Unknown, false
	}
	key, ok := literalKey(n.Dim)
	if !ok {
		return types.Unknown, false
	}
	v, isVar := n.Var.(*syntax.Variable)
	isVar = isVar && v.Name != "" && v.Name != "this"
	kt, ok := ct.ShapeKey(key)
	if !ok {
		// A key missing from a literal array may be added by a write
		// (`$a = []; $a['k'] = 1;`).
		if !isVar || !ct.IsSealedShape() || !e.writesKey(v, key) {
			return types.Unknown, false
		}
		return e.widenKey(types.Of("never"), v, key), true
	}
	if isVar {
		kt = e.widenKey(kt, v, key)
	}
	return kt, true
}

// writesKey reports whether variable v's element `key` is written with that
// literal key by a write reaching v.
func (e *Env) writesKey(v *syntax.Variable, key string) bool {
	ws, _ := e.reachingWrites(v)
	for _, w := range ws {
		if d := w.dim(); d != nil {
			if k, ok := literalKey(d); ok && k == key {
				return true
			}
		}
	}
	return false
}

// widenKey unions the type kt of key `key` of variable v with the values
// written into that key by the writes reaching v: writes to the same
// literal key or to a computed key; a nested write (`$a['key'][…] = …`) or
// a destructuring write makes it unknown.
func (e *Env) widenKey(kt types.Type, v *syntax.Variable, key string) types.Type {
	ws, back := e.reachingWrites(v)
	if len(ws) == 0 {
		return kt
	}
	ts := []types.Type{kt}
	for i, w := range ws {
		if w.nested {
			if k, ok := literalKey(w.key); !ok || k == key {
				return types.Unknown
			}
			continue
		}
		if w.a == nil {
			return types.Unknown
		}
		d := w.dim()
		if d == nil {
			continue // `$a[] = v` appends a new key
		}
		if k, ok := literalKey(d); ok && k != key {
			continue
		}
		if w.a.ByRef {
			return types.Unknown
		}
		t := e.writtenType(w.a)
		if t.IsUnknown() && back[i] {
			continue // a back-edge cycle adds nothing (as for variables)
		}
		ts = append(ts, t)
	}
	return types.Union(ts...)
}

// shapeElem is the union of the values of a sealed shape (unknown when the
// type has no sealed shape), widened by the element writes into variable x
// reaching it (when x is one). A sealed empty shape (`[]`) gives the union
// of those writes, unknown without any or with a nested write.
func (e *Env) shapeElem(t types.Type, x syntax.Expr) types.Type {
	if !t.IsSealedShape() {
		return types.Unknown
	}
	v, ok := unparen(x).(*syntax.Variable)
	isVar := ok && v.Name != "" && v.Name != "this"
	keys := t.ShapeKeys()
	if len(keys) == 0 {
		if !isVar {
			return types.Unknown
		}
		ws, _ := e.reachingWrites(v)
		if len(ws) == 0 {
			return types.Unknown
		}
		for _, w := range ws {
			if w.nested {
				return types.Unknown
			}
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
	t := e.TypeOf(x)
	if el := iterElem(t); !el.IsUnknown() {
		if v, ok := unparen(x).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
			return e.widenVarElem(el, v)
		}
		return el
	}
	if !t.Without("null", "false").IsArrayLike() {
		return types.Unknown
	}
	return e.shapeElem(t, x)
}

// foreachKey is the key type of `foreach (x as $k => …)`: int|string, or
// only the kinds of keys a sealed shape lists plus those the element writes
// into variable x reaching the loop may add.
func (e *Env) foreachKey(x syntax.Expr) types.Type {
	anyKey := types.Of("int", "string")
	t := e.TypeOf(x)
	if !t.Without("null", "false").IsArrayLike() || !t.IsSealedShape() {
		return anyKey
	}
	hasInt, hasStr := false, false
	kind := func(k string) {
		if types.IsIntKey(k) {
			hasInt = true
		} else {
			hasStr = true
		}
	}
	for _, k := range t.ShapeKeys() {
		kind(k.Name)
	}
	if v, ok := unparen(x).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
		ws, _ := e.reachingWrites(v)
		for _, w := range ws {
			var d syntax.Expr
			switch {
			case w.nested:
				d = w.key
			case w.a == nil:
				return anyKey
			default:
				d = w.dim()
				if d == nil {
					hasInt = true // `$x[] = v` appends an integer key
					continue
				}
			}
			k, ok := literalKey(d)
			if !ok {
				return anyKey
			}
			kind(k)
		}
	}
	switch {
	case hasInt && hasStr, !hasInt && !hasStr:
		return anyKey
	case hasInt:
		return types.Int
	}
	return types.String
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
		if k := narrowKey(unparen(x)); k != "" {
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
			add(a.Value, mutation{span: call.Span(), kind: kind, call: call, arg: a, idx: i})
		}
	}
	var body []syntax.Node
	switch s := scope.(type) {
	case *syntax.Function:
		body = []syntax.Node{s.Body}
	case *syntax.Method:
		if s.Body != nil {
			body = []syntax.Node{s.Body}
		}
	case *syntax.Closure:
		body = []syntax.Node{s.Body}
	case *syntax.ArrowFunction:
		body = []syntax.Node{s.Expr}
	case nil:
		for _, st := range e.File.Stmts {
			body = append(body, st)
		}
	}
	for _, b := range body {
		if b == nil {
			continue
		}
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
			case *syntax.Function, *syntax.Method, *syntax.ArrowFunction, *syntax.ClassLike:
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
						d, ok := unparen(x).(*syntax.ArrayDimFetch)
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

// byRefArg reports whether the argument of mutation m binds a by-reference
// parameter of a resolvable callee.
func (e *Env) byRefArg(m mutation) bool {
	var params [][]index.Param
	switch c := m.call.(type) {
	case *syntax.FuncCall:
		if f := e.ResolveFunction(c); f != nil {
			params = append(params, f.Params)
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		for _, cls := range e.TypeOf(c.Var).Classes() {
			if mm := e.Index.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, e.PHP); mm != nil {
				params = append(params, mm.Params)
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		if cls := e.classRef(c.Class); cls != "" {
			if mm := e.Index.FindMethod(cls, id.Value, e.PHP); mm != nil {
				params = append(params, mm.Params)
			}
		}
	case *syntax.New:
		if cls := e.classRef(c.Class); cls != "" {
			if mm := e.Index.FindMethod(cls, "__construct", e.PHP); mm != nil {
				params = append(params, mm.Params)
			}
		}
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
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), a.Name.Value) {
				return p, true
			}
		}
		return index.Param{}, false
	}
	if idx < len(ps) {
		return ps[idx], true
	}
	if len(ps) > 0 && ps[len(ps)-1].Variadic {
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
			if f := e.ResolveFunction(c); f != nil && stubs.Index().Function(f.FQN, e.PHP) == f {
				return false
			}
		}
	}
	return true
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

// nonEmptyBroken reports whether a non-empty fact established at position
// from (a definition or a condition) may no longer hold at use: a mutation
// of the variable lies between them (outside branches exclusive with use),
// or the use sits in a loop (entered after from) that mutates it. Moving
// the internal pointer does not count.
func (e *Env) nonEmptyBroken(scope syntax.Node, name string, from uint32, use syntax.Node) bool {
	muts := e.mutations(scope, name)
	if strings.Contains(name, "->") {
		// A property may also be changed by any (non-builtin) call.
		muts = append(append([]mutation(nil), muts...), e.mutations(scope, anyCall)...)
	}
	if len(muts) == 0 {
		return false
	}
	at := use.Span().Start
	var exclusive []syntax.Span // branches that cannot run on the way to use
	var loops []syntax.Span     // loops around use entered after from
	var child syntax.Node = use
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
	inAny := func(s syntax.Span, spans []syntax.Span) bool {
		for _, x := range spans {
			if s.Start >= x.Start && s.Start < x.End {
				return true
			}
		}
		return false
	}
	for _, m := range muts {
		if m.kind == mutPointer {
			continue
		}
		inLoop := inAny(m.span, loops)
		if !inLoop && (m.span.Start < from || m.span.Start >= at) {
			continue
		}
		if !inLoop && m.span.End > at && m.kind != mutUnset {
			continue // the use is inside the mutating expression (`array_shift($x)`, `$x = f($x)`)
		}
		if inAny(m.span, exclusive) {
			continue
		}
		if e.applies(m) {
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
