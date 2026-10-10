// Package catchcannothandleknownthrowable implements CatchCannotHandleKnownThrowable.
package catchcannothandleknownthrowable

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Catch a type that handles the thrown value."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CatchCannotHandleKnownThrowable" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KCatch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCatchCannotHandleKnownThrowable(ctx, n, message)
}
