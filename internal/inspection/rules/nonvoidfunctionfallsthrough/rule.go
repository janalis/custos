// Package nonvoidfunctionfallsthrough implements NonVoidFunctionFallsThrough.
package nonvoidfunctionfallsthrough

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return a value on every reachable path."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "NonVoidFunctionFallsThrough" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckNonVoidFunctionFallsThrough(ctx, n, message)
}
