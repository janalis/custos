// Package zipopentruthyresult implements the native ZipOpenTruthyResult inspection.
package zipopentruthyresult

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Accept only true from ZipArchive::open."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipOpenTruthyResult" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "open") {
		return
	}
	parent, child := astquery.ParentSkipParens(c)
	suffix := "===true"
	switch p := parent.(type) {
	case *syntax.If, *syntax.While:
	case *syntax.Unary:
		if p.Op.Kind != syntax.TExclaim {
			return
		}
		suffix = "!==true"
		written := ctx.Text(p)
		condition, _ := astquery.ParentSkipParens(p)
		_, isIf := condition.(*syntax.If)
		_, isWhile := condition.(*syntax.While)
		if (isIf || isWhile) && !strings.Contains(written, "/*") && !strings.Contains(written, "//") && !strings.Contains(written, "#") {
			ctx.ReportNode(c, message, astquery.ReplaceFix(p.Span(), "("+ctx.Text(c)+suffix+")"))
		} else {
			ctx.ReportNode(c, message)
		}
		return
	default:
		return
	}
	ctx.ReportNode(c, message, astquery.ReplaceFix(child.Span(), "("+ctx.Text(child)+suffix+")"))
}
