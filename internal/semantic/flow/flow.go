// Package flow provides bounded, file-local value and state analysis.
// Environments are confined to one analysis; snapshots contain no syntax trees.
package flow

import (
	"cmp"
	"sort"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/infer"
)

// Context distinguishes incompatible escaping contracts.
type Context uint8

const (
	HTML Context = 1 << iota
	SQL
	Shell
	Header
	URL
	Path
)

// Limits bound hostile scopes and recursive wrapper queries. MaxRetained
// bounds the bindings and states kept as block inputs of one scope (memory);
// MaxCopied the bindings copied, joined and compared while iterating to a
// fixed point (time). An ordinary long function carries a hundred locals
// through a few hundred blocks, far below both.
const (
	MaxBlocks    = 8192
	MaxTransfers = 100000
	MaxRetained  = 8 * MaxTransfers
	MaxCopied    = 64 * MaxTransfers
	MaxSources   = 64
	MaxDepth     = 8
)

// Source records an originating input, or a symbolic parameter in a summary.
type Source struct {
	Kind      string
	Path      string
	Span      syntax.Span
	Parameter int // -1 for external input
}

// Value is a conservative union of possible inputs. Safe is the intersection
// of proven escaping contexts. Expr and Identity require agreement on every path.
type Value struct {
	interval *intInterval
	Sources  []Source
	Expr     syntax.Expr
	// elements holds the literal-keyed elements of an array variable that
	// are tracked on their own. Only a binding carries them: the value of an
	// expression never does.
	elements *elementSet
	// origins lists the expressions that may have produced a binding the
	// paths disagree on (Expr is then nil), each with what its own paths
	// established. It means nothing while Expr is set.
	origins *originSet
	// tested and unless belong to a binding only, never to the value of an
	// expression. tested is the short-circuit operation whose left operand
	// last read the binding, which was not written since. unless is such an
	// operation whose right operand assigns the binding and tests it in
	// place: the binding is not false (unlessFalse) or not null (unlessNull)
	// except where the last evaluation of the operation skipped that
	// operand. Between the evaluation of the operation and the branch on its
	// outcome, skippedFalse and skippedNull record that the value the
	// skipped operand left in place was not false or not null.
	tested, unless *syntax.Binary
	Length         int64
	Identity       uint32
	Invalidated    uint32
	Safe           Context
	Complete       bool
	NonFalse       bool
	NonNull        bool
	LengthKnown    bool
	headerMask     uint8
	// upload marks the table of uploaded files or one of its entries as PHP
	// built it: only there are the server-chosen fields what they claim.
	upload uint8
	// scalars marks an array whose values are numbers, booleans or null
	// while its keys remain whatever the request sent.
	scalars bool
	// matched marks request data a switch found equal to one of its string
	// labels: the text of a label, not what the request sent.
	matched bool
	// unlessFalse, unlessNull, skippedFalse and skippedNull qualify unless.
	unlessFalse, unlessNull, skippedFalse, skippedNull bool
	// entry marks the binding of a static variable that still holds what it
	// held when its declaration was reached: nothing wrote it since.
	entry bool
	// plain marks a value that is not an object: converting it to text,
	// comparing it, iterating it or reading an offset of it runs no method.
	// deep marks one that, when it is an array, holds no object either, at
	// any depth: what is done with its elements runs none.
	plain, deep bool
}

// Origin is one expression that may have produced the value read, with the
// failure values the guards on its paths exclude.
type Origin struct {
	Expr     syntax.Expr
	NonFalse bool
	NonNull  bool
}

// MaxOrigins bounds the producers followed for one binding.
const MaxOrigins = 4

// originSet is never modified once a binding holds it. A path that gave the
// binding a value no expression accounts for adds nothing to it: a set of one
// producer is a binding some other path left unknown. full records that
// there were more producers than are followed, which no later join undoes.
type originSet struct {
	items []Origin
	full  bool
}

// origins joins the producers of a binding two paths disagree on. It returns
// the set one of them already holds when the other adds nothing to it, and a
// full set past MaxOrigins.
func origins(a, b Value) *originSet {
	if a.Expr == b.Expr && (a.Expr != nil || a.origins == nil && b.origins == nil) {
		return nil
	}
	if held := a.origins; a.Expr == nil && held != nil && (held.full || held.holds(&b)) {
		return held
	}
	if held := b.origins; b.Expr == nil && held != nil && (held.full || held.holds(&a)) {
		return held
	}
	n := &originSet{}
	for _, v := range [2]*Value{&a, &b} {
		switch {
		case v.Expr != nil:
			n.add(Origin{v.Expr, v.NonFalse, v.NonNull})
		case v.origins != nil:
			for _, o := range v.origins.items {
				n.add(o)
			}
		}
	}
	if len(n.items) > MaxOrigins {
		return &originSet{full: true}
	}
	return n
}

// add joins one producer into a set under construction.
func (s *originSet) add(o Origin) {
	for i, known := range s.items {
		if known.Expr == o.Expr {
			s.items[i] = Origin{o.Expr, known.NonFalse && o.NonFalse, known.NonNull && o.NonNull}
			return
		}
	}
	s.items = append(s.items, o)
}

// holds reports that joining v adds nothing to the set: each producer of v
// is there with no more guards than v has.
func (s *originSet) holds(v *Value) bool {
	has := func(o Origin) bool {
		for _, known := range s.items {
			if known.Expr == o.Expr {
				return (!known.NonFalse || o.NonFalse) && (!known.NonNull || o.NonNull)
			}
		}
		return false
	}
	switch {
	case v.Expr != nil:
		return has(Origin{v.Expr, v.NonFalse, v.NonNull})
	case v.origins == nil:
		return true
	case v.origins.full:
		return false
	}
	for _, o := range v.origins.items {
		if !has(o) {
			return false
		}
	}
	return true
}

