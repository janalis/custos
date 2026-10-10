// Package pregquoteusedonreplacement implements the native PregQuoteUsedOnReplacement inspection.
package pregquoteusedonreplacement

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep pattern escaping out of literal replacement text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PregQuoteUsedOnReplacement" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "preg_replace") {
		return
	}
	quote, ok := semanticquery.CallArgument(c.Args, 1, "replacement").(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, quote, "preg_quote") {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(quote.Args, 0, "str"))
	if !known || strings.ContainsAny(text, "$\\") {
		return
	}
	if strings.ContainsAny(text, ".[](){}?*+^|#-=!<>:") {
		input := semanticquery.CallArgument(quote.Args, 0, "str")
		_, literal := input.(*syntax.Literal)
		delimiter := semanticquery.CallArgument(quote.Args, 1, "delimiter")
		_, delimiterLiteral := semanticquery.NativeString(ctx, delimiter)
		written := ctx.Text(quote)
		if literal && (delimiter == nil || delimiterLiteral) && !strings.Contains(written, "/*") && !strings.Contains(written, "//") && !strings.Contains(written, "#") && len(quote.Args.Args) <= 2 {
			ctx.ReportNode(quote, message, astquery.ReplaceFix(quote.Span(), ctx.Text(input)))
		} else {
			ctx.ReportNode(quote, message)
		}
	}
}
