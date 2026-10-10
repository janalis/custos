// Package nancheckedwithequality implements the native NanCheckedWithEquality inspection.
package nancheckedwithequality

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use is_nan to test for NaN."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NanCheckedWithEquality" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsIdentical, syntax.TIsNotEqual, syntax.TIsNotIdentical:
	default:
		return
	}
	other := b.Left
	if !nan(ctx, b.Right) {
		if !nan(ctx, b.Left) {
			return
		}
		other = b.Right
	}
	var fixes []diagnostic.Fix
	if ctx.TypeOf(other).OnlyOf("int", "float") && !strings.Contains(ctx.Text(b), "/*") && !strings.Contains(ctx.Text(b), "//") && !strings.Contains(ctx.Text(b), "#") {
		text := `\is_nan(` + ctx.Text(other) + `)`
		if b.Op.Kind == syntax.TIsNotEqual || b.Op.Kind == syntax.TIsNotIdentical {
			text = "!" + text
		}
		fixes = append(fixes, diagnostic.Fix{Title: "Check for NaN", Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: b.Span(), NewText: text}} }})
	}
	ctx.ReportNode(b, message, fixes...)
}

func nan(ctx *analysis.Context, e syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	return ok && semanticquery.GlobalConstName(ctx, c) == "NAN"
}
