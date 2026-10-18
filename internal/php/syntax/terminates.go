package syntax

import (
	"math"
	"strconv"
	"strings"
	"sync/atomic"
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
		if _, ok := UnwrapParens(x.Expr).(*Throw); ok {
			return true
		}
		return ExitInvocation(x.Expr)
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
	case *While, *For:
		return Endless(x)
	case *DoWhile:
		// A break leaves the loop, and a continue reaches a condition that
		// may end it.
		return Endless(x) || !containsBreak(x.Body) && Terminates(x.Body) && !jumps(x.Body, true, false)
	case *Declare:
		return x.Body != nil && Terminates(x.Body)
	}
	return false
}

// Endless reports a loop that control never falls out of, wherever it
// entered it: `while (true)`, `for (;;)` or `do … while (true)` without a
// `break` or `goto` that leaves it. The search for those jumps is cached on the
// loop, so that asking for every label of a long body stays linear overall.
func Endless(n Node) bool {
	switch x := n.(type) {
	case *While:
		return memoEndless(&x.endless, x.Body, x.Cond, false)
	case *DoWhile:
		return memoEndless(&x.endless, x.Body, x.Cond, false)
	case *For:
		var cond Expr
		if len(x.Cond) != 0 {
			cond = x.Cond[len(x.Cond)-1]
		}
		return memoEndless(&x.endless, x.Body, cond, len(x.Cond) == 0)
	}
	return false
}

// memoEndless caches both literal truth and escaping jumps, so repeated
// label queries do not repeatedly decode a large string condition.
func memoEndless(memo *int32, body Stmt, cond Expr, unconditional bool) bool {
	// memo: 0 = not computed yet, 1 = may end, 2 = endless.
	v := atomic.LoadInt32(memo)
	if v == 0 {
		v = 1
		if (unconditional || isTrueConst(cond)) && !jumps(body, false, true) {
			v = 2
		}
		atomic.StoreInt32(memo, v)
	}
	return v == 2
}

func isTrueConst(e Expr) bool {
	truth, known := constantTruth(e)
	return known && truth
}

// constantTruth recognizes literal PHP truth values without resolving names.
func constantTruth(e Expr) (bool, bool) {
	switch x := UnwrapParens(e).(type) {
	case *ConstFetch:
		if x.Name != nil {
			switch strings.ToLower(strings.TrimPrefix(x.Name.Value, `\`)) {
			case "true":
				return true, true
			case "false", "null":
				return false, true
			}
		}
	case *Literal:
		switch x.LitKind {
		case LitInt:
			v, err := strconv.ParseUint(strings.ReplaceAll(x.Raw, "_", ""), 0, 64)
			return v != 0, err == nil
		case LitString:
			value, ok := LiteralStringValue(x)
			return value != "" && value != "0", ok
		}
	case *Unary:
		if x.Op.Kind == TMinus || x.Op.Kind == TPlus {
			if literal, ok := UnwrapParens(x.Expr).(*Literal); ok && literal.LitKind == LitInt {
				return constantTruth(literal)
			}
		}
		if x.Op.Kind == TExclaim {
			truth, known := constantTruth(x.Expr)
			return !truth, known
		}
	}
	return false, false
}

// containsBreak reports whether a jump that could make control fall out of
// the construct rooted at n appears inside it: a `break` whose level reaches
// past the loops and switches nested in n (nested function-likes are
// skipped), or — when n is a switch — a `continue` targeting n, which PHP
// treats as `break`. A non-literal level (PHP 5.3 `break $n`) is assumed to
// leave n.
func containsBreak(n Node) bool { return jumps(n, false, false) }

// jumps looks for a `break` leaving the construct rooted at n or, when
// continues is set, for a `continue` that targets the loop n is the body of.
func jumps(n Node, continues, escapingGotos bool) bool {
	if n == nil {
		return false
	}
	_, rootSwitch := n.(*Switch)
	found := false
	var gotos []*Goto
	var labels map[string]bool
	var walk func(m Node, depth int)
	walk = func(m Node, depth int) {
		if found {
			return
		}
		switch x := m.(type) {
		case *Function, *Method, *Closure, *ArrowFunction, *PropertyHook, *ClassLike:
			return
		case *Goto:
			if escapingGotos {
				gotos = append(gotos, x)
			}
			return
		case *Label:
			if escapingGotos {
				if labels == nil {
					labels = map[string]bool{}
				}
				labels[x.Name.Value] = true
			}
		case *Break:
			found = !continues && jumpLevel(x.Num) > depth
			return
		case *Continue:
			// A computed level may be the one that targets n.
			level := jumpLevel(x.Num)
			found = (rootSwitch || continues) && (level == depth+1 || level == math.MaxInt32)
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
	for _, g := range gotos {
		if !labels[g.Label.Value] {
			return true
		}
	}
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
