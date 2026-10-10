// Package nocontentresponsewithbody implements the native NoContentResponseWithBody inspection.
package nocontentresponsewithbody

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Omit response content for status 204."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NoContentResponseWithBody" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckNoContentResponseWithBody(ctx, n, message)
}
