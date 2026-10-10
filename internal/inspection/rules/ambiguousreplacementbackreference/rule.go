// Package ambiguousreplacementbackreference implements the AmbiguousReplacementBackreference inspection.
package ambiguousreplacementbackreference

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AmbiguousReplacementBackreference" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Brace the replacement backreference before a literal digit."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "preg_replace")
	if call == nil {
		return
	}
	pattern, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "pattern"))
	if !ok {
		return
	}
	count, known := semanticquery.NativeRegexCaptures(pattern)
	if !known {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 1, "replacement")
	replacement, ok := semanticquery.NativeString(ctx, arg)
	if !ok {
		return
	}
	replacementOut := replacement
	found := false
	for i := 0; i+2 < len(replacement); i++ {
		if replacement[i] == '\\' {
			i++
			continue
		}
		if replacement[i] != '$' || replacement[i+1] < '1' || replacement[i+1] > '9' || replacement[i+2] < '0' || replacement[i+2] > '9' {
			continue
		}
		shorter := int(replacement[i+1] - '0')
		larger := shorter*10 + int(replacement[i+2]-'0')
		if larger <= count || shorter > count {
			continue
		}
		replacementOut = replacement[:i] + "${" + string(replacement[i+1]) + "}" + replacement[i+2:]
		found = true
		break
	}
	if !found {
		return
	}
	literal, ok := arg.(*syntax.Literal)
	if ok && strings.HasPrefix(literal.Raw, "'") && !strings.ContainsAny(replacementOut, "'\\") {
		ctx.ReportNode(call, message, astquery.ReplaceFix(arg.Span(), "'"+replacementOut+"'"))
		return
	}
	ctx.ReportNode(call, message)
}

func (rule) Semantic() {}
