package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// choose builds a switch: the subject runs once, the case expressions are
// compared with it in order until one matches, and a matching case runs into
// the bodies below it until a break. `continue` leaves the switch like
// `break`. Without a matching case the default body runs, wherever it is
// written, or nothing does. A comparison is loose: the only facts its outcome
// establishes about the subject are those of pin and differs.
func (g *graph) choose(n *syntax.Switch, next int, x exits) int {
	g.blocks[next].shared = true
	inner := x.inside(next, next)
	bodies := make([]int, len(n.Cases))
	below, unmatched := next, next
	for i := len(n.Cases) - 1; i >= 0; i-- {
		below = g.list(n.Cases[i].Stmts, below, inner)
		g.blocks[below].shared = true
		bodies[i] = below
		if n.Cases[i].Cond == nil {
			unmatched = below
		}
	}
	// `switch (true)` tests each case expression for truth, like an
	// if/elseif chain: a match establishes the condition, a miss its
	// negation for the cases below and for the default.
	// `switch (false)` runs the first case whose expression is not true.
	conditions := constantName(n.Cond)
	// A plain variable compared with constant labels only is the value the
	// labels were compared with: no label expression can have replaced it.
	// An element of a request table read under a constant key is compared
	// the same way; it has no binding, so the match is recorded for it.
	pinned := plainVariable(n.Cond)
	matching := requestElement(n.Cond)
	for _, c := range n.Cases {
		switch syntax.UnwrapParens(c.Cond).(type) {
		case nil, *syntax.Literal, *syntax.ConstFetch, *syntax.ClassConstFetch:
		default:
			pinned, matching = "", ""
		}
	}
	test := unmatched
	for i := len(n.Cases) - 1; i >= 0; i-- {
		label := n.Cases[i].Cond
		switch {
		case label == nil:
		case conditions == "true":
			test = g.test(label, x, bodies[i], test)
			g.blocks[test].shared = true
		case conditions == "false":
			test = g.test(label, x, test, bodies[i])
			g.blocks[test].shared = true
		default:
			matched, missed := bodies[i], test
			if pinned+matching != "" && allowed(label) {
				matched = g.add(nil, bodies[i])
				g.blocks[matched].pin = pinned + matching
				g.blocks[matched].shared = true
			}
			switch constantName(label) {
			case "false", "null":
				// Every falsy value equals both: the cases below and the
				// default see a subject that is neither.
				if pinned != "" {
					missed = g.add(nil, test)
					g.blocks[missed].differs = n.Cond
					g.blocks[missed].shared = true
				}
			case "true":
				if pinned != "" {
					matched = g.add(nil, bodies[i])
					g.blocks[matched].differs = n.Cond
					g.blocks[matched].shared = true
				}
			}
			test = g.add(nil, matched, missed)
			g.blocks[test].exceptional = x.thrown
			g.blocks[test].nodes = []syntax.Node{label}
			g.blocks[test].shared = true
		}
	}
	return g.atom(n.Cond, test, x)
}

// allowed reports a case label that request text can only match by being it:
// a string literal that is not empty and does not start, after any leading
// whitespace, like a number. PHP compares a string with a non-numeric string
// byte for byte in every version, and an array never equals a string. Any
// other label lets text through a loose comparison: an integer, a float or a
// numeric string equals numeric strings with surrounding whitespace, signs
// and exponents (and, before PHP 8, any text starting with the number, or
// any non-numeric text for 0); `”`, `null`, `false` and `true` equal many
// strings. The only other values that equal an allowed label are the boolean
// true and, before PHP 8, the integer 0: neither holds text.
func allowed(label syntax.Expr) bool {
	literal, ok := syntax.UnwrapParens(label).(*syntax.Literal)
	if !ok || literal.LitKind != syntax.LitString {
		return false
	}
	text, known := literalText(literal)
	text = strings.TrimLeft(text, " \t\n\r\v\f")
	return known && text != "" && !strings.ContainsRune("0123456789+-.", rune(text[0]))
}
