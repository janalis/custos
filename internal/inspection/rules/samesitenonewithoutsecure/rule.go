// Package samesitenonewithoutsecure implements the native SameSiteNoneWithoutSecure inspection.
package samesitenonewithoutsecure

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a Secure cookie with SameSite None."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SameSiteNoneWithoutSecure" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSameSiteNoneWithoutSecure(ctx, n, message)
}
