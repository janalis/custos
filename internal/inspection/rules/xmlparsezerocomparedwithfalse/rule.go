// Package xmlparsezerocomparedwithfalse implements the native XmlParseZeroComparedWithFalse inspection.
package xmlparsezerocomparedwithfalse

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Compare XML parsing failure with integer zero."

type rule struct{}

// New returns the XmlParseZeroComparedWithFalse inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlParseZeroComparedWithFalse" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
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
