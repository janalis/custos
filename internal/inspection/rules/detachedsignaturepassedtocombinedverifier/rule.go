// Package detachedsignaturepassedtocombinedverifier implements DetachedSignaturePassedToCombinedVerifier.
package detachedsignaturepassedtocombinedverifier

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Verify detached signatures with the detached API."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DetachedSignaturePassedToCombinedVerifier" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckDetachedSignaturePassedToCombinedVerifier(ctx, n, message)
}
