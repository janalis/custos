// Package repeatedcookieheaderreplacement implements the native RepeatedCookieHeaderReplacement inspection.
package repeatedcookieheaderreplacement

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Append distinct cookie headers instead of replacing them."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RepeatedCookieHeaderReplacement" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckRepeatedCookieHeaderReplacement(ctx, n, message)
}
