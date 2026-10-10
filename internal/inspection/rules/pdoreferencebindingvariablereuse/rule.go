// Package pdoreferencebindingvariablereuse implements the native PdoReferenceBindingVariableReuse inspection.
package pdoreferencebindingvariablereuse

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Bind distinct parameters to distinct scalar identities."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoReferenceBindingVariableReuse" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoReferenceBindingVariableReuse(ctx, n, message)
}
