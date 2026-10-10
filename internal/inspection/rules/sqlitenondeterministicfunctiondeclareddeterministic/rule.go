// Package sqlitenondeterministicfunctiondeclareddeterministic implements SqliteNondeterministicFunctionDeclaredDeterministic.
package sqlitenondeterministicfunctiondeclareddeterministic

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Remove the deterministic flag from this callback."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteNondeterministicFunctionDeclaredDeterministic" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSqliteNondeterministicFunctionDeclaredDeterministic(ctx, n, message)
}
