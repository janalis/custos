// Package regexutf8subjectknowninvalid implements the native RegexUtf8SubjectKnownInvalid inspection.
package regexutf8subjectknowninvalid

import (
	"strings"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply valid UTF-8 to a Unicode pattern."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexUtf8SubjectKnownInvalid" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	position := 1
	switch name {
	case "preg_match", "preg_match_all", "preg_split":
	case "preg_replace":
		position = 2
	default:
		return
	}
	p, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "pattern"))
	if !k || len(p) < 2 {
		return
	}
	end := strings.LastIndexByte(p, p[0])
	if end < 1 || !strings.Contains(p[end+1:], "u") {
		return
	}
	subject, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, position, "subject"))
	if known && !utf8.ValidString(subject) {
		ctx.ReportNode(c, message)
	}
}
