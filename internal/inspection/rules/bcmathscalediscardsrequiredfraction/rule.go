// Package bcmathscalediscardsrequiredfraction implements the native BcMathScaleDiscardsRequiredFraction inspection.
package bcmathscalediscardsrequiredfraction

import (
	"math/big"
	"regexp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Retain fractional digits in decimal arithmetic."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "BcMathScaleDiscardsRequiredFraction" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "bcadd" && name != "bcsub" && name != "bcmul" {
		return
	}
	scale, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 2, "scale"))
	if !ok || scale < 0 || scale > 1024 {
		return
	}
	operands := semanticquery.ExpansionBcOperands(ctx, c)
	a, ok := decimal(ctx, operands[0])
	if !ok {
		return
	}
	b, ok := decimal(ctx, operands[1])
	if !ok {
		return
	}
	result := new(big.Rat)
	switch name {
	case "bcadd":
		result.Add(a, b)
	case "bcsub":
		result.Sub(a, b)
	case "bcmul":
		result.Mul(a, b)
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(scale), nil)
	if !result.Mul(result, new(big.Rat).SetInt(power)).IsInt() {
		ctx.ReportNode(c, message)
	}
}

var decimalSyntax = regexp.MustCompile(`^[+-]?(?:\d+(?:\.\d*)?|\.\d+)$`)

func decimal(ctx *analysis.Context, e syntax.Expr) (*big.Rat, bool) {
	s, ok := semanticquery.NativeString(ctx, e)
	if !ok || len(s) > 1024 || !decimalSyntax.MatchString(s) {
		return nil, false
	}
	n, ok := new(big.Rat).SetString(s)
	return n, ok
}
