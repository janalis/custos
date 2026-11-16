package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// Value returns the joined value of an expression at its execution point.
func (e *Env) Value(x syntax.Expr) Value {
	if x == nil {
		return Value{Complete: true}
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v, ok := r.values.get(x)
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
	r := e.scope(syntax.EnclosingVariableScope(x))
	v := r.values.at(x)
	return v.Expr, r.complete && v.Expr != nil
}

// Excludes proves that a checked variable cannot equal false or null here.
// The exclusions are facts of the paths that reach the read: they hold of a
// value whose inputs are not all known, in a scope followed to its end.
func (e *Env) Excludes(x syntax.Expr, literal string) bool {
	if x == nil {
		return false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v := r.values.at(x)
	if !r.complete && !beforeCut(r, x) {
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

// beforeCut reports whether x is evaluated before the dynamic code that made
// its scope incomplete, and outside any loop that also holds that code.
func beforeCut(r *scopeResult, x syntax.Node) bool {
	if r.cut == 0 || x.Span().End > r.cut {
		return false
	}
	for p := x.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if p.Span().End > r.cut {
				return false
			}
		}
	}
	return true
}

// Reaching returns the expressions that may have produced the value of a
// plain variable read, in the order the paths were joined, each with the
// failure values its paths exclude. The second result reports that the paths
// agree on one producer. Otherwise the list may leave out paths on which the
// variable is unassigned or holds a value the scope does not account for,
// and it is empty when nothing is known or the host scope budget is exhausted.
// When CompressedOrigins is true, items are finite-contract representatives,
// not a complete list suitable for arbitrary expression predicates.
func (e *Env) Reaching(x syntax.Expr) ([]Origin, bool) {
	if x == nil {
		return nil, false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	v, ok := r.values.get(x)
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

// Replaced reports a read whose binding code the analysis does not follow
// (a callee that may write the globals table) may have replaced since it
// was last written, on some path.
func (e *Env) Replaced(x syntax.Expr) bool {
	return e.Value(x).replaced
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
	v := r.values.at(x)
	return r.complete && !v.Safe.Covers(context) && tainted(v)
}

// Matched reports a value that is the text of a `switch` label the request
// data it was read from was found equal to: text the program wrote.
func (e *Env) Matched(x syntax.Expr) bool {
	if x == nil {
		return false
	}
	r := e.scope(syntax.EnclosingVariableScope(x))
	return r.complete && r.values.at(x).matched
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
	s, ok := r.before[syntax.Identity(at)][v.Identity]
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
	state, ok := r.before[syntax.Identity(at)][id]
	return state, ok
}
