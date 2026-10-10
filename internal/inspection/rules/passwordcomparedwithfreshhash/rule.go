// Package passwordcomparedwithfreshhash implements the native PasswordComparedWithFreshHash inspection.
package passwordcomparedwithfreshhash

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Verify the password against the stored hash."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PasswordComparedWithFreshHash" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPasswordComparedWithFreshHash(ctx, n, message)
}
