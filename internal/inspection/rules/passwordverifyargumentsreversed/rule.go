// Package passwordverifyargumentsreversed implements PasswordVerifyArgumentsReversed.
package passwordverifyargumentsreversed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Pass the password before its stored hash."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PasswordVerifyArgumentsReversed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPasswordVerifyArgumentsReversed(ctx, n, message)
}
