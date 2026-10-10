// Package dateminutesmonthtokenconfusion implements the DateMinutesMonthTokenConfusion inspection.
package dateminutesmonthtokenconfusion

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateMinutesMonthTokenConfusion" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall} }

const message = "Use the minute token in this clock format."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	arg := semanticquery.NativeDateFormat(ctx, n)
	format, ok := semanticquery.NativeString(ctx, arg)
	if !ok {
		return
	}
	tokens := semanticquery.NativeFormatTokens(format)
	if tokens['Y'] || tokens['y'] || tokens['d'] || tokens['j'] {
		return
	}
	found := false
	for i := 0; i+2 < len(format); i++ {
		if format[i] == '\\' {
			i++
			continue
		}
		if strings.ContainsRune("HGhg", rune(format[i])) && format[i+1] == ':' && format[i+2] == 'm' {
			found = true
		}
	}
	if !found {
		return
	}
	literal, ok := arg.(*syntax.Literal)
	if ok && literal.LitKind == syntax.LitString && len(literal.Raw) >= 2 && (literal.Raw[0] == '\'' || literal.Raw[0] == '"') && literal.Raw[1:len(literal.Raw)-1] == format {
		raw := []byte(literal.Raw)
		for i := 1; i+2 < len(raw)-1; i++ {
			if raw[i] == '\\' {
				i++
				continue
			}
			if strings.ContainsRune("HGhg", rune(raw[i])) && raw[i+1] == ':' && raw[i+2] == 'm' {
				raw[i+2] = 'i'
			}
		}
		ctx.ReportNode(n, message, astquery.ReplaceFix(arg.Span(), string(raw)))
		return
	}
	ctx.ReportNode(n, message)
}

func (rule) Semantic() {}
