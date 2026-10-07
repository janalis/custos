package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// issetConstructsCanBeMerged reports `isset(a) && isset(b)` and
// `!isset(a) || !isset(b)` chains that one isset() call can express.
type issetConstructsCanBeMerged struct{}

func init() { register(issetConstructsCanBeMerged{}) }

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
	if p, _ := util.ParentSkipParens(root); p != nil {
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
	if len(frags) < 2 {
		return
	}
	// D3 / D4
	first, second := -1, -1
	var hits [2]*syntax.Isset
	for i, f := range frags {
		is := issetFragment(f, op == syntax.TBooleanOr)
		if is == nil {
			// D7: the second check cannot move across a side effect.
			if first >= 0 && util.MayHaveSideEffects(f) {
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
	span := root.Span()
	ctx.Report(hits[1].Span(), msg, analysis.Fix{
		Title: "Merge isset() calls",
		Edits: func() []analysis.TextEdit {
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
					b.WriteString(sep)
				}
				if i == first { // F1: the merged check takes the first hit's place
					b.WriteString(merged.String())
				} else if p, ok := f.Parent().(*syntax.Paren); ok {
					b.WriteString(ctx.Text(p))
				} else {
					b.WriteString(ctx.Text(f))
				}
			}
			return []analysis.TextEdit{{Span: span, NewText: b.String()}}
		},
	})
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
