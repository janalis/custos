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

// Limits bound hostile scopes and recursive wrapper queries.
const (
	MaxBlocks    = 8192
	MaxTransfers = 100000
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
}

// Env shares one bounded analysis per variable scope among all inspections.
type Env struct {
	file     *syntax.File
	types    *infer.Env
	snapshot *Snapshot
	scopes   map[syntax.Node]*scopeResult
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

// Tainted reports a retained external source without a context-specific escape.
// Incomplete results may retain known unsafe flows, but never prove safety.
func (e *Env) Tainted(x syntax.Expr, context Context) bool {
	v := e.Value(x)
	if !v.Complete {
		return false
	}
	if v.Safe&context != 0 {
		return false
	}
	for _, s := range v.Sources {
		if s.Parameter < 0 {
			return true
		}
	}
	return false
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
	sa, sb := a.Safe, b.Safe
	if len(a.Sources) == 0 && a.Complete {
		sa = HTML | SQL | Shell | Header | URL | Path
	}
	if len(b.Sources) == 0 && b.Complete {
		sb = HTML | SQL | Shell | Header | URL | Path
	}
	v := Value{Complete: a.Complete && b.Complete, Safe: sa & sb, headerMask: a.headerMask & b.headerMask, NonFalse: a.NonFalse && b.NonFalse, NonNull: a.NonNull && b.NonNull, LengthKnown: a.LengthKnown && b.LengthKnown && a.Length == b.Length, Length: a.Length}
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
	return n
}

func join(a, b frame) frame {
	n := clone(a)
	for k, v := range n.vars {
		w, ok := b.vars[k]
		if !ok {
			w = Value{}
		}
		n.vars[k] = merge(v, w)
	}
	for k, v := range b.vars {
		if _, ok := n.vars[k]; !ok {
			n.vars[k] = merge(Value{}, v)
		}
	}
	for k, v := range n.states {
		if b.states[k] != v {
			delete(n.states, k)
		}
	}
	return n
}

func same(a, b frame) bool {
	if len(a.vars) != len(b.vars) || len(a.states) != len(b.states) {
		return false
	}
	for k, v := range a.vars {
		w, ok := b.vars[k]
		if !ok || v.Complete != w.Complete || v.Safe != w.Safe || v.headerMask != w.headerMask || v.NonFalse != w.NonFalse || v.NonNull != w.NonNull || v.LengthKnown != w.LengthKnown || v.LengthKnown && v.Length != w.Length || v.Expr != w.Expr || v.Identity != w.Identity || v.Invalidated != w.Invalidated || len(v.Sources) != len(w.Sources) {
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

// reserve bounds copied and retained frame cells as well as execution work.
func reserve(r *scopeResult, cells int) bool {
	if cells > MaxTransfers-r.facts {
		r.complete = false
		return false
	}
	r.facts += cells
	return true
}
