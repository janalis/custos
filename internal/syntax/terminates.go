package syntax

import (
	"math"
	"strconv"
	"strings"
)

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
		switch UnwrapParens(x.Expr).(type) {
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
	c, ok := UnwrapParens(e).(*ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "true")
}

// containsBreak reports whether a jump that could make control fall out of
// the construct rooted at n appears inside it: a `break` whose level reaches
// past the loops and switches nested in n (nested function-likes are
// skipped), or — when n is a switch — a `continue` targeting n, which PHP
// treats as `break`. A non-literal level (PHP 5.3 `break $n`) is assumed to
// leave n.
func containsBreak(n Node) bool {
	if n == nil {
		return false
	}
	_, rootSwitch := n.(*Switch)
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
			found = jumpLevel(x.Num) > depth
			return
		case *Continue:
			found = rootSwitch && jumpLevel(x.Num) == depth+1
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

// jumpLevel is the number of enclosing loops/switches a break/continue
// leaves: 1 without an argument, the literal's value otherwise (0 means 1,
// as PHP 5.3), and a huge level for anything else.
func jumpLevel(num Expr) int {
	if num == nil {
		return 1
	}
	l, ok := UnwrapParens(num).(*Literal)
	if !ok || l.LitKind != LitInt {
		return math.MaxInt32
	}
	v, err := strconv.ParseInt(strings.ReplaceAll(l.Raw, "_", ""), 0, 32)
	switch {
	case err != nil:
		return math.MaxInt32
	case v < 1:
		return 1
	}
	return int(v)
}
