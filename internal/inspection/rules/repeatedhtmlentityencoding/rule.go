// Package repeatedhtmlentityencoding implements the RepeatedHtmlEntityEncoding inspection.
package repeatedhtmlentityencoding

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RepeatedHtmlEntityEncoding" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Encode HTML entities at one intended output boundary."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	outer, fn := semanticquery.GlobalCall(ctx, n, "htmlspecialchars", "htmlentities")
	if outer == nil {
		return
	}
	inner, _ := semanticquery.GlobalCall(ctx, semanticquery.CallArgument(outer.Args, 0, "string"), fn)
	if inner == nil {
		return
	}
	for i, name := range []string{"flags", "encoding"} {
		a, outerState := semanticquery.CallArgumentBinding(outer.Args, i+1, name)
		b, innerState := semanticquery.CallArgumentBinding(inner.Args, i+1, name)
		if outerState == semanticquery.ArgumentUnknown || innerState == semanticquery.ArgumentUnknown || (a == nil) != (b == nil) || a != nil && !astquery.Equivalent(ctx.File, a, b) {
			return
		}
	}
	if double := semanticquery.CallArgument(outer.Args, 3, "double_encode"); double != nil {
		value, ok := astquery.BoolConst(double)
		if !ok || !value {
			return
		}
	}
	ctx.ReportNode(outer, message)
}

func (rule) Semantic() {}
