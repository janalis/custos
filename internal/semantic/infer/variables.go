package infer

import (
	"slices"
	"strings"

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
	// runs caches runsCode once runsDone is set.
	runs     []mutation
	runsDone bool
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
	var nested []*elemWrite
	var nestedBack []bool
	for i, w := range ws {
		if w.nested {
			nested, nestedBack = append(nested, w), append(nestedBack, back[i])
			continue
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
	return e.withNestedWrites(types.Union(ts...), el.OnlyOf("never"), nested, nestedBack)
}

// staticType is the value a static local holds where its declaration
// binds it: its initializer (null without one) on the first call, else the
// value a former call left, one of the plain assignments of the function.
// Any other write (element, reference, increment, unset, out-argument,
// inline annotation, dynamic) makes it unknown.
func (e *Env) staticType(s *syntax.StaticVar) types.Type {
	name, scope := s.Var.Name, syntax.EnclosingVariableScope(s)
	sv := e.scopeVars(scope)
	if sv.dynamic || sv.refs[name] || sv.shared[name] || len(sv.elemWrites[name]) > 0 {
		return types.Unknown
	}
	ts := []types.Type{types.Null}
	if s.Default != nil {
		ts[0] = e.TypeOf(s.Default)
	}
	for _, d := range e.changing(sv.defs[name]) {
		switch {
		case d.barrier:
		case d.asg != nil:
			ts = append(ts, e.TypeOf(d.asg.Value))
		case d.pos >= s.Span().Start && d.pos < s.Span().End:
		default:
			return types.Unknown
		}
	}
	for _, t := range ts {
		if t.IsUnknown() {
			return types.Unknown
		}
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
	e.remember(undo{kind: undoBase, v: v})
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
		if types.IsArrayAtom(a) {
			arr = append(arr, a)
		} else {
			other = append(other, a)
		}
	}
	// A write into null creates an array (`?array $n; $n['k'] = 1;`),
	// also when the value may be a string instead (`?string $s; $s[0] =
	// 'x';` makes null an array, see below for the string). So does a
	// write into false (deprecated from PHP 8.1), which only the proven
	// Env follows: elsewhere false stays, as the rules reporting offsets
	// of false or an unhandled false expect.
	vivified := func(a string) bool { return a == "null" || e.proven && a == "false" }
	if len(arr) == 0 && !slices.ContainsFunc(other, vivified) {
		// A write into a string offset (`$s[0] = 'x'`) changes its value
		// and may lengthen it: keep the type, not the string's values (an
		// int member keeps its own).
		if slices.Contains(other, "string") && t.HasStringValues() {
			if ws, _ := e.reachingWritesAt(name, use, scope); len(ws) > 0 {
				return t.WithoutStringValues(), true
			}
		}
		return t, false
	}
	ws, back := e.reachingWritesAt(name, use, scope)
	if len(ws) == 0 {
		return t, false
	}
	if slices.ContainsFunc(other, vivified) && e.writeDominatesAt(name, use, scope) {
		other = slices.DeleteFunc(other, vivified) // every path to v writes into the array: no longer null (or false)
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
	return types.Of(append(other, arr...)...).WithElem(el).WithNonEmpty(t.IsNonEmptyArray()).WithArrayKey(e.arrayKeyAt(t, name, use, scope)), true
}

func (e *Env) variableBase(v *syntax.Variable) types.Type {
	if v.Name == "" {
		return types.Unknown
	}
	if v.Name == "this" {
		if e.proven && rebindable(v) {
			return types.Unknown
		}
		cl := syntax.EnclosingClass(v)
		c := e.indexCaches()
		if t, ok := c.this[cl]; ok {
			return t
		}
		fqn := e.selfClass(cl)
		if fqn == "" {
			return e.traitRequired(v)
		}
		t := types.Of(`\` + fqn)
		if k := e.Index.Class(fqn, e.PHP); k != nil && k.Kind == syntax.KindEnum {
			// `if ($this === self::Off) { return; }`: an enum method narrows
			// its case ($this is never reassigned).
			return e.narrowExprAfter(t, v, "this", syntax.EnclosingVariableScope(v), 0)
		}
		if c.this == nil {
			c.this = map[*syntax.ClassLike]types.Type{}
		}
		c.this[cl] = t
		return t
	}
	return e.variableBaseAt(v.Name, v, v)
}

// rebindable reports whether the `$this` read v sits in a closure or an
// arrow function, which Closure::bind, bindTo or call may run bound to
// another object (a static one cannot read `$this` at all).
func rebindable(v *syntax.Variable) bool {
	switch syntax.EnclosingFuncLike(v).(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
		return true
	}
	return false
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
	// A method argument binding a by-value parameter is no definition: kept,
	// it would hide the guards between the definitions before it and use.
	defs = e.changing(defs)
	fwd, back, from := e.reaching(defs, use, scope)
	clob := e.clobbered(sv, fwd, back, use, scope)
	undef := !clob && len(fwd)+len(back) > 0 && e.maybeUndefinedAt(sv, defs, fwd, name, use, scope)
	if read != nil {
		e.noteDynRead(read, clob, undef)
	}
	if clob || e.proven && e.unproven(sv, fwd, back, name, use, scope) {
		return types.Unknown
	}
	ts := make([]types.Type, 0, len(fwd)+len(back))
	// tsAfter holds, for each of ts, where its definition ends: the
	// conditions after it narrow its value (see narrowEach).
	tsAfter := make([]uint32, 0, len(fwd)+1)
	// Values carried by a loop's back edge (and the step of a for loop
	// around v) never passed the conditions outside that loop: they are
	// narrowed apart, from the start of the innermost loop holding both
	// the definition and v.
	var backTs []types.Type
	after, backAfter := uint32(0), uint32(0)
	// partial: a back-edge value was left out, so the values (not the
	// atoms) the other definitions give are not all the variable may hold.
	partial := redefinedInLoop(defs, use, scope)
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
		tsAfter = append(tsAfter, max(d.pos, d.end))
		after = max(after, d.pos, d.end)
	}
	for _, d := range back {
		// An unknown back-edge type (often a cycle through this very use)
		// adds nothing: keep what the forward definitions say.
		if t := d.typ(); !t.IsUnknown() {
			backTs = append(backTs, e.loopCondNarrow(t, d, name, use, scope))
			backAfter = max(backAfter, loopAround(d, use, scope))
		} else {
			partial = true
		}
	}
	if len(ts)+len(backTs) == 0 {
		return types.Unknown
	}
	// The values of the definitions hold only where no write they do not
	// show may have replaced them (see ProvenTypeOf): the proven Env makes
	// such a read unknown, the others keep its atoms.
	if !partial && !e.proven && carriesValues(ts, backTs) {
		partial = e.unproven(sv, fwd, back, name, use, scope)
	}
	if partial {
		for i := range ts {
			ts[i] = ts[i].WithoutValues()
		}
		for i := range backTs {
			backTs[i] = backTs[i].WithoutValues()
		}
	}
	if undef {
		ts = append(ts, types.Null) // read before any assignment on some path
		tsAfter = append(tsAfter, after)
	}
	settle := func(t types.Type) types.Type {
		if t.HasShape() || !t.ArrayKey().IsUnknown() || t.IsNonEmptyArray() {
			if (t.HasShape() || !t.ArrayKey().IsUnknown()) && e.shapeClobberedAt(scope, name, from, use) {
				t = t.WithoutShape().WithArrayKey(types.Unknown)
			}
			if t.IsNonEmptyArray() && e.nonEmptyBroken(scope, name, from, use) {
				t = t.WithNonEmpty(false)
			}
		}
		return t
	}
	if len(backTs) == 0 {
		return e.narrowEach(ts, tsAfter, settle, use, name, scope, after)
	}
	// A union with an unknown member is unknown: keep that rule.
	if all := settle(types.Union(append(ts, backTs...)...)); all.IsUnknown() {
		return all
	}
	parts := []types.Type{e.narrowExprAfter(settle(types.Union(backTs...)), use, name, scope, backAfter)}
	if len(ts) > 0 {
		parts = append(parts, e.narrowEach(ts, tsAfter, settle, use, name, scope, after))
	}
	return types.Union(parts...)
}

// carriesValues reports whether one of the types carries values.
func carriesValues(lists ...[]types.Type) bool {
	for _, l := range lists {
		for _, t := range l {
			if t.HasValues() {
				return true
			}
		}
	}
	return false
}

// maxNarrowEach bounds the definitions narrowEach narrows apart: beyond
// it they are narrowed together, from the last one (reads of variables
// with many conditional definitions would cost a walk each).
const maxNarrowEach = 4

// narrowEach narrows the types ts of the forward definitions reaching use
// (each ending at the matching tsAfter, the last at after) by the
// conditions guarding use, together from the last definition. When the
// result holds null, the definitions are also narrowed apart, each from
// where it ends: a condition between an earlier definition and a later,
// conditional one (`if ($x === null) { return; } if ($c) { $x = f($x); }`)
// still rules null out of the earlier value. Past maxNarrowEach
// definitions, or when one comes out unknown (an impossible path, or the
// analysis gave up), null stays.
func (e *Env) narrowEach(ts []types.Type, tsAfter []uint32, settle func(types.Type) types.Type, use syntax.Expr, name string, scope syntax.Node, after uint32) types.Type {
	all := e.narrowExprAfter(settle(types.Union(ts...)), use, name, scope, after)
	if !all.Has("null") || len(ts) > maxNarrowEach || !slices.ContainsFunc(tsAfter, func(a uint32) bool { return a != after }) {
		return all
	}
	for i, t := range ts {
		nt := e.narrowExprAfter(settle(t), use, name, scope, tsAfter[i])
		if nt.IsUnknown() || nt.Has("null") {
			return all
		}
	}
	// Each value rules null out: the rest of what the conditions left (no
	// type at all, which says nothing, when that was null alone).
	return all.Without("null")
}

// redefinedInLoop reports a use inside one of the definitions of its
// variable (`$i++`, `$x = $x + 1`) in a loop: the value that definition
// stored on the previous iteration reaches the use too, which reaching
// definitions do not list.
func redefinedInLoop(defs []varDef, use, scope syntax.Node) bool {
	pos := use.Span().Start
	for _, d := range defs {
		if d.pos > pos {
			break
		}
		if d.end > pos && !d.barrier {
			if loop := outermostLoop(use, scope); loop != nil && d.pos >= loop.Span().Start {
				return true
			}
		}
	}
	return false
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
			// A value the body stores comes back to the top of the body
			// only once the condition held (this loop being the innermost
			// holding both).
			if in(n.Cond) || in(n.Body) {
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
			if in(n.Body) {
				return t // the innermost loop holding both
			}
		case *syntax.While, *syntax.Foreach:
			if in(n) {
				return t // the innermost loop holding both
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
// scope, or nil. The array a foreach iterates and the initializers of a for
// loop are evaluated once, before the loop: they are not within it.
func outermostLoop(n, scope syntax.Node) syntax.Node {
	return outermostLoopAfter(n, scope, 0)
}

// outermostLoopAfter is outermostLoop among the loops starting after
// position after.
func outermostLoopAfter(n, scope syntax.Node, after uint32) syntax.Node {
	var loop syntax.Node
	child := n
	for p := n.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		switch p := p.(type) {
		case *syntax.Foreach:
			if syntax.Node(p.Expr) != child && p.Span().Start > after {
				loop = p
			}
		case *syntax.For:
			if !slices.ContainsFunc(p.Init, func(e syntax.Expr) bool { return syntax.Node(e) == child }) && p.Span().Start > after {
				loop = p
			}
		case *syntax.While, *syntax.DoWhile:
			if p.Span().Start > after {
				loop = p
			}
		}
	}
	return loop
}
