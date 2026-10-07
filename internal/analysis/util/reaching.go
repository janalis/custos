package util

import "custos/internal/syntax"

// ReachingAssignments returns the plain `=` assignments (by value, as in
// value discovery) to the local variable $name in scope's body that may
// reach the program point at, in source order, and whether the variable's
// entry value (parameter default, or undefined) may reach it. scope is a
// function, method, closure or arrow function containing at.
//
// It is a structural approximation of reaching definitions:
//   - an assignment written before at reaches it unless a later
//     unconditional write dominates at (an assignment statement in a
//     statement list enclosing at, or the value/key variable of a foreach
//     whose body encloses at);
//   - an assignment written after at (or containing it, as in
//     `$v = f($v)`) reaches it only through a loop enclosing both, unless a
//     write inside that loop dominates at;
//   - unreachable assignments (after return/throw/...) never reach;
//   - assignments inside nested functions, closures and arrow functions
//     belong to another scope and are ignored, except in closures importing
//     $name by reference, which may run at any time and always count.
func ReachingAssignments(scope, at syntax.Node, name string) (defs []*syntax.Assign, entry bool) {
	_, body := scopeParts(scope)
	if body == nil || name == "" {
		return nil, true
	}
	atStart := at.Span().Start
	var killers []syntax.Node // unconditional writes dominating at
	var all []*syntax.Assign
	var always []*syntax.Assign // from by-reference closures
	var walk func(n syntax.Node, byRefClosure bool)
	walk = func(n syntax.Node, byRefClosure bool) {
		syntax.Children(n, func(c syntax.Node) {
			switch x := c.(type) {
			case *syntax.Function, *syntax.ClassLike, *syntax.ArrowFunction:
				return
			case *syntax.Closure:
				for _, u := range x.Uses {
					if u.ByRef && u.Var != nil && u.Var.Name == name {
						walk(x, true)
						return
					}
				}
				return
			case *syntax.Assign:
				if t, ok := UnwrapParens(x.Var).(*syntax.Variable); ok && t.NameExpr == nil && t.Name == name {
					if x.Op.Kind == syntax.TEqual && !x.ByRef {
						if byRefClosure {
							always = append(always, x)
						} else {
							all = append(all, x)
						}
					}
					if !byRefClosure && reachDominates(x, at, atStart) {
						killers = append(killers, x)
					}
				}
			case *syntax.Foreach:
				if !byRefClosure && x.Body != nil && NodeContains(x.Body, at) && (reachIsVarNamed(x.Value, name) || reachIsVarNamed(x.Key, name)) {
					killers = append(killers, x)
				}
			}
			walk(c, byRefClosure)
		})
	}
	walk(body, false)

	lastKill := func(within syntax.Node) uint32 {
		var pos uint32
		found := false
		for _, k := range killers {
			if within != nil && !NodeContains(within, k) {
				continue
			}
			if p := reachKillPos(k); !found || p > pos {
				pos, found = p, true
			}
		}
		if !found {
			return 0
		}
		return pos
	}
	hasKill := func(within syntax.Node) bool {
		for _, k := range killers {
			if within == nil || NodeContains(within, k) {
				return true
			}
		}
		return false
	}

	entry = !hasKill(nil)
	kill := lastKill(nil)
	for _, a := range all {
		if !Reachable(a, scope) {
			continue
		}
		if a.Span().End <= atStart { // written before at
			if !hasKill(nil) || a.Span().Start >= kill {
				defs = append(defs, a)
			}
			continue
		}
		// written after at: only through an enclosing loop's back edge
		for l := a.Parent(); l != nil && l != scope; l = l.Parent() {
			if !reachIsLoop(l) || !NodeContains(l, at) {
				continue
			}
			if !hasKill(l) {
				defs = append(defs, a)
			}
			break
		}
	}
	return append(defs, always...), entry
}

// reachDominates reports whether the assignment a is a statement executed on
// every path to at: its expression statement sits in a statement list (block
// or case) enclosing at and ends before at starts.
func reachDominates(a *syntax.Assign, at syntax.Node, atStart uint32) bool {
	if a.Span().End > atStart {
		return false
	}
	var n syntax.Node = a
	for {
		p, ok := n.Parent().(*syntax.Paren)
		if !ok {
			break
		}
		n = p
	}
	st, ok := n.Parent().(*syntax.ExprStmt)
	if !ok {
		return false
	}
	switch list := st.Parent().(type) {
	case *syntax.Block, *syntax.Case:
		return NodeContains(list, at)
	}
	return false
}

func reachKillPos(k syntax.Node) uint32 {
	if f, ok := k.(*syntax.Foreach); ok && f.Body != nil {
		return f.Body.Span().Start
	}
	return k.Span().Start
}

func reachIsLoop(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
		return true
	}
	return false
}

func reachIsVarNamed(e syntax.Expr, name string) bool {
	if e == nil {
		return false
	}
	v, ok := UnwrapParens(e).(*syntax.Variable)
	return ok && v.NameExpr == nil && v.Name == name
}

// NodeContains reports whether n is outer or a descendant of it.
func NodeContains(outer, n syntax.Node) bool {
	for ; n != nil; n = n.Parent() {
		if n == outer {
			return true
		}
	}
	return false
}
