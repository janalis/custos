// Package sessioncookieoptionssetafterstart implements the native SessionCookieOptionsSetAfterStart inspection.
package sessioncookieoptionssetafterstart

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Configure session cookie parameters before starting the session."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionCookieOptionsSetAfterStart" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSessionCookieOptionsSetAfterStart(ctx, n, message)
}
