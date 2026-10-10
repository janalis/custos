// Package pdoexeczerorejectedasfailure implements PdoExecZeroRejectedAsFailure.
package pdoexeczerorejectedasfailure

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare database execution failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoExecZeroRejectedAsFailure" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoExecZeroRejectedAsFailure(ctx, n, message)
}
