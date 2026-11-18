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

// New returns the GmpAutomaticBaseChangesDecimalInput inspection.
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
	if octal.Cmp(decimal) == 0 || !semanticquery.StableLocalValue(ctx, arg, semanticquery.LocalStability{}) {
		return
	}
	base := semanticquery.CallArgument(c.Args, 1, "base")
	if base != nil {
		v, known := semanticquery.NativeInt(ctx, base)
		if !known || v != 0 {
			return
		}
	}
	ctx.ReportNode(c, message, fixes(ctx, c, arg, base)...)
}

// fixes builds the decimal-base edit. It is withheld when the digits come from
// local flow or a late-bound constant: adding a decimal base to a value that
// turns out to be prefixed would throw.
func fixes(ctx *analysis.Context, c *syntax.FuncCall, arg, base syntax.Expr) []diagnostic.Fix {
	if !proven(arg) {
		return nil
	}
	if base != nil {
		if literal, ok := syntax.UnwrapParens(base).(*syntax.Literal); ok {
			return []diagnostic.Fix{astquery.ReplaceFix(literal.Span(), "10")}
		}
		return nil
	}
	if len(c.Args.Args) != 1 {
		return nil
	}
	a, ok := c.Args.Args[0].(*syntax.Arg)
	if !ok || a.Unpack || strings.ContainsAny(ctx.SpanText(syntax.Span{Start: a.Span().End, End: c.Args.Span().End}), "/#") {
		return nil
	}
	suffix := ", 10"
	if a.Name != nil {
		suffix = ", base: 10"
	}
	at := a.Span().End
	return []diagnostic.Fix{astquery.ReplaceFix(syntax.Span{Start: at, End: at}, suffix)}
}

// proven reports whether the digits are fixed by the argument expression
// itself: no variable and no `static::` constant, which a subclass may
// redefine.
func proven(arg syntax.Expr) bool {
	ok := true
	syntax.Inspect(arg, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Variable:
			ok = false
		case *syntax.ClassConstFetch:
			if class, isName := n.Class.(*syntax.Name); isName && strings.EqualFold(class.Value, "static") {
				ok = false
			}
		}
		return ok
	})
	return ok
}
