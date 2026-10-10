// Package pdogroupedfetchreadsremovedcolumn implements PdoGroupedFetchReadsRemovedColumn.
package pdogroupedfetchreadsremovedcolumn

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the group column from the group key."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoGroupedFetchReadsRemovedColumn" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoGroupedFetchReadsRemovedColumn(ctx, n, message)
}
