// Package pdoidentifierplaceholder implements the native PdoIdentifierPlaceholder inspection.
package pdoidentifierplaceholder

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Select SQL identifiers from a trusted allowlist."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoIdentifierPlaceholder" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoIdentifierPlaceholder(ctx, n, message)
}
