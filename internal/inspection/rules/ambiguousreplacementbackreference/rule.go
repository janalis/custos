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
	replacementOut := braceReferences(replacement, count, 0)
	if replacementOut == replacement {
		return
	}
	content, quote, literal := astquery.QuotedStringRaw(arg)
	if literal {
		raw := ctx.Text(arg)
		body := braceReferences(content, count, quote)
		out := raw[:len(raw)-len(content)-1] + body + raw[len(raw)-1:]
		// Raw and decoded escapes need not expose the same tokens (for
		// example \x24 encodes a dollar). Only rewrite when all references
		// were changed and every other decoded byte stays identical.
		if decoded, ok := astquery.StringLiteralValue(out); ok && decoded == replacementOut {
			ctx.ReportNode(call, message, astquery.ReplaceFix(arg.Span(), out))
			return
		}
	}
	ctx.ReportNode(call, message)
}

// braceReferences preserves all surrounding bytes. On double-quoted PHP
// source the dollar must be escaped after adding braces to avoid interpolation.
func braceReferences(text string, count int, quote byte) string {
	var out strings.Builder
	start := 0
	for i := 0; i+2 < len(text); i++ {
		if text[i] != '$' || text[i+1] < '1' || text[i+1] > '9' || text[i+2] < '0' || text[i+2] > '9' {
			continue
		}
		slashes := 0
		for j := i - 1; j >= 0 && text[j] == '\\'; j-- {
			slashes++
		}
		decodedSlashes := slashes
		switch quote {
		case '\'':
			decodedSlashes = (slashes + 1) / 2
		case '"':
			decodedSlashes = slashes / 2
		}
		if decodedSlashes%2 != 0 {
			continue
		}
		shorter := int(text[i+1] - '0')
		larger := shorter*10 + int(text[i+2]-'0')
		if larger <= count || shorter > count {
			continue
		}
		out.WriteString(text[start:i])
		if quote == '"' && slashes%2 == 0 {
			out.WriteByte('\\')
		}
		out.WriteString("${")
		out.WriteByte(text[i+1])
		out.WriteByte('}')
		start = i + 2
		i++
	}
	if start == 0 {
		return text
	}
	out.WriteString(text[start:])
	return out.String()
}

func (rule) Semantic() {}
