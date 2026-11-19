// Package traitinstantiation implements the native TraitInstantiation inspection.
package traitinstantiation

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Instantiate a class using the trait."

type rule struct{}

// New returns the TraitInstantiation inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TraitInstantiation" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	v := n.(*syntax.New)
	c := semanticquery.NativeNewClass(ctx, v)
	if c != nil && c.Kind == syntax.KindTrait {
		ctx.ReportNode(v, message)
	}
}
