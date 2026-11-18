package flow

import (
	"sort"

	"custos/internal/php/syntax"
)

// Labeled reports that a jump to a `goto` label can land between what a
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
func (e *Env) Labeled(at syntax.Node) bool {
	return e.LabelledBefore(at, at)
}

// LabelledBefore is Labeled for a value checked in one place and used in
// another: the labels count up to use, the variables are the ones of checked.
func (e *Env) LabelledBefore(checked, use syntax.Node) bool {
	scope := syntax.EnclosingVariableScope(use)
	r := e.scope(scope)
	if r.labeled == 0 || r.labeled > use.Span().Start {
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
