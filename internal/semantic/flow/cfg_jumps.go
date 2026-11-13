package flow

import (
	"strconv"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// leaves returns the finally blocks a jump to label runs: those of the try
// statements it is in that do not hold the label, outermost first. PHP
// allows no jump into or out of a finally block itself.
func (g *graph) leaves(label string, finals []final) []final {
	if len(finals) == 0 {
		return nil
	}
	if g.written == nil {
		g.written = map[string]*syntax.Label{}
		look := func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Function, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
				return n == g.body
			case *syntax.Label:
				g.written[n.Name.Value] = n
			}
			return true
		}
		if g.body == nil {
			syntax.InspectFile(g.env.file, look)
		} else {
			syntax.Inspect(g.body, look)
		}
	}
	target := g.written[label]
	for i, f := range finals {
		if target == nil || !flowContains(f.try, target) {
			return finals[i:]
		}
	}
	return nil
}

// through returns where a jump to at starts when it runs finals first: the
// innermost block runs first, and each continues into the next one out.
func (g *graph) through(finals []final, at int) int {
	for _, f := range finals {
		at = g.stmt(f.body, at, f.out)
	}
	return at
}

// flowContains reports whether inner is written inside outer.
func flowContains(outer, inner syntax.Node) bool {
	return inner.Span().Start >= outer.Span().Start && inner.Span().End <= outer.Span().End
}

// repeats returns the successors of a loop test: the body and, unless the
// condition is absent or the constant true, what follows the loop. A loop
// that always repeats is left through `break`, `return` or a throw only.
func repeats(cond syntax.Node, body, next int) []int {
	if cond == nil {
		return []int{body}
	}
	if test, ok := cond.(syntax.Expr); ok && constantName(test) == "true" {
		return []int{body}
	}
	return []int{body, next}
}

// jumpsLevels reports a `break` or `continue` written with a level inside a
// try statement, outside the functions declared in it.
func jumpsLevels(n *syntax.Try) bool {
	found := false
	syntax.Inspect(n, func(child syntax.Node) bool {
		switch jump := child.(type) {
		case *syntax.Function, *syntax.ClassLike, *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.Break:
			found = found || jump.Num != nil
		case *syntax.Continue:
			found = found || jump.Num != nil
		}
		return !found
	})
	return found
}

// inside returns the exits of a loop or switch body.
func (x exits) inside(brk, cont int) exits {
	x.brk, x.cont = brk, cont
	x.levels = append(x.levels[:len(x.levels):len(x.levels)], [2]int{brk, cont})
	return x
}

// leaveLevels returns the target of `break` (which 0) or `continue` (which 1),
// over the number of levels a literal operand names.
func (g *graph) leaveLevels(operand syntax.Expr, x exits, which int) int {
	levels := int64(0)
	if literal, ok := syntax.UnwrapParens(operand).(*syntax.Literal); ok {
		levels, _ = strconv.ParseInt(literal.Raw, 10, 64)
	}
	switch {
	case operand == nil || levels == 1:
		return [2]int{x.brk, x.cont}[which]
	case levels < 1 || levels > int64(len(x.levels)):
		g.complete = false
		return [2]int{x.brk, x.cont}[which]
	}
	return x.levels[len(x.levels)-int(levels)][which]
}

// leave builds the block of an expression that never completes: a throw, an
// exit or a call that does not return.
func (g *graph) leave(x syntax.Expr, out exits) (int, bool) {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Exit:
		return g.atom(n, 0, out), true
	case *syntax.Throw:
		// A throw is an expression since PHP 8.0 only.
		if g.env.types.PHP.AtLeast(phpversion.PHP80) {
			return g.atom(n, out.thrown, out), true
		}
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		if g.env.never(n) {
			return g.atom(n, out.thrown, out), true
		}
	}
	return 0, false
}
