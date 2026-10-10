// Package opensslsigningfailureignored implements OpenSslSigningFailureIgnored.
package opensslsigningfailureignored

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check signing success before publishing the signature."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OpenSslSigningFailureIgnored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckOpenSslSigningFailureIgnored(ctx, n, message)
}
