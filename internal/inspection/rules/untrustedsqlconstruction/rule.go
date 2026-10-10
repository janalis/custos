// Package untrustedsqlconstruction implements the native UntrustedSqlConstruction inspection.
package untrustedsqlconstruction

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Bind untrusted values as SQL parameters."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedSqlConstruction" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall, syntax.KFuncCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckUntrustedSQLConstruction(ctx, n, message)
}
