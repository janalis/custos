package unusedgotolabel

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// unusedGotoLabel reports goto labels no goto in the same function targets.
type unusedGotoLabel struct{}

func (unusedGotoLabel) ID() string               { return "UnusedGotoLabel" }
func (unusedGotoLabel) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KLabel} }
func (unusedGotoLabel) Check(ctx *analysis.Context, n syntax.Node) {
	l := n.(*syntax.Label)
	fn := syntax.EnclosingFuncLike(l) // D1
	if _, hook := fn.(*syntax.PropertyHook); hook && len(ctx.File.Errors) != 0 {
		return
	}
	if fn == nil {
		return
	}
	// A parsed label in a function-like belongs to its block body;
	// expression scopes cannot contain labels and recovered hooks were rejected.
	body := syntax.FuncLikeBody(fn) // D2
	used := gotoTargets(ctx, body)[l.Name.Value]
	if used {
		return
	}
	span := l.Span()
	f := ctx.File
	ctx.Report(span, "Label '"+l.Name.Value+"' is never targeted by a goto; remove it.", diagnostic.Fix{
		Title: "Remove the label",
		Edits: func() []diagnostic.TextEdit {
			if _, _, ok := astquery.StmtList(f, l); !ok {
				// F2: the label is the whole unbraced body of a control
				// structure; deleting it would hand that role to the
				// next statement, so keep an empty statement instead.
				return []diagnostic.TextEdit{{Span: span, NewText: ";"}}
			}
			return []diagnostic.TextEdit{{Span: astquery.WithLeadingWhitespace(f, span)}}
		},
	})
}

// gotoTargets returns the labels targeted by a goto in body (D3), computed
// once per function body and file (a walk per label was quadratic).
func gotoTargets(ctx *analysis.Context, body *syntax.Block) map[string]bool {
	cache := ctx.Memo("gotoTargets", func() any { return map[*syntax.Block]map[string]bool{} }).(map[*syntax.Block]map[string]bool)
	if m, ok := cache[body]; ok {
		return m
	}
	m := map[string]bool{}
	syntax.Inspect(body, func(x syntax.Node) bool {
		// A goto cannot leave its own function: nested function-likes and
		// classes are separate scopes (custos diverges).
		if _, ok := x.(*syntax.ClassLike); ok || (x != syntax.Node(body) && syntax.IsFuncLike(x)) {
			return false
		}
		if g, ok := x.(*syntax.Goto); ok && g.Label != nil {
			m[g.Label.Value] = true
		}
		return true
	})
	cache[body] = m
	return m
}
