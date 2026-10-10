// Package selectfailuretreatedasreadiness implements the native SelectFailureTreatedAsReadiness inspection.
package selectfailuretreatedasreadiness

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle selection failure before processing ready streams."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SelectFailureTreatedAsReadiness" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	value, ok := semanticquery.NativeInt(ctx, b.Right)
	if !ok || value != 0 {
		return
	}
	call, ok := semanticquery.NativeValue(ctx, b.Left).(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltin(ctx, call, "stream_select") {
		ctx.ReportNode(b, message)
	}
}
