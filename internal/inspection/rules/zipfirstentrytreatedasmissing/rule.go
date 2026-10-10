// Package zipfirstentrytreatedasmissing implements the native ZipFirstEntryTreatedAsMissing inspection.
package zipfirstentrytreatedasmissing

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Distinguish a missing ZIP entry from index zero."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipFirstEntryTreatedAsMissing" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "locateName") {
		return
	}
	if !semanticquery.ExpansionCTruthy(c) {
		return
	}
	p, child := astquery.ParentSkipParens(c)
	if a, ok := p.(*syntax.Assign); ok {
		p, child = astquery.ParentSkipParens(a)
	}
	suffix := " !== false"
	if u, ok := p.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		prefix := ctx.SpanText(syntax.Span{Start: u.Span().Start, End: child.Span().Start})
		if strings.Contains(prefix, "/*") || strings.Contains(prefix, "//") || strings.Contains(prefix, "#") {
			ctx.ReportNode(u, message)
			return
		}
		text := ctx.Text(child)
		if _, assignment := child.(*syntax.Assign); assignment {
			text = "(" + text + ")"
		}
		ctx.ReportNode(u, message, astquery.ReplaceFix(u.Span(), text+" === false"))
		return
	}
	ctx.ReportNode(child, message, astquery.ReplaceFix(child.Span(), "("+ctx.Text(child)+")"+suffix))
}
