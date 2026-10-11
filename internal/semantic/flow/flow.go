// Package flow provides bounded, file-local value and state analysis.
// Environments are confined to one analysis; snapshots contain no syntax trees.
package flow

import (
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
// MaxCopied the bindings copied while iterating to a fixed point (time). An ordinary long function carries a hundred locals through a few
// hundred blocks, far below both.
const (
	MaxBlocks    = 8192
	MaxTransfers = 100000
	MaxRetained  = 4 * MaxTransfers
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
	Sources     []Source
	Safe        Context
	Complete    bool
	Expr        syntax.Expr
	Identity    uint32
	Invalidated uint32
	NonFalse    bool
	NonNull     bool
	// elements marks an array variable some of whose literal-keyed elements
	// are tracked as bindings of their own.
	elements    bool
	Length      int64
	LengthKnown bool
	headerMask  uint8
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

// Env shares one bounded analysis per variable scope among all inspections.
type Env struct {
	file     *syntax.File
	types    *infer.Env
	snapshot *Snapshot
	scopes   map[syntax.Node]*scopeResult
	// global caches the file's function-bound global names once scanned, and
	// globalTaint the request data those functions write to them.
	global       map[string]bool
	globalTaint  map[string]Value
	writers      map[globalWriter]bool
	globalsKnown bool
}

type scopeResult struct {
	values     map[syntax.Expr]Value
	before     map[syntax.Node]map[uint32]State
	calls      []Call
	returns    []Value
	callIndex  map[syntax.Node]int
	complete   bool
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
}

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
func (e *Env) Excludes(x syntax.Expr, literal string) bool {
	v := e.Value(x)
	if !v.Complete {
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
	if len(a.Sources) == 0 {
		sa = HTML | SQL | Shell | Header | URL | Path
	}
	if len(b.Sources) == 0 {
		sb = HTML | SQL | Shell | Header | URL | Path
	}
	v := Value{Complete: a.Complete && b.Complete, Safe: sa & sb, headerMask: a.headerMask & b.headerMask, NonFalse: a.NonFalse && b.NonFalse, NonNull: a.NonNull && b.NonNull, LengthKnown: a.LengthKnown && b.LengthKnown && a.Length == b.Length, Length: a.Length}
	v.elements = a.elements && b.elements
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
	for _, xs := range [][]Source{a.Sources, b.Sources} {
		for _, s := range xs {
			found := false
			for _, t := range v.Sources {
				if s == t {
					found = true
					break
				}
			}
			if !found {
				if len(v.Sources) == MaxSources {
					v.Complete = false
					continue
				}
				v.Sources = append(v.Sources, s)
			}
		}
	}
	return v
}

func clone(f frame) frame {
	n := frame{vars: map[string]Value{}, states: map[uint32]State{}}
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
	case !known || old == aliasHeld:
		f.aliased[name] = group
	case group == aliasHeld || group == old:
	default:
		for other, g := range f.aliased {
			if g >= aliasGroup && (g == old || g == group) {
				f.aliased[other] = aliasAny
			}
		}
		f.aliased[name] = aliasAny
	}
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
	f.aliased[source] = group
	f.aliased[target] = group
}

// set stores a binding. An array whose value is replaced or no longer known
// loses the elements tracked for it.
func (f *frame) set(name string, v Value) {
	if f.vars[name].elements && !v.elements {
		for k := range f.vars {
			if element(k, name) {
				delete(f.vars, k)
			}
		}
	}
	f.vars[name] = v
}

// elementKey names the binding of a literal-keyed element of an array variable.
func elementKey(name, key string) string { return name + "\x00" + key }

// element reports the binding of an element; of array name when name is given.
func element(key, name string) bool {
	i := strings.IndexByte(key, 0)
	return i >= 0 && (name == "" || key[:i] == name)
}

// write binds a plain variable. A write through a name bound by `=&` gives
// the same value to the names bound with it. A write through a name whose
// reference is kept somewhere the analysis does not follow may reach every
// other such name: what they held is gone, and the request data written may
// now be read under them. The walk is charged to the scope's budget.
func (f *frame) write(r *scopeResult, name string, v Value) {
	if group, shared := f.aliased[name]; shared && group != aliasHeld {
		r.operations += len(f.aliased)
		for other, g := range f.aliased {
			switch {
			case other == name || g != group:
			case group == aliasAny:
				f.set(other, unsettled(f.vars[other], v))
			default:
				f.set(other, v)
			}
		}
	}
	f.set(name, v)
}

// through is a write to storage the analysis does not follow (a property, an
// array slot): it may reach any variable whose reference is kept there.
func (f *frame) through(r *scopeResult, v Value) {
	if len(f.aliased) == 0 {
		return
	}
	r.operations += len(f.aliased)
	for other, g := range f.aliased {
		if g == aliasAny {
			f.set(other, unsettled(f.vars[other], v))
		}
	}
}

// forget discards what is known about every binding and operation, as after
// a write whose target the analysis does not follow. Request data stays.
func (f *frame) forget(r *scopeResult) {
	f.unsettle(r, Value{})
	f.states = map[uint32]State{}
}

// unsettle makes every binding unknown, keeping its request data and adding
// what was written. Tracked array elements are dropped with their arrays.
func (f *frame) unsettle(r *scopeResult, written Value) {
	r.operations += len(f.vars)
	for name, old := range f.vars {
		if element(name, "") {
			delete(f.vars, name)
		} else {
			f.vars[name] = unsettled(old, written)
		}
	}
}

// unsettled is the value of a binding that code the analysis does not follow
// may have replaced. Nothing is known about it any more, except that request
// data it held, or that such code writes (written), may still be there:
// unknown must not read as untainted.
func unsettled(old, written Value) Value {
	v := Value{}
	if tainted(old) {
		v = Value{Complete: true, Sources: old.Sources, Safe: old.Safe}
	}
	if tainted(written) {
		if !v.Complete {
			v = Value{Complete: true}
		}
		v = merge(v, Value{Complete: true, Sources: written.Sources, Safe: written.Safe})
	}
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

func join(a, b frame) frame {
	n := clone(a)
	for k, v := range n.vars {
		w, ok := b.vars[k]
		if !ok {
			// An element tracked on one path only is read from its array.
			if element(k, "") {
				delete(n.vars, k)
				continue
			}
			w = Value{}
		}
		n.vars[k] = merge(v, w)
	}
	for k, v := range b.vars {
		if _, ok := n.vars[k]; !ok && !element(k, "") {
			n.vars[k] = merge(Value{}, v)
		}
	}
	for k, v := range n.states {
		if common, ok := agree(v, b.states[k]); ok {
			n.states[k] = common
		} else {
			delete(n.states, k)
		}
	}
	// A definite binding made on one path only may not hold.
	for k, group := range n.aliased {
		if _, both := b.aliased[k]; !both && group >= aliasGroup {
			n.alias(k, aliasAny)
		}
	}
	for k, group := range b.aliased {
		if _, both := a.aliased[k]; !both && group >= aliasGroup {
			group = aliasAny
		}
		n.alias(k, group)
	}
	return n
}

func same(a, b frame) bool {
	if len(a.vars) != len(b.vars) || len(a.states) != len(b.states) || len(a.aliased) != len(b.aliased) {
		return false
	}
	for k, group := range a.aliased {
		if other, ok := b.aliased[k]; !ok || other != group {
			return false
		}
	}
	for k, v := range a.vars {
		w, ok := b.vars[k]
		if !ok || v.Complete != w.Complete || v.Safe != w.Safe || v.headerMask != w.headerMask || v.NonFalse != w.NonFalse || v.NonNull != w.NonNull || v.LengthKnown != w.LengthKnown || v.LengthKnown && v.Length != w.Length || v.Expr != w.Expr || v.elements != w.elements || v.Identity != w.Identity || v.Invalidated != w.Invalidated || len(v.Sources) != len(w.Sources) {
			return false
		}
		for i, s := range v.Sources {
			if s != w.Sources[i] {
				return false
			}
		}
	}
	for k, v := range a.states {
		if b.states[k] != v {
			return false
		}
	}
	return true
}

// reserve bounds the operation states remembered per node.
func reserve(r *scopeResult, cells int) bool {
	if cells > MaxTransfers-r.facts {
		r.complete = false
		return false
	}
	r.facts += cells
	return true
}

// retain bounds the frame cells kept as block inputs.
func retain(r *scopeResult, cells int) bool {
	if cells > MaxRetained-r.retained {
		r.complete = false
		return false
	}
	r.retained += cells
	return true
}
