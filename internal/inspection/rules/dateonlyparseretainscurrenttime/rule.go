// Package dateonlyparseretainscurrenttime implements the DateOnlyParseRetainsCurrentTime inspection.
package dateonlyparseretainscurrenttime

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateOnlyParseRetainsCurrentTime" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KStaticCall} }

const message = "Reset omitted time fields when parsing a date-only value."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := semanticquery.NativeDateStatic(ctx, n, "createFromFormat")
	if call == nil {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 0, "format")
	format, ok := semanticquery.NativeString(ctx, arg)
	if !ok {
		return
	}
	tokens := semanticquery.NativeFormatTokens(format)
	if tokens['!'] || tokens['|'] {
		return
	}
	if !tokens['Y'] && !tokens['y'] && !tokens['m'] && !tokens['n'] && !tokens['d'] && !tokens['j'] && !tokens['z'] {
		return
	}
	for _, c := range []byte("HhGgisuvU") {
		if tokens[c] {
			return
		}
	}
	literal, ok := arg.(*syntax.Literal)
	if ok && literal.LitKind == syntax.LitString && (strings.HasPrefix(literal.Raw, "'") || strings.HasPrefix(literal.Raw, `"`)) {
		replacement := literal.Raw[:1] + "!" + literal.Raw[1:]
		ctx.ReportNode(call, message, astquery.ReplaceFix(arg.Span(), replacement))
		return
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}
