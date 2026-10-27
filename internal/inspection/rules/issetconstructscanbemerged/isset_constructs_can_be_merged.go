package issetconstructscanbemerged

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// issetConstructsCanBeMerged reports `isset(a) && isset(b)` and
// `!isset(a) || !isset(b)` chains that one isset() call can express.
type issetConstructsCanBeMerged struct{}

func (issetConstructsCanBeMerged) ID() string { return "IssetConstructsCanBeMerged" }
func (issetConstructsCanBeMerged) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary}
}

func (issetConstructsCanBeMerged) Check(ctx *analysis.Context, n syntax.Node) {
	root := n.(*syntax.Binary)
	op := root.Op.Kind
	if op != syntax.TBooleanAnd && op != syntax.TBooleanOr {
		return
	}
	// D1 (looking through all parentheses, see spec Divergences).
	if p, _ := astquery.ParentSkipParens(root); p != nil {
		if pb, ok := p.(*syntax.Binary); ok && pb.Op.Kind == op {
			return
		}
	}
	// D2
	var frags []syntax.Expr
	var flatten func(e syntax.Expr)
	flatten = func(e syntax.Expr) {
		e = syntax.UnwrapParens(e)
		if b, ok := e.(*syntax.Binary); ok && b.Op.Kind == op {
			flatten(b.Left)
			flatten(b.Right)
			return
		}
		if e != nil {
			frags = append(frags, e)
		}
	}
	flatten(root.Left)
	flatten(root.Right)
	// D3 / D4
	first, second := -1, -1
	var hits [2]*syntax.Isset
	for i, f := range frags {
		is := issetFragment(f, op == syntax.TBooleanOr)
		if is == nil {
			// D7: the second check cannot move across a side effect.
			if first >= 0 && flowquery.MayHaveSideEffects(f) {
				first = -1
			}
			continue
		}
		if first < 0 {
			first, hits[0] = i, is
			continue
		}
		second, hits[1] = i, is
		break
	}
	if second < 0 {
		return
	}
	prefix, sep, msg := "isset(", " && ", "Merge this check into the preceding isset() call."
	if op == syntax.TBooleanOr {
		prefix, sep, msg = "!isset(", " || ", "Merge this check into the preceding !isset() call."
	}
	// F2: fragments between the hits short-circuit the second check, so it
	// only moves in front of them when evaluating it early is unobservable.
	span := root.Span()
	// F3 (custos): no fix would keep the comments between the checks.
	if second != first+1 && !plainVariables(hits[1]) || astquery.HasComment(ctx.File, span) {
		ctx.Report(hits[1].Span(), msg)
		return
	}
	ctx.Report(hits[1].Span(), msg, diagnostic.Fix{
		Title: "Merge isset() calls",
		Edits: func() []diagnostic.TextEdit {
			var merged strings.Builder
			merged.WriteString(prefix)
			k := 0
			for _, is := range hits {
				for _, v := range is.Vars {
					if k > 0 {
						merged.WriteString(", ")
					}
					merged.WriteString(ctx.Text(v))
					k++
				}
			}
			merged.WriteByte(')')
			var b strings.Builder
			for i, f := range frags {
				if i == second {
					continue
				}
				if b.Len() > 0 {
					b.WriteString(lineBreakSep(ctx.File, outer(frags[i-1]).Span().End, outer(f).Span().Start, sep))
				}
				if i == first { // F1: the merged check takes the first hit's place
					b.WriteString(merged.String())
				} else {
					b.WriteString(ctx.Text(outer(f)))
				}
			}
			return []diagnostic.TextEdit{{Span: span, NewText: b.String()}}
		},
	})
}

// outer returns the fragment with the parentheses written around it.
func outer(f syntax.Expr) syntax.Node {
	if p, ok := f.Parent().(*syntax.Paren); ok {
		return p
	}
	return f
}

// lineBreakSep returns the separator written before a fragment: sep, or,
// when the source broke the line between the two fragments, the operator
// with the same line break and indentation, on the side it was written
// (F1a).
func lineBreakSep(f *syntax.File, from, to uint32, sep string) string {
	gap := string(f.Src[from:to])
	nl := strings.LastIndexByte(gap, '\n')
	if nl < 0 {
		return sep
	}
	indent := gap[nl+1:]
	indent = indent[:len(indent)-len(strings.TrimLeft(indent, " \t"))]
	op := strings.TrimSpace(sep)
	if strings.IndexAny(gap, "&|") < nl {
		return " " + op + "\n" + indent
	}
	return "\n" + indent + op + " "
}

// issetFragment returns the isset construct a chain fragment qualifies with:
// `isset(...)` in && chains, `!isset(...)` (no parentheses) in || chains.
func issetFragment(f syntax.Expr, negated bool) *syntax.Isset {
	if negated {
		u, ok := f.(*syntax.Unary)
		if !ok || u.Op.Kind != syntax.TExclaim {
			return nil
		}
		f = u.Expr
	}
	is, _ := f.(*syntax.Isset)
	if is != nil && len(is.Vars) == 0 {
		return nil
	}
	return is
}

// plainVariables reports whether every argument of is is a named variable:
// testing it cannot throw, call a magic method or evaluate an offset.
func plainVariables(is *syntax.Isset) bool {
	for _, v := range is.Vars {
		if va, ok := v.(*syntax.Variable); !ok || va.NameExpr != nil {
			return false
		}
	}
	return true
}
