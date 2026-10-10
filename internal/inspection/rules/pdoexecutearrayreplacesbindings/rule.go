// Package pdoexecutearrayreplacesbindings implements the native PdoExecuteArrayReplacesBindings inspection.
package pdoexecutearrayreplacesbindings

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Omit the empty parameter array to retain existing bindings."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoExecuteArrayReplacesBindings" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoExecuteArrayReplacesBindings(ctx, n, message)
}
