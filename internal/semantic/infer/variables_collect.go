package infer

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Collection of a scope's variable definitions and element writes.

// scopeVars collects variable definitions (assignments, params, foreach,
// catch, inline @var) of one scope, in source order.
func (e *Env) scopeVars(scope syntax.Node) *scopeVars {
	if sv, ok := e.scopes[scope]; ok {
		return sv
	}
	sv := &scopeVars{defs: map[string][]varDef{}, elemWrites: map[string][]varDef{}}
	e.scopes[scope] = sv
	add := func(name string, pos uint32, t func() types.Type) {
		if name != "" {
			sv.defs[name] = append(sv.defs[name], varDef{pos: pos, typ: t})
		}
	}
	var params []*syntax.Param
	var body []syntax.Node
	switch s := scope.(type) {
	case *syntax.Function:
		params, body = s.Params, []syntax.Node{s.Body}
	case *syntax.PropertyHook:
		params = s.Params
		if s.Body != nil {
			body = []syntax.Node{s.Body}
		}
		if strings.EqualFold(s.Name.Value, "set") && len(s.Params) == 0 {
			add("value", s.Span().Start, func() types.Type { return e.hookValueType(s) })
		}
	case *syntax.Method:
		params = s.Params
		if s.Body != nil {
			body = []syntax.Node{s.Body}
		}
	case *syntax.Closure:
		params, body = s.Params, []syntax.Node{s.Body}
		for _, u := range s.Uses {
			add(u.Var.Name, u.Span().Start, func() types.Type {
				if u.ByRef {
					return types.Unknown
				}
				return e.captureType(s, u.Var.Name)
			})
		}
	case *syntax.ArrowFunction:
		params, body = s.Params, []syntax.Node{s.Expr}
		captured := map[string]bool{}
		for _, p := range s.Params {
			captured[p.Var.Name] = true // parameters shadow implicit imports
		}
		syntax.Inspect(s.Expr, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
				return false
			case *syntax.Variable:
				if n.Name != "" && n.Name != "this" && !captured[n.Name] {
					name := n.Name
					captured[name] = true
					add(name, s.Span().Start, func() types.Type { return e.captureType(s, name) })
				}
			}
			return true
		})
	case nil:
		for _, st := range e.File.Stmts {
			body = append(body, st)
		}
	}
	for _, p := range params {
		add(p.Var.Name, p.Span().Start, func() types.Type { return e.paramType(scope, p) })
		if p.ByRef {
			sv.shared = mark(sv.shared, p.Var.Name)
		}
	}
	// body holds no nil node: the parser always builds function and closure
	// bodies (parseBlock) and arrow function expressions (BadExpr at worst);
	// an abstract method's missing body is not added.
	for _, b := range body {
		syntax.Inspect(b, func(n syntax.Node) bool {
			e.noteDynamic(n, sv)
			e.noteRefs(n, sv)
			switch n := n.(type) {
			case *syntax.Closure:
				if n != scope {
					// A closure importing a variable by reference may write it
					// whenever it is called: its type is unknown from here on.
					for _, u := range n.Uses {
						if u.ByRef && u.Var != nil {
							add(u.Var.Name, n.Span().End, func() types.Type { return types.Unknown })
						}
					}
				}
				return n == scope
			case *syntax.Function, *syntax.Method, *syntax.ArrowFunction, *syntax.PropertyHook, *syntax.ClassLike:
				return n == scope // do not descend into nested scopes
			case *syntax.Assign:
				e.collectDimWrites(n, n, n.Var, sv)
				end := n.Span().End
				var kill syntax.Span
				// Any assignment replaces the value (`.=` is a string, `+=` a
				// number), not only `=` (by reference: not a kill). So does
				// destructuring by value, which writes null where the key is
				// missing.
				if _, plain := n.Var.(*syntax.Variable); (plain || destructuresByValue(n.Var)) && !n.ByRef {
					kill = e.arrowDefinitionSpan(n)
					if es, ok := n.Parent().(*syntax.ExprStmt); ok {
						switch parent := es.Parent().(type) {
						case *syntax.Block:
							kill = parent.Span()
						case *syntax.File:
							kill = syntax.Span{End: uint32(len(e.File.Src))}
						}
					}
				}
				var asg *syntax.Assign
				if v, plain := n.Var.(*syntax.Variable); plain && n.Op.Kind == syntax.TEqual && !n.ByRef && v.NameExpr == nil {
					asg = n
				}
				e.collectAssignTargets(n, func(name string, pos uint32, t func() types.Type) {
					if name != "" {
						if e.optionalAssertion(n) {
							kill = syntax.Span{}
							asg = nil
							t = func() types.Type { return types.Mixed }
						}
						sv.defs[name] = append(sv.defs[name], varDef{pos: pos, end: end, kill: kill, typ: t, asg: asg})
					}
				})
			case *syntax.IncDec:
				e.collectDimMutation(n, n.Var, sv)
				if v := asVariable(n.Var); v != nil {
					sv.defs[v.Name] = append(sv.defs[v.Name], varDef{pos: n.Span().Start, end: n.Span().End, kill: e.statementKill(n), typ: func() types.Type {
						if e.optionalAssertion(n) {
							return types.Mixed
						}
						return e.incDecStored(n)
					}})
				}
			case *syntax.Unset:
				for _, x := range n.Vars {
					if v := asVariable(x); v != nil {
						sv.defs[v.Name] = append(sv.defs[v.Name], varDef{pos: x.Span().Start, end: x.Span().End, kill: e.statementKill(n), typ: func() types.Type { return types.Null }})
					}
				}
			case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
				e.collectOutArgs(n.(syntax.Expr), sv)
			case *syntax.If:
				for _, name := range branchAssigned(n) {
					sv.defs[name] = append(sv.defs[name], varDef{pos: n.Span().Start, barrier: true, kill: joinSpan(n, e.File)})
				}
			case *syntax.Foreach:
				it := n.Expr
				// The loop binds its targets at the start of every iteration:
				// inside the body they hide earlier definitions.
				var kill syntax.Span
				if n.Body != nil {
					kill = n.Body.Span()
				}
				addKill := func(name string, pos uint32, t func() types.Type) {
					if name != "" {
						sv.defs[name] = append(sv.defs[name], varDef{pos: pos, kill: kill, typ: t})
					}
				}
				if n.Key != nil {
					if kv, ok := n.Key.(*syntax.Variable); ok {
						addKill(kv.Name, n.Key.Span().Start, func() types.Type { return e.foreachKey(it) })
					}
				}
				e.collectTargets(n.Value, n.Value.Span().Start, func() types.Type { return e.foreachElem(it) }, addKill)
				collectLoopDimWrites(n, n.Key, sv)
				collectLoopDimWrites(n, n.Value, sv)
			case *syntax.Catch:
				if n.Var != nil {
					var atoms []string
					for _, t := range n.Types {
						atoms = append(atoms, `\`+e.Names.Class(t.Value, t.Span().Start))
					}
					add(n.Var.Name, n.Var.Span().Start, func() types.Type { return types.Of(atoms...) })
				}
			case *syntax.Global:
				// types.Unknown provenance.
				for _, x := range n.Vars {
					if v, ok := x.(*syntax.Variable); ok {
						add(v.Name, v.Span().Start, func() types.Type { return types.Unknown })
						sv.shared = mark(sv.shared, v.Name)
					}
				}
				return false
			case *syntax.StaticStmt:
				for _, s := range n.Vars {
					if s.Var != nil {
						add(s.Var.Name, s.Var.Span().Start, func() types.Type { return e.TypeOf(s) })
						sv.statics = mark(sv.statics, s.Var.Name)
					}
				}
				return false
			}
			return true
		})
	}
	// Inline `/** @var Type $name */` comments.
	if !e.native {
		e.inlineVarDocs(scope, func(name string, pos, end uint32, t func() types.Type) {
			sv.defs[name] = append(sv.defs[name], varDef{pos: pos, docEnd: end, typ: t, doc: true})
		})
	}
	if e.native {
		refDefs(sv)
	}
	for name, defs := range sv.defs {
		sortDefs(defs)
		sv.defs[name] = defs
	}
	return sv
}

// mark adds name to set m, created on first use.
func mark(m map[string]bool, name string) map[string]bool {
	if m == nil {
		m = map[string]bool{}
	}
	m[name] = true
	return m
}

// noteRefs records in sv the locals node n binds by reference. Where
// declarations are proofs, a write through the other name changes them
// unseen: two locals bound together see each other's later writes (see
// refDefs), a local bound to an array element is unknown from the binding
// on.
func (e *Env) noteRefs(n syntax.Node, sv *scopeVars) {
	var elem *syntax.Variable
	switch n := n.(type) {
	case *syntax.Assign:
		a, b := asVariable(n.Var), asVariable(n.Value)
		if !n.ByRef || a == nil || b == nil || a.Name == "" || b.Name == "" {
			return
		}
		sv.refs = mark(sv.refs, a.Name)
		sv.refs = mark(sv.refs, b.Name)
		sv.pairs = append(sv.pairs, refPair{a.Name, b.Name, n.Span().End})
		return
	case *syntax.ArrayItem:
		if n.ByRef {
			elem = asVariable(n.Value)
		}
	case *syntax.Foreach:
		if n.ByRef {
			elem = asVariable(n.Value)
		}
	case *syntax.Closure:
		for _, u := range n.Uses {
			if u.ByRef {
				sv.refs = mark(sv.refs, u.Var.Name) // the closure may write it whenever called
			}
		}
		return
	}
	if elem == nil || elem.Name == "" {
		return
	}
	sv.refs = mark(sv.refs, elem.Name)
	if e.native {
		sv.defs[elem.Name] = append(sv.defs[elem.Name], varDef{pos: n.Span().End, typ: func() types.Type { return types.Unknown }})
	}
}

// refPair records `$a = &$b` ending at pos.
type refPair struct {
	a, b string
	pos  uint32
}

// refDefs makes every write to one local of a reference pair, after the
// binding, a definition of the other too.
func refDefs(sv *scopeVars) {
	for _, p := range sv.pairs {
		for _, d := range [2][2]string{{p.a, p.b}, {p.b, p.a}} {
			for _, def := range sv.defs[d[0]] {
				if def.pos > p.pos {
					sv.defs[d[1]] = append(sv.defs[d[1]], varDef{pos: def.pos, end: def.end, typ: def.typ})
				}
			}
		}
	}
}

// collectDimMutation records ++/-- on array elements. Nested mutations use
// the same conservative first-level invalidation as nested assignments.
func (e *Env) collectDimMutation(n *syntax.IncDec, target syntax.Expr, sv *scopeVars) {
	d, ok := syntax.UnwrapParens(target).(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	inner, nested := innermostDim(d)
	if v := asVariable(inner.Var); v != nil {
		w := &elemWrite{inc: n, nested: nested, key: inner.Dim, stmt: n}
		sv.elemWrites[v.Name] = append(sv.elemWrites[v.Name], varDef{pos: n.Span().Start, end: n.Span().End, w: w})
	}
}

func sortDefs(d []varDef) {
	slices.SortStableFunc(d, func(a, b varDef) int { return cmp.Compare(a.pos, b.pos) })
}

// maxVarDefs caps the definitions (assignments, element writes, inline
// annotations) of one variable in one scope that reads consider; beyond it
// the variable is unknown (element writes: an unknown write). Each read
// walks them, so thousands of assignments to one variable with as many
// reads were quadratic; real code stays far below (same cap as value
// discovery in inspection queries).
const maxVarDefs = 512

// collectTargets registers variables written by an assignment target
// (plain variable or list/array destructuring).
func (e *Env) collectTargets(target syntax.Expr, pos uint32, t func() types.Type, add func(string, uint32, func() types.Type)) {
	switch v := target.(type) {
	case *syntax.Variable:
		add(v.Name, pos, t)
	case *syntax.List:
		e.collectItemTargets(v.Items, pos, t, add)
	case *syntax.Array:
		e.collectItemTargets(v.Items, pos, t, add)
	}
}

// eachWriteTarget calls f with each target write target x assigns: x
// itself, or each target of a destructuring (`[$a, $b['k']]`,
// `list(…)`, nested ones at any depth). It stops, returning false, as soon
// as f returns false.
func eachWriteTarget(x syntax.Expr, f func(syntax.Expr) bool) bool {
	var items []*syntax.ArrayItem
	switch l := x.(type) {
	case *syntax.List:
		items = l.Items
	case *syntax.Array:
		items = l.Items
	default:
		return f(x)
	}
	for _, it := range items {
		if it != nil && it.Value != nil && !eachWriteTarget(it.Value, f) {
			return false
		}
	}
	return true
}

// destructuresByValue reports whether target is a list() or `[...]`
// destructuring pattern none of whose items, at any depth, binds by
// reference: each variable in it is assigned, null when its key is
// missing.
func destructuresByValue(target syntax.Expr) bool {
	var items []*syntax.ArrayItem
	switch l := target.(type) {
	case *syntax.List:
		items = l.Items
	case *syntax.Array:
		items = l.Items
	default:
		return false
	}
	for _, it := range items {
		if it == nil || it.Value == nil {
			continue
		}
		if it.ByRef {
			return false
		}
		switch it.Value.(type) {
		case *syntax.List, *syntax.Array:
			if !destructuresByValue(it.Value) {
				return false
			}
		}
	}
	return true
}

// collectAssignTargets registers the targets of assignment n. Destructuring
// a variable (`[$a, $b] = $v`) reads its keys as `$v[0]`, `$v[1]` would:
// from its base type, with the element writes reaching n applied key by key.
func (e *Env) collectAssignTargets(n *syntax.Assign, add func(string, uint32, func() types.Type)) {
	t := func() types.Type { return e.TypeOf(n) }
	var items []*syntax.ArrayItem
	switch l := n.Var.(type) {
	case *syntax.List:
		items = l.Items
	case *syntax.Array:
		items = l.Items
	}
	v, ok := syntax.UnwrapParens(n.Value).(*syntax.Variable)
	if items == nil || !ok || v.Name == "" || v.Name == "this" || n.Op.Kind != syntax.TEqual {
		e.collectTargets(n.Var, n.Span().Start, t, add)
		return
	}
	e.collectItems(items, n.Span().Start, func(key string) func() types.Type {
		return func() types.Type {
			ct := e.baseType(v)
			if !ct.Without("null").IsArrayLike() {
				return types.Unknown
			}
			if !ct.HasShape() && !ct.Has("array") {
				return e.widenVarElem(ct.Elem(), v) // `T[]` elements
			}
			kt, _ := e.shapeKeyOf(ct, key, v)
			return kt
		}
	}, add)
}

// collectItemTargets registers destructuring items; an item reads the
// destructured value's shape key (explicit literal key or position).
func (e *Env) collectItemTargets(items []*syntax.ArrayItem, pos uint32, t func() types.Type, add func(string, uint32, func() types.Type)) {
	e.collectItems(items, pos, func(key string) func() types.Type { return shapeTarget(t, key) }, add)
}

// collectItems registers destructuring items, typing the item of a
// literal key (explicit or positional) with keyType.
func (e *Env) collectItems(items []*syntax.ArrayItem, pos uint32, keyType func(string) func() types.Type, add func(string, uint32, func() types.Type)) {
	for i, it := range items {
		if it == nil || it.Value == nil {
			continue
		}
		key, ok := strconv.Itoa(i), it.Key == nil
		if it.Key != nil {
			key, ok = syntax.LiteralArrayKey(it.Key)
		}
		item := func() types.Type { return types.Unknown }
		if ok && !it.ByRef {
			item = keyType(key)
		}
		e.collectTargets(it.Value, pos, item, add)
	}
}

// collectDimWrites records the element writes `$x[...] = v` made by target,
// part of assignment outer (destructuring targets are recorded as unknown
// writes: a is nil).
func (e *Env) collectDimWrites(outer, a *syntax.Assign, target syntax.Expr, sv *scopeVars) {
	eachOffsetWrite(target, func(d, inner *syntax.ArrayDimFetch, nested bool) bool {
		if v := writtenVar(inner); v != nil {
			w := &elemWrite{stmt: outer}
			if d == target {
				w.a = a
			}
			if nested {
				// Remember which first-level key it changes.
				w.nested, w.key = true, inner.Dim
			}
			sv.elemWrites[v.Name] = append(sv.elemWrites[v.Name], varDef{pos: outer.Span().Start, end: outer.Span().End, w: w})
		}
		return true
	})
}

// collectLoopDimWrites records the element writes target, the key or
// value target of loop fe (`foreach ($xs as $a['k'])`), makes on each
// iteration, as unknown writes.
func collectLoopDimWrites(fe *syntax.Foreach, target syntax.Expr, sv *scopeVars) {
	eachOffsetWrite(target, func(d, inner *syntax.ArrayDimFetch, nested bool) bool {
		if v := writtenVar(inner); v != nil {
			w := &elemWrite{loop: fe, nested: nested, key: inner.Dim}
			sv.elemWrites[v.Name] = append(sv.elemWrites[v.Name], varDef{pos: d.Span().Start, end: d.Span().End, w: w})
		}
		return true
	})
}

// eachOffsetWrite calls f with each write into an offset that write target
// x makes (x itself, or a target of a destructuring; see eachWriteTarget):
// d is the written element, inner the first-level one (whose array is
// written: d itself unless nested). It stops, returning false, as soon as f
// returns false.
func eachOffsetWrite(x syntax.Expr, f func(d, inner *syntax.ArrayDimFetch, nested bool) bool) bool {
	return eachWriteTarget(x, func(t syntax.Expr) bool {
		d, ok := t.(*syntax.ArrayDimFetch)
		if !ok {
			return true
		}
		inner, nested := innermostDim(d)
		return f(d, inner, nested)
	})
}

// innermostDim returns the first-level element of element d (`$a['k']` of
// `$a['k']['l']`, parentheses aside: `(($a)['k'])['l']`), and whether it is
// not d itself.
func innermostDim(d *syntax.ArrayDimFetch) (*syntax.ArrayDimFetch, bool) {
	inner := d
	for {
		parent, ok := syntax.UnwrapParens(inner.Var).(*syntax.ArrayDimFetch)
		if !ok {
			return inner, inner != d
		}
		inner = parent
	}
}

// writtenVar returns the variable whose element inner is, parentheses
// aside (`($a)['k']`), when named statically (nil otherwise).
func writtenVar(inner *syntax.ArrayDimFetch) *syntax.Variable {
	if v, ok := syntax.UnwrapParens(inner.Var).(*syntax.Variable); ok && v.Name != "" {
		return v
	}
	return nil
}
