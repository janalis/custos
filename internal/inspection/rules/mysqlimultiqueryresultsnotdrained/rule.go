// Package mysqlimultiqueryresultsnotdrained implements MysqliMultiQueryResultsNotDrained.
package mysqlimultiqueryresultsnotdrained

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Drain all multi-query results before another query."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliMultiQueryResultsNotDrained" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckMysqliMultiQueryResultsNotDrained(ctx, n, message)
}
