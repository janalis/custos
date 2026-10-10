// Package sodiumsecretboxnoncereused implements SodiumSecretboxNonceReused.
package sodiumsecretboxnoncereused

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a fresh nonce for each distinct message."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumSecretboxNonceReused" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSodiumSecretboxNonceReused(ctx, n, message)
}
