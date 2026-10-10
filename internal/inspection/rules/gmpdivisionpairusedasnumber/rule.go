// Package gmpdivisionpairusedasnumber implements the native GmpDivisionPairUsedAsNumber inspection.
package gmpdivisionpairusedasnumber

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Select a quotient or remainder before using the GMP result."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GmpDivisionPairUsedAsNumber" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	var positions []string
	switch name {
	case "gmp_strval", "gmp_abs", "gmp_neg", "gmp_sqrt", "gmp_intval":
		positions = []string{"num"}
	case "gmp_add", "gmp_sub", "gmp_mul", "gmp_cmp", "gmp_div_q", "gmp_div_r", "gmp_div_qr", "gmp_mod":
		positions = []string{"num1", "num2"}
	default:
		return
	}
	for i, argName := range positions {
		v := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, i, argName))
		producer, ok := v.(*syntax.FuncCall)
		if ok && semanticquery.NativeBuiltin(ctx, producer, "gmp_div_qr") {
			ctx.ReportNode(c, message)
			return
		}
	}
}