// excluding returns the set with a failure value excluded for every producer.
func (s *originSet) excluding(nonFalse, nonNull bool) *originSet {
	if s == nil {
		return nil
	}
	n := &originSet{items: make([]Origin, len(s.items)), full: s.full}
	for i, o := range s.items {
		n.items[i] = Origin{o.Expr, o.NonFalse || nonFalse, o.NonNull || nonNull}
	}
	return n
}

// sameOrigins compares two sets by content.
func sameOrigins(a, b *originSet) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil || a.full != b.full || len(a.items) != len(b.items) {
		return false
	}
	for i, o := range a.items {
		if o != b.items[i] {
			return false
		}
	}
	return true
}

const (
	uploadEntry uint8 = iota + 1
	uploadTable
)

// elementSet is never modified once a binding holds it: a write makes a new
// one, which lets every copy of the array value share it.
type elementSet struct{ items map[string]Value }

// without returns the set less one element, which a callee holding a
// reference to it may have rewritten; nothing when the element is not known.
func (s *elementSet) without(key string, known bool) *elementSet {
	if s == nil || !known {
		return nil
	}
	n := &elementSet{items: map[string]Value{}}
	for k, v := range s.items {
		if k != key {
			n.items[k] = v
		}
	}
	if len(n.items) == 0 {
		return nil
	}
	return n
}

// with returns the set with one more element written.
func (s *elementSet) with(key string, v Value) *elementSet {
	n := &elementSet{items: map[string]Value{key: v}}
	if s != nil {
		for k, old := range s.items {
			if k != key {
				n.items[k] = old
			}
		}
	}
	return n
}

// common keeps the elements tracked in both sets, each holding what either
// may. It returns a itself when b adds nothing to it.
func common(a, b *elementSet) *elementSet {
	if a == b {
		return a
	}
	if a == nil || b == nil {
		return nil
	}
	kept := true
	for k, v := range a.items {
		w, ok := b.items[k]
		if !ok || !equal(merge(v, w), v) {
			kept = false
			break
		}
	}
	if kept {
		return a
	}
	n := &elementSet{items: map[string]Value{}}
	for k, v := range a.items {
		if w, ok := b.items[k]; ok {
			n.items[k] = merge(v, w)
		}
	}
	if len(n.items) == 0 {
		return nil
	}
	return n
}

// State is the most recent operation common to every reaching path.
type State struct {
	Operation  string
	Span       syntax.Span
	Successful bool
}

// Call describes a reached call and its argument values (in written order).
type Call struct {
	Node      syntax.Node
	Name      string
	Receiver  syntax.Expr
	Arguments []Value
}

type frame struct {
	vars   map[string]Value
	states map[uint32]State
	// aliased names are bound to storage that code outside the tracked
	// bindings can write: aliasHeld for storage only a callee can reach,
	// aliasAny for a reference kept where the analysis does not follow it (a
	// property, an array slot, a loop variable), and a larger number for the
	// names one `$a = &$b` binds together, which a write through either
	// replaces for both. nil for the common scope without such bindings.
	aliased map[string]uint32
	// definite tells that a name without a binding is a variable nothing
	// assigned: it is null, not false. It holds in a function that captures
	// nothing, until a write the analysis cannot name (`$$n = …`, `extract`,
	// an included file) may have created any variable.
	definite bool
	// wild tells that some name is bound to a variable the analysis cannot
	// name (`$r = &$$n`): a write through a reference kept somewhere may
	// reach any variable.
	wild bool
	// related lists what holds between the left operand of a short-circuit
	// operation and the variable its right operand assigns in place (see
	// relation). The list is never modified once a frame holds it.
	related []relation
}

// relation tells, for a short-circuit operation whose right operand assigns
// a variable and tests it in place, that wherever the left operand, read
// from the values now in place, would skip that operand, the variable is not
// false (relatedFalse) or not null (relatedNull). It holds across the turns
// of a loop that keep the left operand deciding with the assignment: the
// value a turn leaves in place where the next one skips passed the test.
type relation struct {
	by   *syntax.Binary
	bits uint8
}

const (
	relatedFalse uint8 = 1 << iota
	relatedNull
)

// relates returns what is recorded for an operation.
func (f *frame) relates(by *syntax.Binary) uint8 {
	for _, known := range f.related {
		if known.by == by {
			return known.bits
		}
	}
	return 0
}

// relate records what holds for an operation, in place of what did.
func (f *frame) relate(by *syntax.Binary, bits uint8) {
	kept := make([]relation, 0, len(f.related)+1)
	for _, known := range f.related {
		if known.by != by {
			kept = append(kept, known)
		}
	}
	if bits != 0 {
		kept = append(kept, relation{by, bits})
	}
	f.related = kept
}

// settled is what holds for an operation whatever its left operand decides:
// the variable is not false, or not null, on its own. A variable nothing
// assigned is null.
func (f *frame) settled(by *syntax.Binary) uint8 {
	v, tracked := f.vars[assignedInPlace(by.Right).Var.(*syntax.Variable).Name]
	return excluded(v, !tracked && f.definite)
}

// excluded is what a value settles: it is not false, or not null. null
// reports the value of a variable nothing assigned.
func excluded(v Value, null bool) uint8 {
	bits := uint8(0)
	if v.NonFalse || null {
		bits |= relatedFalse
	}
	if v.NonNull {
		bits |= relatedNull
	}
	return bits
}

