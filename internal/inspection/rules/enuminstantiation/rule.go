// Package enuminstantiation implements the native EnumInstantiation inspection.
package enuminstantiation

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use an enum case instead of instantiation."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "EnumInstantiation" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	v := n.(*syntax.New)
	c := semanticquery.NativeNewClass(ctx, v)
	if c != nil && c.Kind == syntax.KindEnum {
		ctx.ReportNode(v, message)
	}
}
