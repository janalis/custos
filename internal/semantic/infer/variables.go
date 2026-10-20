package infer

import (
	"cmp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

type varDef struct {
	pos uint32
	end uint32 // end of the defining construct (0 = pos); uses inside it do not see it
	// kill is the span of the block in which this definition is an
	// unconditional statement (`$x = v;` directly in a `{}` block): uses
	// later in that block no longer see earlier definitions. Zero: none.
	kill syntax.Span
	typ  func() types.Type
	doc  bool // inline @var annotation: overrides the next definition
	// asg is the plain `$x = value;` assignment making the definition
	// (nil for other kinds); boolean aliases read it (see aliasCond).
	asg *syntax.Assign
	// docEnd is the end of the annotation's comment (doc defs only).
	docEnd uint32
	// barrier marks an if/elseif/else chain whose every branch assigns the
	// variable (or leaves): uses within kill (after the chain, same block)
	// forget definitions made before it. Barriers carry no type.
	barrier bool
	// w marks an element write (see elemDefs); such entries carry no type.
	w *elemWrite
	// arg is the argument passing the variable to a method resolved only
	// when typed (see collectDeferredMethodOut): a by-value parameter
	// leaves the variable as it was (see changing).
	arg *mutation
}

// changing drops from defs the definitions that leave the variable as it
// was: a method argument binding a by-value parameter (whose kill would
// otherwise hide the definitions before it).
func (e *Env) changing(defs []varDef) []varDef {
	if !slices.ContainsFunc(defs, func(d varDef) bool { return d.arg != nil }) {
		return defs
	}
	return slices.DeleteFunc(slices.Clone(defs), func(d varDef) bool { return d.arg != nil && !e.byRefArg(*d.arg) })
}

type scopeVars struct {
	defs map[string][]varDef
	// elemWrites lists, per variable, the writes into its elements (see
	// elemWrite), as definitions positioned at the writing assignment.
	elemWrites map[string][]varDef
	// edefs caches elemDefs.
	edefs map[string][]varDef
	// muts lists the operations that may change each variable (lazy, see
	// mutations()).
	muts map[string][]mutation
	// exits lists the exit regions of the scope (lazy, see exitRegions).
	exits     []exitRegion
	exitsDone bool
	// clobbers holds the positions of extract(), one-argument parse_str(),
	// `$$name = …` writes, eval() and include/require, which may set any local
	// (see noteDynamic); dynamic is set by any of them.
	clobbers []uint32
	dynamic  bool
	// statics holds the static locals of the scope: a call that enters the
	// function again may assign them (see condAt).
	statics map[string]bool
	// shared holds the by-reference parameters and global locals of the
	// scope, which any call may write through the variable they alias;
	// refs the locals bound by reference in the scope (`$a = &$b`, `[&$x]`,
	// `foreach (… as &$v)`, `use (&$x)`), which a write through the other
	// name changes (see condAt).
	shared, refs map[string]bool
	pairs        []refPair
}

// widenElem unions the element type el of the array read by n with every
// value directly written into the same variable's elements by a write that
// can reach the read, so `$a = ['k' => 'x']; $a['l'] = [];` does not type
// `$a['l']` as string.
func (e *Env) widenElem(el types.Type, n *syntax.ArrayDimFetch) types.Type {
	v, ok := n.Var.(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return el
	}
	ws, back := e.reachingWrites(v)
	if k, ok := syntax.LiteralArrayKey(n.Dim); n.Dim != nil && ok {
		ws, back = writesForKey(ws, back, k)
	}
	return e.widenWrites(el, ws, back)
}

// writesForKey keeps the element writes that may store into literal key k:
// writes to k itself, to a computed key, appends when k is an integer, and
// the nested and unknown writes widenWrites judges itself. `$a[1] =
// explode(…)` does not change `$a[0]`.
func writesForKey(ws []*elemWrite, back []bool, k string) ([]*elemWrite, []bool) {
	var kw []*elemWrite
	var kb []bool
	for i, w := range ws {
		if !w.nested && w.known() {
			d := w.dim()
			if d == nil && !types.IsIntKey(k) {
				continue // appends add integer keys
			}
			if wk, ok := syntax.LiteralArrayKey(d); d != nil && ok && wk != k {
				continue
			}
		}
		kw = append(kw, w)
		kb = append(kb, back[i])
	}
	return kw, kb
}

// widenVarElem unions el with every value written into the elements of
// variable v by a write reaching v (see widenElem). Nested writes are
// skipped (see widenKey); an unknown write makes the result unknown.
func (e *Env) widenVarElem(el types.Type, v *syntax.Variable) types.Type {
	ws, back := e.reachingWrites(v)
	return e.widenWrites(el, ws, back)
}

// widenWrites is widenVarElem for the reaching writes ws (back: back edge).
func (e *Env) widenWrites(el types.Type, ws []*elemWrite, back []bool) types.Type {
	if len(ws) == 0 {
		return el
	}
	ts := []types.Type{el}
	for i, w := range ws {
		if w.nested {
			continue // changes an element already counted; see widenKey
		}
		if !w.known() {
			return types.Unknown
		}
		t := e.elementStoredType(w)
		if t.IsUnknown() && back[i] {
			continue // a back-edge cycle adds nothing (as for variables)
		}
		ts = append(ts, t)
	}
	return types.Union(ts...)
}

// writtenType is the value an element write stores.
func (e *Env) writtenType(a *syntax.Assign) types.Type {
	if a.Op.Kind == syntax.TEqual || a.Op.Kind == syntax.TCoalesceEqual {
		return e.TypeOf(a.Value)
	}
	return e.TypeOf(a)
}

// elementStoredType separates the value stored by a mutation from its
// expression result: a postfix increment returns the previous value.
func (e *Env) elementStoredType(w *elemWrite) types.Type {
	if w.inc != nil {
		if e.optionalAssertion(w.inc) {
			return types.Mixed
		}
		return e.incDecStored(w.inc)
	}
	if e.optionalAssertion(w.a) {
		return types.Mixed
	}
	return e.writtenType(w.a)
}

// variableType is the type of a variable read: its reaching definitions
// (variableBase) with the element writes reaching the read applied to its
// array members (withElemWrites), so the value is right wherever it flows.
func (e *Env) variableType(v *syntax.Variable) types.Type {
	if v.Name == "" || v.Name == "this" {
		return e.variableBase(v)
	}
	t, known := e.bases[v] // computed by baseType
	if !known {
		t = e.variableBase(v)
	}
	w, changed := e.withElemWrites(t, v)
	if changed && !known {
		e.setBase(v, t)
	}
	return w
}

func (e *Env) setBase(v *syntax.Variable, t types.Type) {
	if e.bases == nil {
		e.bases = map[*syntax.Variable]types.Type{}
	}
	e.bases[v] = t
}

// baseType is TypeOf, except that a variable read gives its type before
// the element writes reaching it: for the readers that apply those writes
// themselves, key by key (dimType, foreachElem, foreachKey, pointer
// functions), keeping shapes precise.
func (e *Env) baseType(x syntax.Expr) types.Type {
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return e.TypeOf(x)
	}
	if t, ok := e.bases[v]; ok {
		return t
	}
	if t, ok := e.cache[v]; ok {
		return t // typed already, and no write changed it
	}
	if e.busy[v] {
		return types.Unknown // recursion, as TypeOf
	}
	e.busy[v] = true
	t := e.variableBase(v)
	delete(e.busy, v)
	e.setBase(v, t)
	return t
}

