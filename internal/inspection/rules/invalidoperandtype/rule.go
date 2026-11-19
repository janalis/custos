// Package invalidoperandtype implements the native
// InvalidOperandType inspection.
package invalidoperandtype

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "InvalidOperandType" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary, syntax.KAssign} }

func (rule) Semantic() {}

// Check reports the operands of an arithmetic or bitwise operation when
// native declarations prove they make it throw.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !ctx.PHP.AtLeast(phpversion.PHP80) {
		return
	}
	for _, m := range semanticquery.InvalidOperands(ctx, n, false) {
		ctx.ReportNode(m.Operand, m.Message())
	}
}