// rewritten applies a write to what is recorded: the variable assigned in
// place keeps what the value written has, and a variable the left operand
// reads may now decide otherwise, which leaves what holds on its own.
func (f *frame) rewritten(r *scopeResult, name string, v Value) {
	if _, shared := f.aliased[name]; shared {
		// The names bound with it are written too.
		f.related = nil
		return
	}
	r.operations += len(f.related)
	var kept []relation
	for i, known := range f.related {
		bits := known.bits
		if c, _ := compares(known.by.Left); assignedInPlace(known.by.Right).Var.(*syntax.Variable).Name == name {
			bits &= excluded(v, false)
		} else if plainVariable(c.left) == name || plainVariable(c.right) == name {
			bits &= f.settled(known.by)
		}
		if bits == known.bits && kept == nil {
			continue
		}
		if kept == nil {
			kept = append(make([]relation, 0, len(f.related)), f.related[:i]...)
		}
		if bits != 0 {
			kept = append(kept, relation{known.by, bits})
		}
	}
	if kept != nil {
		f.related = kept
	}
}

// correlated joins what is recorded on two paths: what holds on both, by
// record or on its own. It reports whether the record of the frame changed.
func (f *frame) correlated(b frame) bool {
	var joined []relation
	for _, known := range f.related {
		if bits := (known.bits | f.settled(known.by)) & (b.relates(known.by) | b.settled(known.by)); bits != 0 {
			joined = append(joined, relation{known.by, bits})
		}
	}
	for _, known := range b.related {
		if bits := f.settled(known.by) & (known.bits | b.settled(known.by)); bits != 0 && f.relates(known.by) == 0 {
			joined = append(joined, relation{known.by, bits})
		}
	}
	changed := len(joined) != len(f.related)
	for i := 0; !changed && i < len(joined); i++ {
		changed = joined[i] != f.related[i]
	}
	f.related = joined
	return changed
}

const (
	aliasHeld uint32 = iota
	aliasAny
	aliasGroup // first identifier of a definite group
)

// imported is the reserved binding, under a name no variable has, for
// request data an import (`extract`, one-argument `parse_str`) may have put
// in any variable, tracked or not. It joins and clones like any binding.
const imported = ""

// importing joins request data an import may have spread over the scope.
func (f *frame) importing(v Value) {
	if tainted(v) {
		f.vars[imported] = unsettled(f.vars[imported], v)
	}
}

// reserved reports a binding that is not a variable: a record kept under a
// name no variable has (see requestElement, received, imported).
func reserved(name string) bool {
	return name == imported || strings.ContainsAny(name, "\x00&")
}

