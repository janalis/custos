// Package documentedoperandtype implements the native
// DocumentedOperandType inspection.
package documentedoperandtype

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "DocumentedOperandType" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary, syntax.KAssign} }

func (rule) Semantic() {}

// Check reports the operands of an arithmetic or bitwise operation when
// PHPDoc-aware types show they make it throw and native declarations alone
// do not (those are InvalidOperandType's).
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !ctx.PHP.AtLeast(phpversion.PHP80) {
		return
	}
	for _, m := range semanticquery.InvalidOperands(ctx, n, true) {
		ctx.ReportNode(m.Operand, m.Message())
	}
}
