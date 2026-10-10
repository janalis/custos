// Package sqliteforeignkeysettinginsidetransaction implements SqliteForeignKeySettingInsideTransaction.
package sqliteforeignkeysettinginsidetransaction

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Change foreign-key enforcement before the transaction."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteForeignKeySettingInsideTransaction" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSqliteForeignKeySettingInsideTransaction(ctx, n, message)
}
