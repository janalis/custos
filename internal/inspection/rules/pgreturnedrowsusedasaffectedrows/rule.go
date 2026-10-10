// Package pgreturnedrowsusedasaffectedrows implements PgReturnedRowsUsedAsAffectedRows.
package pgreturnedrowsusedasaffectedrows

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use affected-row count for this statement."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgReturnedRowsUsedAsAffectedRows" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPgReturnedRowsUsedAsAffectedRows(ctx, n, message)
}