// requestElement returns the reserved binding name, one no variable has, of
// an element of a request table read under one constant key: the binding
// records that a switch just found the element equal to a fixed text. The
// uploads table is left out, its entries are arrays PHP built.
func requestElement(x syntax.Expr) string {
	fetch, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
	if !ok {
		return ""
	}
	table, ok := fetch.Var.(*syntax.Variable)
	key, constant := arrayKey(fetch.Dim)
	if !ok || !constant || table.NameExpr != nil || !external(table.Name) || table.Name == "_FILES" {
		return ""
	}
	// An escape sequence may not spell the key it seems to.
	if literal, written := syntax.UnwrapParens(fetch.Dim).(*syntax.Literal); written && strings.Contains(literal.Raw, `\`) {
		return ""
	}
	return table.Name + "\x00" + key
}

// match records that the request table element of that reserved name holds
// a fixed text on this path. Any code can rewrite the table: the record
// does not outlive an expression that may run some (see implicit).
func (f *frame) match(r *scopeResult, element string) {
	f.vars[element] = Value{Complete: true, NonFalse: true}
	f.alias(element, aliasHeld)
	r.matching = true
}

// unmatchAll discards every recorded match: what is called, a PHP function
// included, may run code that rewrites the request tables (a callback, a
// conversion method of an argument, an error handler), and so may a
// destructor when a value is replaced.
func (f *frame) unmatchAll(r *scopeResult) {
	if !r.matching {
		return
	}
	r.operations += len(f.aliased)
	for name := range f.aliased {
		if strings.Contains(name, "\x00") {
			delete(f.aliased, name)
			delete(f.vars, name)
		}
	}
}

// matched reports that the element still holds the text it was matched with
// on every path: no join with a path without the match, no call and no
// write to the table came between, and no reference to the table or one of
// its elements exists.
func (f *frame) matched(element string) bool {
	table, _, _ := strings.Cut(element, "\x00")
	_, shared := f.aliased[table]
	return element != "" && !shared && f.vars[element].NonFalse
}

// unmatch discards the matches recorded for the elements of a request table
// that is written.
func (f *frame) unmatch(r *scopeResult, table string) {
	if !external(table) {
		return
	}
	r.operations += len(f.aliased)
	for name := range f.aliased {
		if strings.HasPrefix(name, table+"\x00") {
			delete(f.aliased, name)
			delete(f.vars, name)
		}
	}
}

// Env shares one bounded analysis per variable scope among all inspections.
type Env struct {
	file     *syntax.File
	types    *infer.Env
	snapshot *Snapshot
	// consulted records the contracts read while summaries are extracted.
	consulted   map[string]bool
	scopes      map[syntax.Node]*scopeResult
	localBodies map[string]syntax.Node
	// global caches the file's function-bound global names once scanned, and
	// globalTaint the request data those functions write to them.
	global       map[string]bool
	globalTaint  map[string]Value
	writers      map[globalWriter]bool
	globalsKnown bool
	// overridden caches, per receiver class and method, whether a class of
	// the project replaces the method (see replaced).
	overridden map[[2]string]bool
	// extendedBy caches whether a class of the project extends a class.
	extendedBy map[string]bool
	// properties caches, per class and property name, whether reading the
	// property of the object itself runs no code (see backed).
	properties map[[2]string]bool
	// kindOf caches what each declared type read admits (see kinds).
	kindOf map[string]uint8
	// statics caches, per scope, what its code does to each of its static
	// variables (see declared).
	statics map[syntax.Node]map[string]static
}

type scopeResult struct {
	generator bool
	values    map[syntax.Expr]Value
	before    map[syntax.Node]map[uint32]State
	calls     []Call
	returns   []Value
	callIndex map[syntax.Node]int
	complete  bool
	// stale lists the assignments tested in place in a short-circuit
	// operand whose earlier result can reach the path that skips them.
	stale map[*syntax.Assign]bool
	// matching marks a scope in which a switch matched a request table
	// element with a fixed text (see frame.match).
	matching bool
	// leaves marks a scope no path of which completes normally.
	leaves bool
	// spent marks a scope that ran out of one of its memory budgets:
	// nothing more is worth computing for it.
	spent      bool
	operations int
	facts      int
	retained   int
	copies     int
	// global marks a file-level scope whose variables a function of the file
	// can replace through the globals table.
	global bool
	// holders is the request data closures of the scope write to the names
	// they capture by reference.
	holders map[string]Value
	// exit is the request data each variable may hold when the scope ends,
	// and globalWrites what the scope stores in literal globals table entries.
	exit         map[string]Value
	globalWrites map[string]Value
	// globalNames are the names the scope binds with `global`.
	globalNames map[string]bool
	// statics are the names the scope binds with `static`, each with what a
	// call that enters the function again can do to it (see staticKept).
	statics map[string]static
	// broken names the static variables the analysis assumed never to be
	// false where the function can be entered again, and found that they may
	// be. The scope is then analysed again without the assumption for the
	// names in distrust, or for every name (distrustAll).
	broken, distrust map[string]bool
	distrustAll      bool
	// runs marks a scope that may run code of the project other than its
	// own: a call that is not known to make none, a constructor, a callback.
	runs bool
	// modern marks a scope analysed for PHP 8 or later, where writing a
	// property of a value that is no object is an error.
	modern bool
	// receivers tells, for each method call on a variable, what the variable
	// holds wherever the call is reached (see follows).
	receivers map[syntax.Node]uint8
	// labelled is where the first `goto` label of the scope is written, plus
	// one: a statement after it can be reached without running the ones
	// written before it.
	labelled uint32
	// landings is read when a rule first asks about a label of the scope.
	landings *landings
}

// static describes a static variable of a scope: what a call that enters the
// function again can do to it, and what the code of the scope gives it (see
// declared).
type static struct {
	state               uint8
	starts, plain, deep bool
}

const (
	// staticKept: the function never assigns the variable; another
	// invocation leaves it as it is.
	staticKept uint8 = iota
	// staticHeld: the function assigns it, and it is assumed not to be
	// false wherever the function, having written it, calls out, reaches
	// its declaration again or ends, which is checked.
	staticHeld
	// staticLoose: the function assigns it and nothing is assumed.
	staticLoose
)

// New creates an environment without computing scope flow eagerly.
func New(file *syntax.File, types *infer.Env, snapshot *Snapshot) *Env {
	return &Env{file: file, types: types, snapshot: snapshot, scopes: map[syntax.Node]*scopeResult{}}
}

// Value returns the joined value of an expression at its execution point.
func (e *Env) Value(x syntax.Expr) Value {
	if x == nil {
		return Value{Complete: true}
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v, ok := r.values[x]
	if !ok {
		return Value{}
	}
	v.Complete = v.Complete && r.complete
	return v
}

// Resolve returns a single reaching expression only when every path agrees.
func (e *Env) Resolve(x syntax.Expr) (syntax.Expr, bool) {
	if x == nil {
		return nil, false
	}
	v := e.Value(x)
	return v.Expr, e.scope(syntax.EnclosingVariableScope(x)).complete && v.Expr != nil
}

// Excludes proves that a checked variable cannot equal false or null here.
// The exclusions are facts of the paths that reach the read: they hold of a
// value whose inputs are not all known, in a scope followed to its end.
func (e *Env) Excludes(x syntax.Expr, literal string) bool {
	if x == nil {
		return false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v := r.values[x]
	if !r.complete {
		return false
	}
	switch literal {
	case "false":
		return v.NonFalse
	case "null":
		return v.NonNull
	}
	return false
}

// Labelled reports that a jump to a `goto` label can land between what a
// node reads and the node: what is written before the node need not have run
// when it does, and only the facts of the paths that reach it say anything
// about it. A label counts when it is written before the node and after the
// first place its scope names one of the variables the node reads or a
// variable that can hold the same value as one of them (see landings). A
// label written before all of that is passed before anything is known of
// them, and one written after the node is only reached past it. Every label
// before the node counts at file level, where any call may write the
// variables, and for a node that reads no variable or reads `$this`.
//
// The node is the expression whose check is in question, not the statement
// around it: the variables of a loop body or of an element key say nothing
// about where the checked value was tested.
func (e *Env) Labelled(at syntax.Node) bool {
	return e.LabelledBefore(at, at)
}

// LabelledBefore is Labelled for a value checked in one place and used in
// another: the labels count up to use, the variables are the ones of checked.
func (e *Env) LabelledBefore(checked, use syntax.Node) bool {
	scope := syntax.EnclosingVariableScope(use)
	r := e.scope(scope)
	if r.labelled == 0 || r.labelled > use.Span().Start {
		return false
	}
	if scope == nil {
		return true
	}
	if r.landings == nil {
		r.landings = e.landings(scope)
	}
	since, named := use.Span().Start, false
	syntax.Inspect(checked, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok {
			// A variable of a function declared in the node is not known.
			from := r.landings.since[v.Name]
			if v.Name == "this" {
				from = 0
			}
			since, named = min(since, from), true
		}
		return true
	})
	if !named {
		since = 0
	}
	labels := r.landings.labels
	next := sort.Search(len(labels), func(i int) bool { return labels[i] >= since })
	return next < len(labels) && labels[next] < use.Span().Start
}

// landings tells where a jump can land in a scope and since where each of
// its variables is spoken of: the first place the scope names the variable
// or one that can hold the same value. Two variables can when a plain
// assignment, an assignment by reference or a `??=` writes one from the
// other, when a `foreach` takes one out of the other, or through a chain of
// those: a check of one is then a check of the other, wherever it is
// written. Nothing else ties two variables. A call hands on no value the
// rules trace, so neither its receiver nor its arguments count in the value
// written; the result of an operator other than `??`, or of a string that
// embeds a variable, is a new value; an element key only selects; and a
// statement that only reads (`echo`, `return`, a condition) passes nothing.
// A parameter is named from the start. Code that writes variables it does
// not name (a name chosen at run time, `extract`, `parse_str` filling the
// scope, `compact`, `get_defined_vars`, `$GLOBALS`, code evaluated or
// included) names them all where it is written.
type landings struct {
	labels []uint32
	since  map[string]uint32
}

func (e *Env) landings(scope syntax.Node) *landings {
	l := &landings{since: map[string]uint32{}}
	// Variables that can hold the same value belong to one group, named
	// after one of them.
	group := map[string]string{}
	var find func(name string) string
	find = func(name string) string {
		if parent := group[name]; parent != name {
			group[name] = find(parent)
		}
		return group[name]
	}
	first := map[string]uint32{}
	wild := ^uint32(0)
	mention := func(v *syntax.Variable) {
		if _, seen := group[v.Name]; !seen {
			group[v.Name], first[v.Name] = v.Name, v.Span().Start
		}
	}
	// tie puts the variables a value can pass between in one group.
	tie := func(sides ...syntax.Node) {
		held := ""
		for _, side := range sides {
			passing(side, func(v *syntax.Variable) {
				mention(v)
				if held == "" {
					held = v.Name
					return
				}
				a, b := find(held), find(v.Name)
				group[b] = a
				first[a] = min(first[a], first[b])
			})
		}
	}
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.Var != nil {
			group[p.Var.Name], first[p.Var.Name] = p.Var.Name, 0
		}
	}
	syntax.Inspect(syntax.VariableScopeBody(scope), func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.Closure:
			// The variables a closure takes are named where it is made.
			for _, use := range n.Uses {
				mention(use.Var)
			}
			return false
		case *syntax.Label:
			l.labels = append(l.labels, n.Span().Start)
		case *syntax.Eval, *syntax.Include:
			wild = min(wild, n.Span().Start)
		case *syntax.FuncCall:
			switch name, builtin := e.callName(n); {
			case !builtin:
			case name == "extract", name == "compact", name == "get_defined_vars", name == "parse_str" && n.Args != nil && len(n.Args.Args) == 1:
				wild = min(wild, n.Span().Start)
			}
		case *syntax.Assign:
			if n.Op.Kind == syntax.TEqual || n.Op.Kind == syntax.TCoalesceEqual {
				tie(n.Var, n.Value)
			}
		case *syntax.Foreach:
			tie(n.Expr, n.Key, n.Value)
		case *syntax.Variable:
			if n.NameExpr != nil || n.Name == "GLOBALS" {
				wild = min(wild, n.Span().Start)
			} else {
				mention(n)
			}
		}
		return true
	})
	for name := range group {
		l.since[name] = min(first[find(name)], wild)
	}
	return l
}

// passing visits the variables of an expression whose value the expression
// can hand on as it is: not the receiver or the arguments of a call, not the
// operands of an operator other than `??`, not what a string embeds, not an
// element key. A closure hands on the variables it takes.
func passing(side syntax.Node, visit func(*syntax.Variable)) {
	if side == nil {
		return
	}
	syntax.Inspect(side, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Variable:
			if n.NameExpr == nil {
				visit(n)
			}
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.InterpolatedString, *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.Binary:
			return n.Op.Kind == syntax.TCoalesce
		case *syntax.ArrayDimFetch:
			passing(n.Var, visit)
			return false
		case *syntax.Closure:
			for _, use := range n.Uses {
				visit(use.Var)
			}
			return false
		}
		return true
	})
}

// Reaching returns the expressions that may have produced the value of a
// plain variable read, in the order the paths were joined, each with the
// failure values its paths exclude. The second result reports that the paths
// agree on one producer. Otherwise the list may leave out paths on which the
// variable is unassigned or holds a value the scope does not account for,
// and it is empty when nothing is known or more than MaxOrigins reach.
func (e *Env) Reaching(x syntax.Expr) ([]Origin, bool) {
	if x == nil {
		return nil, false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v, ok := r.values[x]
	switch {
	case !ok || !r.complete:
		return nil, false
	case v.Expr != nil:
		return []Origin{{v.Expr, v.NonFalse, v.NonNull}}, true
	case v.origins != nil:
		return v.origins.items, false
	}
	return nil, false
}

// ExactLength returns a byte length established by a strict strlen guard.
func (e *Env) ExactLength(x syntax.Expr) (int64, bool) {
	v := e.Value(x)
	return v.Length, v.Complete && v.LengthKnown
}

// Tainted reports a retained external source without a context-specific
// escape. Request data is a may-property: an operand of unknown value adds
// none and erases none, so the value need not be completely known. A scope
// the analysis could not follow to its end proves nothing either way.
func (e *Env) Tainted(x syntax.Expr, context Context) bool {
	if x == nil {
		return false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v := r.values[x]
	return r.complete && v.Safe&context == 0 && tainted(v)
}

// Matched reports a value that is the text of a `switch` label the request
// data it was read from was found equal to: text the program wrote.
func (e *Env) Matched(x syntax.Expr) bool {
	if x == nil {
		return false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	return r.complete && r.values[x].matched
}

// Calls returns reached calls in a scope, excluding nested scopes. Returned
// slices are read-only. nil denotes the file-level scope.
func (e *Env) Calls(scope syntax.Node) []Call { return e.scope(scope).calls }

// StateBefore returns the last common operation on the receiver's proven
// allocation identity. Operations are normalized builtin/function names or
// lower-case method names. Conditional disagreement and escaping invalidate it.
// When operations is nonempty, the common operation must be listed there.
func (e *Env) StateBefore(at syntax.Node, receiver syntax.Expr, operations ...string) (State, bool) {
	r := e.scope(syntax.EnclosingVariableScope(at))
	v := e.Value(receiver)
	if !r.complete || !v.Complete || v.Identity == 0 {
		return State{}, false
	}
	s, ok := r.before[at][v.Identity]
	if !ok {
		return State{}, false
	}
	if len(operations) == 0 {
		return s, true
	}
	for _, op := range operations {
		if strings.EqualFold(op, s.Operation) {
			return s, true
		}
	}
	return State{}, false
}

const (
	sessionID   uint32 = ^uint32(0)
	outputID           = sessionID - 1
	committedID        = sessionID - 2
	responseID         = sessionID - 3
	cookieID           = sessionID - 4
	bufferID           = sessionID - 5
)

// GlobalStateBefore returns the common PHP request state in domain session,
// output, committed-output, response, cookie-header or output-buffer.
func (e *Env) GlobalStateBefore(at syntax.Node, domain string) (State, bool) {
	id := uint32(0)
	switch domain {
	case "session":
		id = sessionID
	case "output":
		id = outputID
	case "committed-output":
		id = committedID
	case "response":
		id = responseID
	case "cookie-header":
		id = cookieID
	case "output-buffer":
		id = bufferID
	default:
		return State{}, false
	}
	r := e.scope(syntax.EnclosingVariableScope(at))
	if !r.complete {
		return State{}, false
	}
	state, ok := r.before[at][id]
	return state, ok
}

func merge(a, b Value) Value {
	// A value without request data has nothing to escape: it is neutral for
	// the escaping proof of the other operand, whether or not it is known.
	sa, sb := a.Safe, b.Safe
	ua, ub := a.upload, b.upload
	ca, cb := a.scalars, b.scalars
	switch {
	case len(a.Sources) == 0 && len(b.Sources) == 0:
		sa, sb, ua, ub, ca, cb = HTML|SQL|Shell|Header|URL|Path, HTML|SQL|Shell|Header|URL|Path, 0, 0, false, false
	case len(a.Sources) == 0:
		sa, ua, ca = HTML|SQL|Shell|Header|URL|Path, ub, cb
	case len(b.Sources) == 0:
		sb, ub, cb = HTML|SQL|Shell|Header|URL|Path, ua, ca
	}
	v := Value{Complete: a.Complete && b.Complete, Safe: sa & sb, headerMask: a.headerMask & b.headerMask, NonFalse: a.NonFalse && b.NonFalse, NonNull: a.NonNull && b.NonNull, LengthKnown: a.LengthKnown && b.LengthKnown && a.Length == b.Length, Length: a.Length}
	v.interval = joinInts(a, b)
	v.elements = common(a.elements, b.elements)
	// Text of the program joined with a matched label is text of the program.
	v.matched = (a.matched || b.matched) && (a.matched || a.written()) && (b.matched || b.written())
	v.entry = a.entry && b.entry
	v.plain, v.deep = a.plain && b.plain, a.deep && b.deep
	v.scalars = ca && cb
	if ua == ub {
		v.upload = ua
	}
	if a.Expr == b.Expr {
		v.Expr = a.Expr
	}
	if a.Invalidated == b.Invalidated {
		v.Invalidated = a.Invalidated
	} else {
		v.Invalidated = ^uint32(0)
	}
	if a.Identity == b.Identity {
		v.Identity = a.Identity
	}
	var whole bool
	if v.Sources, whole = union(a.Sources, b.Sources); !whole {
		v.Complete = false
	}
	return v
}

// written reports the value of a literal: text the program wrote.
func (v *Value) written() bool {
	switch v.Expr.(type) {
	case *syntax.Literal, *syntax.StringPart:
		return v.Complete && len(v.Sources) == 0
	}
	return false
}

// union joins two source lists, each in the order of before, without
// repeating a source. It returns a itself when b adds nothing, so that a
// value joined again with what it already holds stays recognisably the
// same; the result never shares spare capacity a later append could write
// to. Sources past MaxSources are dropped, which the second result reports.
func union(a, b []Source) ([]Source, bool) {
	switch {
	case len(b) == 0 || len(a) == len(b) && &a[0] == &b[0]:
		return a[:len(a):len(a)], true
	case len(a) == 0:
		return b[:len(b):len(b)], true
	}
	i, j := 0, 0
	for ; j < len(b); j++ {
		for i < len(a) && before(a[i], b[j]) {
			i++
		}
		if i == len(a) || a[i] != b[j] {
			break
		}
	}
	if j == len(b) {
		return a[:len(a):len(a)], true
	}
	out := make([]Source, 0, min(len(a)+len(b), MaxSources))
	i, j = 0, 0
	for i < len(a) || j < len(b) {
		var next Source
		switch {
		case j == len(b) || i < len(a) && before(a[i], b[j]):
			next = a[i]
			i++
		case i == len(a) || before(b[j], a[i]):
			next = b[j]
			j++
		default:
			next = a[i]
			i++
			j++
		}
		if len(out) == MaxSources {
			return out, false
		}
		out = append(out, next)
	}
	return out, true
}

// before orders sources: by position, then by what distinguishes two sources
// at one position (a parameter, another file).
func before(a, b Source) bool {
	if a.Span.Start != b.Span.Start {
		return a.Span.Start < b.Span.Start
	}
	return cmp.Or(cmp.Compare(a.Span.End, b.Span.End), cmp.Compare(a.Parameter, b.Parameter), cmp.Compare(a.Path, b.Path), cmp.Compare(a.Kind, b.Kind)) < 0
}

// equal reports two values that hold the same facts. Element sets are
// compared by identity, which merge keeps when nothing changes.
func equal(v, w Value) bool {
	if !sameInts(v.interval, w.interval) || v.tested != w.tested || v.unless != w.unless || v.unlessFalse != w.unlessFalse || v.unlessNull != w.unlessNull || v.skippedFalse != w.skippedFalse || v.skippedNull != w.skippedNull || v.matched != w.matched || v.entry != w.entry || v.plain != w.plain || v.deep != w.deep {
		return false
	}
	if v.Complete != w.Complete || v.Safe != w.Safe || v.headerMask != w.headerMask || v.NonFalse != w.NonFalse || v.NonNull != w.NonNull || v.LengthKnown != w.LengthKnown || v.LengthKnown && v.Length != w.Length || v.Expr != w.Expr || v.elements != w.elements || !sameOrigins(v.origins, w.origins) || v.upload != w.upload || v.scalars != w.scalars || v.Identity != w.Identity || v.Invalidated != w.Invalidated || len(v.Sources) != len(w.Sources) {
		return false
	}
	if len(v.Sources) > 0 && &v.Sources[0] == &w.Sources[0] {
		return true
	}
	for i, s := range v.Sources {
		if s != w.Sources[i] {
			return false
		}
	}
	return true
}

func clone(f frame) frame {
	n := frame{vars: make(map[string]Value, len(f.vars)), states: map[uint32]State{}, definite: f.definite, wild: f.wild, related: f.related}
	for k, v := range f.vars {
		n.vars[k] = v
	}
	for k, v := range f.states {
		n.states[k] = v
	}
	if len(f.aliased) > 0 {
		n.aliased = make(map[string]uint32, len(f.aliased))
		for k, group := range f.aliased {
			n.aliased[k] = group
		}
	}
	return n
}

// alias records a name whose storage is shared outside the tracked bindings.
// Two different bindings of one name leave only that it is shared somewhere,
// which then holds for every name bound together with it.
func (f *frame) alias(name string, group uint32) {
	if name == "" {
		return
	}
	if f.aliased == nil {
		f.aliased = map[string]uint32{}
	}
	old, known := f.aliased[name]
	switch {
	case known && (group == aliasHeld || group == old):
		return
	case !known || old == aliasHeld:
		f.aliased[name] = group
	default:
		for other, g := range f.aliased {
			if g >= aliasGroup && (g == old || g == group) {
				f.aliased[other] = aliasAny
			}
		}
		f.aliased[name] = aliasAny
	}
	// What related the name to another one may no longer hold.
	f.related = nil
}

// bind makes target a reference to source, as `$target = &$source` does: the
// two names, and the names already bound with source, now share one value.
func (f *frame) bind(target, source string, site uint32) {
	group, known := f.aliased[source]
	if !known || group == aliasHeld {
		group = site + aliasGroup
	}
	if f.aliased == nil {
		f.aliased = map[string]uint32{}
	}
	f.related = nil
	f.aliased[source] = group
	f.aliased[target] = group
}

// write binds a plain variable. A write through a name bound by `=&` gives
// the same value to the names bound with it. A write through a name whose
// reference is kept somewhere the analysis does not follow may reach every
// other such name: what they held is gone, and the request data written may
// now be read under them. The walk is charged to the scope's budget.
func (f *frame) write(r *scopeResult, name string, v Value) {
	f.unmatch(r, name)
	if len(f.related) > 0 {
		f.rewritten(r, name, v)
	}
	if group, shared := f.aliased[name]; shared && group == aliasAny && f.wild {
		f.unsettle(r, v)
	}
	if group, shared := f.aliased[name]; shared && group != aliasHeld {
		r.operations += len(f.aliased)
		for other, g := range f.aliased {
			switch {
			case other == name || g != group:
			case group == aliasAny:
				f.vars[other] = unsettled(f.vars[other], v)
			default:
				f.vars[other] = v
			}
		}
	}
	f.vars[name] = v
}

// detach makes name a variable of its own again, as `unset`, `static` and
// `global` do: the storage it was bound to keeps its value for the other
// names, and the name starts over with nothing known.
func (f *frame) detach(name string) {
	f.related = nil
	delete(f.aliased, name)
	f.vars[name] = Value{}
}

// through is a write to storage the analysis does not follow (a property, an
// array slot): it may reach any variable whose reference is kept there.
func (f *frame) through(r *scopeResult, v Value) {
	if len(f.aliased) == 0 {
		return
	}
	f.related = nil
	r.operations += len(f.aliased)
	for other, g := range f.aliased {
		if g == aliasAny {
			f.vars[other] = unsettled(f.vars[other], v)
		}
	}
}

// forget discards what is known about every binding and operation, as after
// a write whose target the analysis does not follow. Request data stays.
func (f *frame) forget(r *scopeResult) {
	f.definite = false
	f.unsettle(r, Value{})
	f.states = map[uint32]State{}
}

// unsettle makes every binding unknown, keeping its request data and adding
// what was written.
func (f *frame) unsettle(r *scopeResult, written Value) {
	f.related = nil
	r.operations += len(f.vars)
	for name, old := range f.vars {
		f.vars[name] = unsettled(old, written)
	}
}

// unsettled is the value of a binding that code the analysis does not follow
// may have replaced. Nothing is known about it any more, except that request
// data it held, or that such code writes (written), may still be there:
// unknown must not read as untainted. The same goes for what a parameter
// brought in, which is request data when a caller passes some.
func unsettled(old, written Value) Value {
	v := Value{}
	if len(old.Sources) > 0 {
		v = Value{Complete: true, Sources: old.Sources, Safe: old.Safe}
	}
	if len(written.Sources) > 0 {
		if !v.Complete {
			v = Value{Complete: true}
		}
		v = merge(v, Value{Complete: true, Sources: written.Sources, Safe: written.Safe})
	}
	v.origins = origins(Value{}, written)
	return v
}

// tainted reports a retained request source.
func tainted(v Value) bool {
	for _, s := range v.Sources {
		if s.Parameter < 0 {
			return true
		}
	}
	return false
}

// agree keeps an operation every path performed; its success must hold on each.
func agree(a, b State) (State, bool) {
	if a.Operation != b.Operation || a.Span != b.Span {
		return State{}, false
	}
	a.Successful = a.Successful && b.Successful
	return a, true
}

// rejoin joins a path that split off inside an expression into the first
// one, charging the frame copied for it and the join to the scope's budget.
func rejoin(first, second frame, r *scopeResult) frame {
	_, work := first.absorb(second)
	r.copies += work + len(second.vars) + len(second.states)
	return first
}

// absorb joins another path into the frame in place. It reports whether the
// frame changed and the work done, in bindings visited and sources compared.
func (f *frame) absorb(b frame) (bool, int) {
	return f.absorbAt(b, false)
}

func (f *frame) absorbAt(b frame, head bool) (bool, int) {
	changed, work, shared := false, len(f.vars)+len(f.states), 0
	// What relates two variables is joined from what each path holds of
	// them, before their values are.
	if len(f.related)+len(b.related) > 0 && f.correlated(b) {
		changed = true
	}
	for k, v := range f.vars {
		w, ok := b.vars[k]
		if ok {
			shared++
		}
		if ok && equal(v, w) && (len(v.Sources) > 0 || v.Safe == HTML|SQL|Shell|Header|URL|Path && v.upload == 0 && !v.scalars) {
			// Joining a value with itself leaves it as it is, once a value
			// without request data has the form a join gives it.
			continue
		}
		if len(v.Sources) > 0 && len(w.Sources) > 0 && &v.Sources[0] != &w.Sources[0] {
			work += (len(v.Sources) + len(w.Sources)) / 2
		}
		m := merge(v, w)
		if head {
			m.interval = widenInterval(v, m)
		}
		m.correlate(v, w)
		if !ok && b.definite && !reserved(k) {
			// The other path never assigned the variable: it is null there.
			m.NonFalse, m.plain, m.deep = v.NonFalse, v.plain, v.deep
		}
		if m.origins = origins(v, w); !equal(m, v) {
			f.vars[k] = m
			changed = true
		}
	}
	if shared < len(b.vars) {
		work += len(b.vars)
		for k, v := range b.vars {
			if _, ok := f.vars[k]; ok {
				continue
			}
			m := merge(Value{}, v)
			m.origins = origins(Value{}, v)
			if f.definite && !reserved(k) {
				m.NonFalse, m.plain, m.deep = v.NonFalse, v.plain, v.deep
			}
			f.vars[k] = m
			changed = true
		}
	}
	for k, v := range f.states {
		if common, ok := agree(v, b.states[k]); !ok {
			delete(f.states, k)
			changed = true
		} else if common != v {
			f.states[k] = common
			changed = true
		}
	}
	if f.definite && !b.definite || !f.wild && b.wild {
		f.definite, f.wild = f.definite && b.definite, true && (f.wild || b.wild)
		changed = true
	}
	// A definite binding one path lacks, or holds otherwise, may not hold.
	for k, group := range f.aliased {
		if other, both := b.aliased[k]; group >= aliasGroup && (!both || other == aliasHeld) {
			f.alias(k, aliasAny)
			changed = true
		}
	}
	for k, group := range b.aliased {
		mine, both := f.aliased[k]
		if group >= aliasGroup && (!both || mine == aliasHeld) {
			// The names it is bound with on the other path may not be either.
			f.alias(k, group)
			group = aliasAny
		}
		f.alias(k, group)
		changed = changed || !both || f.aliased[k] != mine
	}
	return changed, work
}

// correlate gives the join of two paths what both know of a binding through
// a short-circuit operation: that its left operand read the binding last,
// and what holds of the binding unless its right operand was skipped.
func (v *Value) correlate(a, b Value) {
	if a.tested == b.tested {
		v.tested = a.tested
	}
	if a.unless == b.unless {
		v.unless, v.unlessFalse, v.unlessNull = a.unless, a.unlessFalse && b.unlessFalse, a.unlessNull && b.unlessNull
		v.skippedFalse, v.skippedNull = a.skippedFalse && b.skippedFalse, a.skippedNull && b.skippedNull
	}
}

// unguarded returns the set without any exclusion.
func (s *originSet) unguarded() *originSet {
	if s == nil {
		return nil
	}
	n := &originSet{items: make([]Origin, len(s.items)), full: s.full}
	for i, o := range s.items {
		n.items[i] = Origin{Expr: o.Expr}
	}
	return n
}

// reserve bounds the operation states remembered per node.
func reserve(r *scopeResult, cells int) bool {
	if cells > MaxTransfers-r.facts {
		r.spent = true
		r.complete = false
		return false
	}
	r.facts += cells
	return true
}

// retain bounds the frame cells kept as block inputs.
func retain(r *scopeResult, cells int) bool {
	if cells > MaxRetained-r.retained {
		r.spent = true
		r.complete = false
		return false
	}
	r.retained += cells
	return true
}

// retained reports effects useful to either request-data or failure analysis.
func retained(v Value) bool { return tainted(v) || v.Expr != nil || v.origins != nil }

// effects joins local writes, retaining their syntax only in the environment.
func effects(a, b Value) Value {
	v := merge(a, b)
	v.origins = origins(a, b)
	return v
}

// producers exports local syntax only after the whole scope was followed.
// Partial generator or budget-limited bodies cannot establish that a write ran.
func (r *scopeResult) producers(v Value) *originSet {
	if !r.complete || r.generator {
		return nil
	}
	return origins(Value{}, v)
}
