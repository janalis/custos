package util

import "custos/internal/syntax"

// Terminates reports whether control never falls through s to the next
// statement of its list: return, throw, exit/die, break, continue, goto,
// blocks containing such a statement, if/else chains (with an else) whose
// every branch terminates, try statements whose body and catches all
// terminate (or whose finally terminates), switches with a default where
// every case ends abruptly without break, and infinite loops
// (`while (true)`, `for (;;)`, `do … while (true)`) without break.
func Terminates(s syntax.Stmt) bool {
	switch x := s.(type) {
	case nil:
		return false
	case *syntax.Return, *syntax.Break, *syntax.Continue, *syntax.Goto:
		return true
	case *syntax.ExprStmt:
		switch UnwrapParens(x.Expr).(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
		return false
	case *syntax.Block:
		return listTerminates(x.Stmts)
	case *syntax.If:
		if x.Else == nil || !Terminates(x.Body) || !Terminates(x.Else.Body) {
			return false
		}
		for _, ei := range x.ElseIfs {
			if !Terminates(ei.Body) {
				return false
			}
		}
		return true
	case *syntax.Try:
		if x.Finally != nil && x.Finally.Body != nil && Terminates(x.Finally.Body) {
			return true
		}
		if x.Body == nil || !Terminates(x.Body) {
			return false
		}
		for _, c := range x.Catches {
			if c.Body == nil || !Terminates(c.Body) {
				return false
			}
		}
		return true
	case *syntax.Switch:
		hasDefault := false
		for _, c := range x.Cases {
			if c.Cond == nil {
				hasDefault = true
			}
		}
		if !hasDefault || len(x.Cases) == 0 {
			return false
		}
		if containsBreak(x) {
			return false
		}
		return listTerminates(x.Cases[len(x.Cases)-1].Stmts)
	case *syntax.While:
		return isTrueConst(x.Cond) && !containsBreak(x.Body)
	case *syntax.DoWhile:
		return Terminates(x.Body) || (isTrueConst(x.Cond) && !containsBreak(x.Body))
	case *syntax.For:
		return len(x.Cond) == 0 && !containsBreak(x.Body)
	case *syntax.Declare:
		return x.Body != nil && Terminates(x.Body)
	}
	return false
}

func listTerminates(list []syntax.Stmt) bool {
	for _, s := range list {
		if Terminates(s) {
			return true
		}
	}
	return false
}

func isTrueConst(e syntax.Expr) bool {
	v, ok := BoolConst(UnwrapParens(e))
	return ok && v
}

// containsBreak reports whether a `break` that could leave the construct
// rooted at n appears inside it (nested loops/switches are skipped unless
// the break has a level > 1; nested function-likes are skipped).
func containsBreak(n syntax.Node) bool {
	if n == nil {
		return false
	}
	found := false
	var walk func(m syntax.Node, depth int)
	walk = func(m syntax.Node, depth int) {
		if found {
			return
		}
		switch x := m.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return
		case *syntax.Break:
			lvl := 1
			if l, ok := x.Num.(*syntax.Literal); ok && l.LitKind == syntax.LitInt && l.Raw != "1" {
				lvl = 2
			}
			if lvl > depth {
				found = true
			}
			return
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Switch:
			if m != n {
				syntax.Children(m, func(c syntax.Node) { walk(c, depth+1) })
				return
			}
		}
		syntax.Children(m, func(c syntax.Node) { walk(c, depth) })
	}
	walk(n, 0)
	return found
}

// Reachable reports whether node n (inside scope's body) can be reached
// from the scope entry: no statement on the path from n up to scope is
// preceded, in its statement list, by a statement that Terminates. Branches
// guarded by constant conditions are considered reachable. scope may be nil
// (file level).
func Reachable(n, scope syntax.Node) bool {
	var child syntax.Node = n
	for p := n.Parent(); p != nil && child != scope; p = p.Parent() {
		var list []syntax.Stmt
		switch x := p.(type) {
		case *syntax.Block:
			list = x.Stmts
		case *syntax.Case:
			list = x.Stmts
		case *syntax.Namespace:
			list = x.Stmts
		}
		for _, s := range list {
			if syntax.Node(s) == child {
				break
			}
			if Terminates(s) {
				return false
			}
		}
		if p == scope {
			break
		}
		child = p
	}
	return true
}
