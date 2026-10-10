// Package redirectcontinuesprotectedexecution implements the native RedirectContinuesProtectedExecution inspection.
package redirectcontinuesprotectedexecution

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Terminate the unauthorized branch after redirecting."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RedirectContinuesProtectedExecution" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckRedirectContinuesProtectedExecution(ctx, n, message)
}
