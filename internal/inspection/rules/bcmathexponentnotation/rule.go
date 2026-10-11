// Package bcmathexponentnotation implements the native BcMathExponentNotation inspection.
package bcmathexponentnotation

import (
	"regexp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use ordinary decimal notation for BCMath operands."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "BcMathExponentNotation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	operands := semanticquery.ExpansionBcOperands(ctx, c)
	// The shared operand list omits the exponent slot, which BCMath parses
	// with the same well-formedness contract as the base.
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "bcpow", "bcpowmod":
		operands = append(operands, semanticquery.CallArgument(c.Args, 1, "exponent"))
	}
	for _, operand := range operands {
		if s, ok := semanticquery.NativeString(ctx, operand); ok && exponent.MatchString(s) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

var exponent = regexp.MustCompile(`^[+-]?(?:\d+(?:\.\d*)?|\.\d+)[eE][+-]?\d+$`)