// withElemWrites applies the element writes reaching variable read v to
// the array members of its type t: their element type gains the written
// values (unknown, as `array`, after a nested or unknown write) and shapes
// are dropped. `$a = ['x' => 1]; $a['y'] = 'a'; $b = $a;` types $b as
// (int|string)[], not as the sealed shape {x: int}. Non-array members
// (strings, ArrayAccess objects) are unchanged.
// changed is false when t is returned as is.
func (e *Env) withElemWrites(t types.Type, v *syntax.Variable) (types.Type, bool) {
	return e.withElemWritesAt(t, v.Name, v, syntax.EnclosingVariableScope(v))
}

func (e *Env) withElemWritesAt(t types.Type, name string, use, scope syntax.Node) (types.Type, bool) {
	if t.IsUnknown() {
		return t, false
	}
	var arr, other []string
	for _, a := range t.Atoms() {
		if a == "array" || strings.HasSuffix(a, "[]") {
			arr = append(arr, a)
		} else {
			other = append(other, a)
		}
	}
	// A write into null creates an array (`?array $n; $n['k'] = 1;`).
	nullOnly := len(other) == 1 && other[0] == "null"
	if len(arr) == 0 && !nullOnly {
		return t, false
	}
	ws, back := e.reachingWritesAt(name, use, scope)
	if len(ws) == 0 {
		return t, false
	}
	if nullOnly && e.writeDominatesAt(name, use, scope) {
		other = nil // every path to v writes into the array: no longer null
	}
	el := t.Elem()
	switch {
	case len(arr) == 0:
		el = types.Of("never") // null only: the writes fill a new array
	case t.IsSealedShape():
		ts := []types.Type{types.Of("never")}
		if !el.IsUnknown() {
			ts = append(ts, el)
		}
		for _, k := range t.ShapeKeys() {
			ts = append(ts, k.Type)
		}
		el = types.Union(ts...)
	case slices.Contains(arr, "array"):
		el = types.Unknown // elements of a plain array are unknown
	}
	for _, w := range ws {
		if w.nested {
			el = types.Unknown // changes an element in place
		}
	}
	if !el.IsUnknown() {
		el = e.widenWrites(el, ws, back)
	}
	if el.IsUnknown() || el.OnlyOf("never") {
		arr = []string{"array"}
	} else {
		arr = arr[:0]
		for _, a := range el.Atoms() {
			arr = append(arr, a+"[]")
		}
	}
	return types.Of(append(other, arr...)...).WithNonEmpty(t.IsNonEmptyArray()).WithArrayKey(e.arrayKeyAt(t, name, use, scope)), true
}

