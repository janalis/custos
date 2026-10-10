// Package pdofetchcolumnfalsyvalueloss implements the native PdoFetchColumnFalsyValueLoss inspection.
package pdofetchcolumnfalsyvalueloss

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare fetch failure strictly with false."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoFetchColumnFalsyValueLoss" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoFetchColumnFalsyValueLoss(ctx, n, message)
}
