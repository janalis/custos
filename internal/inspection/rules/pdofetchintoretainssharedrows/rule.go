// Package pdofetchintoretainssharedrows implements PdoFetchIntoRetainsSharedRows.
package pdofetchintoretainssharedrows

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clone fetched objects before retaining snapshots."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoFetchIntoRetainsSharedRows" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoFetchIntoRetainsSharedRows(ctx, n, message)
}