func (e *Env) variableBase(v *syntax.Variable) types.Type {
	if v.Name == "" {
		return types.Unknown
	}
	if v.Name == "this" {
		fqn := e.selfClass(syntax.EnclosingClass(v))
		if fqn == "" {
			return types.Unknown
		}
		t := types.Of(`\` + fqn)
		if c := e.Index.Class(fqn, e.PHP); c != nil && c.Kind == syntax.KindEnum {
			// `if ($this === self::Off) { return; }`: an enum method narrows
			// its case ($this is never reassigned).
			t = e.narrowExprAfter(t, v, "this", syntax.EnclosingVariableScope(v), 0)
		}
		return t
	}
	return e.variableBaseAt(v.Name, v, v)
}

// TypeOfVarAt types variable name as a read written at expression at would
// be typed (without the element writes reaching it): the definitions
// reaching at, narrowed by the conditions guarding at. It asks what holds of
// a variable at a place that does not read it, such as a call consuming a
// value computed from it earlier.
func (e *Env) TypeOfVarAt(name string, at syntax.Expr) types.Type {
	if name == "" || name == "this" {
		return types.Unknown
	}
	return e.variableBaseAt(name, at, nil)
}

// ReachesUnchanged reports whether the definitions of variable read v that
// reach it are exactly those reaching expression at, a later place of its
// scope: no assignment between them may have given the variable another
// value (see TypeOfVarAt).
func (e *Env) ReachesUnchanged(v *syntax.Variable, at syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(v)
	if v.Name == "" || v.NameExpr != nil || syntax.EnclosingVariableScope(at) != scope || at.Span().Start < v.Span().End {
		return false
	}
	defs := e.scopeVars(scope).defs[v.Name]
	if len(defs) > maxVarDefs {
		return false
	}
	positions := func(use syntax.Node) []uint32 {
		fwd, back, _ := e.reaching(defs, use, scope)
		var out []uint32
		for _, d := range append(fwd, back...) {
			out = append(out, d.pos)
		}
		slices.Sort(out)
		return out
	}
	return slices.Equal(positions(v), positions(at))
}

// variableBaseAt is variableBase for variable name read at use; read is the
// variable read itself, whose dynamic-read facts it records (nil for none).
func (e *Env) variableBaseAt(name string, use syntax.Expr, read *syntax.Variable) types.Type {
	scope := syntax.EnclosingVariableScope(use)
	sv := e.scopeVars(scope)
	defs := sv.defs[name]
	if len(defs) > maxVarDefs {
		return types.Unknown
	}
	fwd, back, from := e.reaching(defs, use, scope)
	clob := e.clobbered(sv, fwd, back, use, scope)
	undef := !clob && len(fwd)+len(back) > 0 && e.maybeUndefinedAt(sv, defs, fwd, name, use, scope)
	if read != nil {
		e.noteDynRead(read, clob, undef)
	}
	if clob {
		return types.Unknown
	}
	ts := make([]types.Type, 0, len(fwd)+len(back))
	// Values carried by a loop's back edge (and the step of a for loop
	// around v) never passed the conditions outside that loop: they are
	// narrowed apart, from the start of the innermost loop holding both
	// the definition and v.
	var backTs []types.Type
	after, backAfter := uint32(0), uint32(0)
	for _, d := range fwd {
		t := d.typ()
		if !t.IsUnknown() && e.forStep(d, use, scope) {
			// The step of a for loop around v runs before the condition
			// is tested again: a back edge, not a definition after it.
			if nt := e.loopCondNarrow(t, d, name, use, scope); !nt.IsUnknown() {
				backTs = append(backTs, nt)
				backAfter = max(backAfter, loopAround(d, use, scope))
			}
			continue
		}
		ts = append(ts, t)
		after = max(after, d.pos, d.end)
	}
	for _, d := range back {
		// An unknown back-edge type (often a cycle through this very use)
		// adds nothing: keep what the forward definitions say.
		if t := d.typ(); !t.IsUnknown() {
			backTs = append(backTs, e.loopCondNarrow(t, d, name, use, scope))
			backAfter = max(backAfter, loopAround(d, use, scope))
		}
	}
	if len(ts)+len(backTs) == 0 {
		return types.Unknown
	}
	if undef {
		ts = append(ts, types.Null) // read before any assignment on some path
	}
	settle := func(t types.Type) types.Type {
		if t.HasShape() || !t.ArrayKey().IsUnknown() || t.IsNonEmptyArray() {
			if (t.HasShape() || !t.ArrayKey().IsUnknown()) && e.shapeClobbered(scope, name) {
				t = t.WithoutShape().WithArrayKey(types.Unknown)
			}
			if t.IsNonEmptyArray() && e.nonEmptyBroken(scope, name, from, use) {
				t = t.WithNonEmpty(false)
			}
		}
		return t
	}
	if len(backTs) == 0 {
		return e.narrowExprAfter(settle(types.Union(ts...)), use, name, scope, after)
	}
	// A union with an unknown member is unknown: keep that rule.
	if all := settle(types.Union(append(ts, backTs...)...)); all.IsUnknown() {
		return all
	}
	parts := []types.Type{e.narrowExprAfter(settle(types.Union(backTs...)), use, name, scope, backAfter)}
	if len(ts) > 0 {
		parts = append(parts, e.narrowExprAfter(settle(types.Union(ts...)), use, name, scope, after))
	}
	return types.Union(parts...)
}

// loopAround returns the start of the innermost loop holding both
// definition d and use, where the conditions a value d carries around
// the loop passes again begin (d's own end, should no loop hold both).
func loopAround(d varDef, use, scope syntax.Node) uint32 {
	start, found := max(d.pos, d.end), false
	for p := use.Parent(); p != nil && p != scope && !found; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if sp := p.Span(); d.pos >= sp.Start && d.pos < sp.End {
				start, found = sp.Start, true
			}
		}
	}
	return start
}

// forStep reports whether definition d sits in the step expressions of a
// for loop whose body holds use.
func (e *Env) forStep(d varDef, use, scope syntax.Node) bool {
	child := use
	for p := use.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		if f, ok := p.(*syntax.For); ok && child == syntax.Node(f.Body) {
			for _, x := range f.Loop {
				if sp := x.Span(); d.pos >= sp.Start && d.pos < sp.End {
					return true
				}
			}
		}
	}
	return false
}

// loopCondNarrow narrows the type t of back-edge definition d of name read
// at use when d sits in the condition (or step) of a loop around use that is tested before
// the body runs again: `do { … } while ($e = $e->getPrevious());` only
// loops with a truthy $e; `for (…; $x !== null; $x = next($x))` likewise.
func (e *Env) loopCondNarrow(t types.Type, d varDef, name string, use, scope syntax.Node) types.Type {
	in := func(x syntax.Node) bool {
		sp := x.Span()
		return d.pos >= sp.Start && d.pos < sp.End
	}
	for p := use.Parent(); p != nil && p != scope; p = p.Parent() {
		switch n := p.(type) {
		case *syntax.DoWhile:
			if in(n.Cond) {
				return e.applyCond(t, n.Cond, name, true)
			}
		case *syntax.For:
			if len(n.Cond) == 0 {
				continue
			}
			for _, x := range append(slices.Clone(n.Loop), n.Cond...) {
				if in(x) {
					return e.applyCond(t, n.Cond[len(n.Cond)-1], name, true)
				}
			}
		}
	}
	return t
}

// iterElem is the element type of iterating over t: unknown as soon as one
// member (plain `array`, `iterable`, a Traversable class…) has unknown
// elements, so `string|iterable` does not iterate as `string`.
func iterElem(t types.Type) types.Type {
	for _, a := range t.Without("null", "false").Atoms() {
		if !strings.HasSuffix(a, "[]") {
			return types.Unknown
		}
	}
	return t.Elem()
}

// outermostLoop returns the outermost loop statement enclosing n within
// scope, or nil.
func outermostLoop(n, scope syntax.Node) syntax.Node {
	var loop syntax.Node
	for p := n.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			loop = p
		}
	}
	return loop
}

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
				// number), not only `=` (by reference: not a kill).
				if _, plain := n.Var.(*syntax.Variable); plain && !n.ByRef {
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
			case *syntax.Catch:
				if n.Var != nil {
					var atoms []string
					for _, t := range n.Types {
						atoms = append(atoms, `\`+e.Names.Class(t.Value, t.Span().Start))
					}
					add(n.Var.Name, n.Var.Span().Start, func() types.Type { return types.Of(atoms...) })
				}
			case *syntax.Global, *syntax.StaticStmt:
				// types.Unknown provenance.
				_, static := n.(*syntax.StaticStmt)
				syntax.Inspect(n, func(m syntax.Node) bool {
					if v, ok := m.(*syntax.Variable); ok {
						add(v.Name, v.Span().Start, func() types.Type { return types.Unknown })
						if static {
							sv.statics = mark(sv.statics, v.Name)
						} else {
							sv.shared = mark(sv.shared, v.Name)
						}
					}
					return true
				})
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
	inner := d
	for {
		parent, ok := syntax.UnwrapParens(inner.Var).(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		inner = parent
	}
	if v := asVariable(inner.Var); v != nil {
		w := &elemWrite{inc: n, nested: inner != d, key: inner.Dim}
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
	add := func(name string, w *elemWrite) {
		sv.elemWrites[name] = append(sv.elemWrites[name], varDef{pos: outer.Span().Start, end: outer.Span().End, w: w})
	}
	switch t := target.(type) {
	case *syntax.ArrayDimFetch:
		if v, ok := t.Var.(*syntax.Variable); ok && v.Name != "" {
			add(v.Name, &elemWrite{a: a})
			return
		}
		// Nested write: remember which first-level key it changes.
		inner := t
		for {
			d, ok := inner.Var.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			inner = d
		}
		if v, ok := inner.Var.(*syntax.Variable); ok && v.Name != "" {
			add(v.Name, &elemWrite{a: a, nested: true, key: inner.Dim})
		}
	case *syntax.List:
		for _, it := range t.Items {
			if it != nil && it.Value != nil {
				e.collectDimWrites(outer, nil, it.Value, sv)
			}
		}
	case *syntax.Array:
		for _, it := range t.Items {
			if it != nil && it.Value != nil {
				e.collectDimWrites(outer, nil, it.Value, sv)
			}
		}
	}
}

func (e *Env) paramType(scope syntax.Node, p *syntax.Param) types.Type {
	at := p.Span().Start
	declared := types.FromNode(p.Type, e.resolver(at))
	if p.Variadic && !declared.IsUnknown() {
		atoms := make([]string, 0, len(declared.Atoms()))
		for _, a := range declared.Atoms() {
			atoms = append(atoms, a+"[]")
		}
		declared = types.Of(atoms...)
	}
	if lit, ok := p.Default.(*syntax.ConstFetch); ok && strings.EqualFold(lit.Name.Value, "null") && !declared.IsUnknown() {
		declared = types.Union(declared, types.Null)
	}
	doc := ""
	if d := e.DocOf(scope); d != nil && !e.native {
		for _, dp := range d.EffectiveParams() {
			if dp.Name == p.Var.Name {
				doc = dp.Type
			}
		}
	}
	if doc != "" {
		dt := e.docType(doc, e.resolverFor(scope, at))
		if declared.IsUnknown() || declared.HasAny("array", "iterable", "mixed") {
			return dt
		}
		return declared.WithTypeArgsFrom(dt)
	}
	if m, ok := scope.(*syntax.Method); ok && !e.native && (declared.IsUnknown() || declared.HasAny("array", "iterable", "mixed")) {
		if dt := e.inheritedParamDoc(m, nodeIndex(m.Params, p)); !dt.IsUnknown() && (!declared.HasAny("array", "iterable") || dt.Without("null").IsArrayLike()) {
			return dt
		}
	}
	if e.native || p.Variadic {
		return declared
	}
	// A callback's untyped parameter receives what its caller passes; one
	// declared `array` (or `iterable`) keeps the passed array's facts.
	var ct types.Type
	switch fn := scope.(type) {
	case *syntax.Closure:
		if declared.IsUnknown() || declared.OnlyOf("array", "iterable") {
			ct = e.contextParamType(fn, nodeIndex(fn.Params, p))
		}
	case *syntax.ArrowFunction:
		if declared.IsUnknown() || declared.OnlyOf("array", "iterable") {
			ct = e.contextParamType(fn, nodeIndex(fn.Params, p))
		}
	}
	if declared.IsUnknown() || ct.IsArrayLike() {
		return ct
	}
	return declared
}

func (e *Env) inlineVarDocs(scope syntax.Node, addDoc func(string, uint32, uint32, func() types.Type)) {
	var span syntax.Span
	if scope == nil {
		span = syntax.Span{Start: 0, End: uint32(len(e.File.Src))}
	} else {
		span = scope.Span()
	}
	toks := e.File.Tokens
	first := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= span.Start })
	excluded := promotedHookDocSpans(scope)
	hook := 0
	for _, t := range toks[first:] {
		if t.Start >= span.End {
			break
		}
		for hook < len(excluded) && excluded[hook].End <= t.Start {
			hook++
		}
		if hook < len(excluded) && excluded[hook].Start <= t.Start {
			continue
		}
		if t.Kind != syntax.TDocComment && t.Kind != syntax.TComment {
			continue
		}
		text := string(e.File.Src[t.Start:t.End])
		if !strings.Contains(text, "@var") && !strings.Contains(text, "@phpstan-var") && !strings.Contains(text, "@psalm-var") {
			continue
		}
		d := phpdoc.Parse(text)
		for _, hint := range d.EffectiveVars() {
			typ, name := hint.Type, hint.Name
			if name == "" {
				continue
			}
			ty := e.docType(typ, e.resolverFor(e.nodeAt(scope, t.Start), t.Start))
			addDoc(name, t.Start, t.End, func() types.Type { return ty })
		}
	}
}

// semicolonBetween reports whether a `;` token lies in [from, to).
func (e *Env) semicolonBetween(from, to uint32) bool {
	toks := e.File.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= from })
	for ; i < len(toks) && toks[i].Start < to; i++ {
		if toks[i].Kind == syntax.TSemicolon {
			return true
		}
	}
	return false
}

// nodeAt returns scope itself (or nil for file scope); used to resolve doc
// names relative to the enclosing declarations.
func (e *Env) nodeAt(scope syntax.Node, _ uint32) syntax.Node { return scope }

// branchAssigned returns the variables assigned (plain `$x = …;` at the top
// level of the branch) in every branch of a complete if/elseif/else chain;
// branches that always leave (return/throw/…) do not constrain the result.
func branchAssigned(n *syntax.If) []string {
	if n.Else == nil {
		return nil
	}
	bodies := []syntax.Stmt{n.Body}
	for _, ei := range n.ElseIfs {
		bodies = append(bodies, ei.Body)
	}
	bodies = append(bodies, n.Else.Body)
	var common map[string]bool
	assignedSomewhere := false
	for _, b := range bodies {
		if terminates(b) {
			continue
		}
		set := topLevelAssigns(b)
		if common == nil {
			common = set
		} else {
			for k := range common {
				if !set[k] {
					delete(common, k)
				}
			}
		}
		assignedSomewhere = true
	}
	if !assignedSomewhere {
		return nil
	}
	out := make([]string, 0, len(common))
	for k := range common {
		out = append(out, k)
	}
	return out
}

func topLevelAssigns(s syntax.Stmt) map[string]bool {
	set := map[string]bool{}
	stmts := []syntax.Stmt{s}
	if b, ok := s.(*syntax.Block); ok {
		stmts = b.Stmts
	}
	for _, st := range stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			continue
		}
		if a, ok := es.Expr.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual {
			if v, ok := a.Var.(*syntax.Variable); ok && v.Name != "" {
				set[v.Name] = true
			}
		}
	}
	return set
}

// joinSpan is the region after an if-chain, up to the end of the enclosing
// statement list.
func joinSpan(n *syntax.If, f *syntax.File) syntax.Span {
	end := uint32(len(f.Src))
	if p := n.Parent(); p != nil {
		end = p.Span().End
	}
	return syntax.Span{Start: n.Span().End, End: end}
}
