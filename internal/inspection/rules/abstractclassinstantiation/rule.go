// Package abstractclassinstantiation implements the native AbstractClassInstantiation inspection.
package abstractclassinstantiation

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Instantiate a concrete class."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AbstractClassInstantiation" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	v := n.(*syntax.New)
	c := semanticquery.NativeNewClass(ctx, v)
	if c != nil && c.Abstract {
		ctx.ReportNode(v, message)
	}
}
