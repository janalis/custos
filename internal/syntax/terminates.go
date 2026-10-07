package syntax

import "strings"

// Terminates reports whether control never falls through s to the next
// statement of its list: return, throw, exit/die, break, continue, goto,
// blocks containing such a statement, if/else chains (with an else) whose
// every branch terminates, try statements whose body and catches all
// terminate (or whose finally terminates), switches with a default where
// every case ends abruptly without break, and infinite loops
// (`while (true)`, `for (;;)`, `do … while (true)`) without break.
func Terminates(s Stmt) bool {
	switch x := s.(type) {
	case nil:
		return false
	case *Return, *Break, *Continue, *Goto:
		return true
	case *ExprStmt:
		switch unwrapParens(x.Expr).(type) {
		case *Throw, *Exit:
			return true
		}
		return false
	case *Block:
		return FirstTerminating(x) < len(x.Stmts)
	case *If:
		if x.Else == nil || !Terminates(x.Body) || !Terminates(x.Else.Body) {
			return false
		}
		for _, ei := range x.ElseIfs {
			if !Terminates(ei.Body) {
				return false
			}
		}
		return true
	case *Try:
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
	case *Switch:
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
		last := x.Cases[len(x.Cases)-1]
		return FirstTerminating(last) < len(last.Stmts)
	case *While:
		return isTrueConst(x.Cond) && !containsBreak(x.Body)
	case *DoWhile:
		return Terminates(x.Body) || (isTrueConst(x.Cond) && !containsBreak(x.Body))
	case *For:
		return len(x.Cond) == 0 && !containsBreak(x.Body)
	case *Declare:
		return x.Body != nil && Terminates(x.Body)
	}
	return false
}

func isTrueConst(e Expr) bool {
	c, ok := unwrapParens(e).(*ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "true")
}

func unwrapParens(e Expr) Expr {
	for {
		p, ok := e.(*Paren)
		if !ok || p.Expr == nil {
			return e
		}
		e = p.Expr
	}
}

// containsBreak reports whether a `break` that could leave the construct
// rooted at n appears inside it (nested loops/switches are skipped unless
// the break has a level > 1; nested function-likes are skipped).
func containsBreak(n Node) bool {
	if n == nil {
		return false
	}
	found := false
	var walk func(m Node, depth int)
	walk = func(m Node, depth int) {
		if found {
			return
		}
		switch x := m.(type) {
		case *Function, *Method, *Closure, *ArrowFunction, *ClassLike:
			return
		case *Break:
			lvl := 1
			if l, ok := x.Num.(*Literal); ok && l.LitKind == LitInt && l.Raw != "1" {
				lvl = 2
			}
			if lvl > depth {
				found = true
			}
			return
		case *While, *DoWhile, *For, *Foreach, *Switch:
			if m != n {
				Children(m, func(c Node) { walk(c, depth+1) })
				return
			}
		}
		Children(m, func(c Node) { walk(c, depth) })
	}
	walk(n, 0)
	return found
}
