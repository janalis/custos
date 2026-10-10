// Package sodiumsecretboxkeylengthmismatch implements SodiumSecretboxKeyLengthMismatch.
package sodiumsecretboxkeylengthmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a correctly sized secretbox key."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumSecretboxKeyLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSodiumSecretboxKeyLengthMismatch(ctx, n, message)
}
