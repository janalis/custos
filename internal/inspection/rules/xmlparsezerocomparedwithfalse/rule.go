// Package xmlparsezerocomparedwithfalse implements the native XmlParseZeroComparedWithFalse inspection.
package xmlparsezerocomparedwithfalse

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare XML parsing failure with integer zero."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlParseZeroComparedWithFalse" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		truth, known := astquery.BoolConst(syntax.UnwrapParens(pair[1]))
		if !known || truth {
			continue
		}
		c, ok := semanticquery.NativeLocalValue(ctx, pair[0]).(*syntax.FuncCall)
		if ok && semanticquery.NativeBuiltin(ctx, c, "xml_parse") {
			ctx.ReportNode(n, message)
			return
		}
	}
}
