package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// ternaryOperatorSimplify reports `BINARY ? true : false` (and the inverted
// form) and suggests the condition itself, its negation or a bool cast.
type ternaryOperatorSimplify struct{}

func init() { register(ternaryOperatorSimplify{}) }

func (ternaryOperatorSimplify) ID() string { return "TernaryOperatorSimplify" }

func (ternaryOperatorSimplify) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KTernary}
}

var oppositeComparison = map[syntax.TokenKind]string{
	syntax.TIsEqual:          "!=",
	syntax.TIsIdentical:      "!==",
	syntax.TIsNotEqual:       "==",
	syntax.TIsNotIdentical:   "===",
	syntax.TGreater:          "<=",
	syntax.TLess:             ">=",
	syntax.TIsGreaterOrEqual: "<",
	syntax.TIsSmallerOrEqual: ">",
}

func (ternaryOperatorSimplify) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Then == nil || t.Cond == nil || t.Else == nil || t.Span().Len() == 0 { // E3
		return
	}
	// D2
	thenV, ok1 := util.BoolConst(util.UnwrapParens(t.Then))
	_, ok2 := util.BoolConst(util.UnwrapParens(t.Else))
	if !ok1 || !ok2 {
		return
	}
	inverted := !thenV // D3
	// D1/D4
	var repl string
	switch b := util.UnwrapParens(t.Cond).(type) {
	case *syntax.Binary:
		text := ctx.Text(b)
		if opp, cmp := oppositeComparison[b.Op.Kind]; cmp {
			repl = text
			if inverted {
				if isOrdering(b.Op.Kind) && !(nonFloatScalar(ctx, b.Left) && nonFloatScalar(ctx, b.Right)) {
					// Ordering comparisons are not complementary for NAN
					// (and incomparable arrays/objects): negate instead.
					repl = "!(" + text + ")"
				} else {
					repl = ctx.Text(b.Left) + " " + opp + " " + ctx.Text(b.Right)
				}
			}
		} else if b.Op.Kind == syntax.TBooleanAnd || b.Op.Kind == syntax.TBooleanOr {
			repl = "(" + text + ")"
			if inverted {
				repl = "!" + repl
			}
		} else {
			repl = otherBinaryRepl(text, inverted)
		}
	case *syntax.Instanceof:
		repl = otherBinaryRepl(ctx.Text(b), inverted)
	default:
		return // E1
	}
	span := t.Span()
	ctx.Report(span, "Replace the ternary with '"+repl+"'.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

func otherBinaryRepl(text string, inverted bool) string {
	if inverted {
		return "!(" + text + ")"
	}
	return "(bool)(" + text + ")"
}

func isOrdering(k syntax.TokenKind) bool {
	switch k {
	case syntax.TGreater, syntax.TLess, syntax.TIsGreaterOrEqual, syntax.TIsSmallerOrEqual:
		return true
	}
	return false
}

// nonFloatScalar reports whether e is known to be an int, string, bool or
// null, for which `a < b` and `a >= b` are always complementary.
func nonFloatScalar(ctx *analysis.Context, e syntax.Expr) bool {
	return ctx.TypeOf(e).OnlyOf("int", "string", "bool", "true", "false", "null")
}
