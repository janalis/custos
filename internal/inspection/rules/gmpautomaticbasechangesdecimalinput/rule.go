// Package gmpautomaticbasechangesdecimalinput implements the native GmpAutomaticBaseChangesDecimalInput inspection.
package gmpautomaticbasechangesdecimalinput

import (
	"math/big"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Specify decimal base for zero-prefixed GMP input."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GmpAutomaticBaseChangesDecimalInput" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "gmp_init") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "num")
	s, ok := semanticquery.NativeString(ctx, arg)
	if !ok || len(s) < 2 || len(s) > 1024 || s[0] != '0' || strings.Trim(s, "01234567") != "" {
		return
	}
	octal, _ := new(big.Int).SetString(s, 8)
	decimal, _ := new(big.Int).SetString(s, 10)
	if octal.Cmp(decimal) == 0 {
		return
	}
	base := semanticquery.CallArgument(c.Args, 1, "base")
	if base != nil {
		v, known := semanticquery.NativeInt(ctx, base)
		if !known || v != 0 {
			return
		}
	}
	var fixes []diagnostic.Fix
	if base != nil {
		if literal, ok := syntax.UnwrapParens(base).(*syntax.Literal); ok {
			fixes = append(fixes, astquery.ReplaceFix(literal.Span(), "10"))
		}
	} else if c.Args != nil && len(c.Args.Args) == 1 {
		a, ok := c.Args.Args[0].(*syntax.Arg)
		if ok && !a.Unpack && !strings.ContainsAny(ctx.SpanText(syntax.Span{Start: a.Span().End, End: c.Args.Span().End}), "/#") {
			suffix := ", 10"
			if a.Name != nil {
				suffix = ", base: 10"
			}
			at := a.Span().End
			fixes = append(fixes, astquery.ReplaceFix(syntax.Span{Start: at, End: at}, suffix))
		}
	}
	ctx.ReportNode(c, message, fixes...)
}
