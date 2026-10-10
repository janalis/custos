// Package localizednumbertrailinginputaccepted implements the native LocalizedNumberTrailingInputAccepted inspection.
package localizednumbertrailinginputaccepted

import (
	"regexp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject unparsed number suffixes."

var trailingNumber = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?[A-DF-Za-df-z][A-Za-z]*$`)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "LocalizedNumberTrailingInputAccepted" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KEcho} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	e := n.(*syntax.Echo)
	if len(e.Exprs) != 1 {
		return
	}
	value := e.Exprs[0]
	parse, ok := semanticquery.NativeValue(ctx, value).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, parse, "NumberFormatter", "parse") {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(parse.Args, 0, "string"))
	if !known || !trailingNumber.MatchString(text) {
		return
	}
	offset := semanticquery.CallArgument(parse.Args, 2, "offset")
	input := semanticquery.CallArgument(parse.Args, 0, "string")
	for p := e.Parent(); p != nil; p = p.Parent() {
		if branch, ok := p.(*syntax.If); ok {
			checked := false
			syntax.Inspect(branch.Cond, func(n syntax.Node) bool {
				b, ok := n.(*syntax.Binary)
				if !ok || b.Op.Kind != syntax.TIsIdentical || !semanticquery.ExpansionCSame(ctx, b.Left, offset) {
					return true
				}
				length, ok := b.Right.(*syntax.FuncCall)
				if ok && semanticquery.NativeBuiltin(ctx, length, "strlen") && semanticquery.ExpansionCSame(ctx, semanticquery.CallArgument(length.Args, 0, "string"), input) {
					checked = true
				}
				return true
			})
			if checked {
				return
			}
		}
	}

	ctx.Report(syntax.Span{Start: e.Span().Start, End: e.Span().End - 1}, message)
}
