// Package regexbodydoesnotcompile implements the native RegexBodyDoesNotCompile inspection.
package regexbodydoesnotcompile

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Correct the regular expression body."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexBodyDoesNotCompile" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "preg_match" && name != "preg_match_all" && name != "preg_replace" && name != "preg_split" && name != "preg_replace_callback" {
		return
	}
	pattern, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "pattern"))
	if k && malformed(pattern) {
		ctx.ReportNode(c, message)
	}
}

// malformed recognizes only ordinary structural failures in a small PCRE subset.
func malformed(p string) bool {
	if len(p) < 2 || p[0] != '/' && p[0] != '~' {
		return false
	}
	end := strings.LastIndexByte(p, p[0])
	if end == 0 {
		return false
	}
	if strings.Contains(p[1:end], "(?") || strings.Contains(p[1:end], "(*") {
		return false
	}
	for _, m := range p[end+1:] {
		if !strings.ContainsRune("imsADSUJur", m) {
			return false
		}
	}
	depth := 0
	class := false
	for i := 1; i < end; i++ {
		c := p[i]
		if c == '\\' {
			if i+1 >= end || strings.ContainsRune("QEc", rune(p[i+1])) {
				return false
			}
			i++
			continue
		}
		if c == p[0] {
			return false
		}
		if class {
			if c == '[' {
				return false
			}
			if c == ']' {
				class = false
			}
			continue
		}
		if c == '[' {
			if i+1 < end && (p[i+1] == ']' || p[i+1] == '^' && i+2 < end && p[i+2] == ']') {
				return false
			}
			class = true
			continue
		}
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
			if depth < 0 {
				return true
			}
		}
	}
	return class || depth != 0
}
