// Package transactionexceptionwithoutrollback implements the native TransactionExceptionWithoutRollback inspection.
package transactionexceptionwithoutrollback

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Roll back the active transaction on the caught failure path."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TransactionExceptionWithoutRollback" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KReturn} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckTransactionExceptionWithoutRollback(ctx, n, message)
}
