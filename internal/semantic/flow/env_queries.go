package flow

import (
	"sort"
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
