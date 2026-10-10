// Package regexreplacementreferencesmissinggroup implements the native RegexReplacementReferencesMissingGroup inspection.
package regexreplacementreferencesmissinggroup

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reference a capture present in the pattern."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexReplacementReferencesMissingGroup" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "preg_replace") {
		return
	}
	p, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "pattern"))
	replacement, rk := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "replacement"))
	if !k || !rk {
		return
	}
	count, known := semanticquery.NativeRegexCaptures(p)
	if known && missing(replacement, count) {
		ctx.ReportNode(c, message)
	}
}

func missing(s string, count int) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			if i+1 >= len(s) {
				continue
			}
			if s[i+1] == '\\' || s[i+1] == '$' {
				i++
				continue
			}
		}
		if s[i] != '$' && s[i] != '\\' {
			continue
		}
		j := i + 1
		braced := j < len(s) && s[j] == '{'
		if braced {
			j++
		}
		start := j
		number := 0
		for j < len(s) && s[j] >= '0' && s[j] <= '9' && j-start < 2 {
			number = number*10 + int(s[j]-'0')
			j++
		}
		if j == start || braced && (j >= len(s) || s[j] != '}') {
			continue
		}
		if number > count {
			return true
		}
		i = j - 1
	}
	return false
}
