// Package selectfailuretreatedasreadiness implements the native SelectFailureTreatedAsReadiness inspection.
package selectfailuretreatedasreadiness

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
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
	if !ok || !semanticquery.NativeBuiltin(ctx, call, "stream_select") || failureExcluded(ctx, b) {
		return
	}
	ctx.ReportNode(b, message)
}

// failureExcluded reports whether a test of the same stored result already
// ruled failure out before the readiness comparison runs.
// Direct calls never qualify: a second call is a different result.
func failureExcluded(ctx *analysis.Context, b *syntax.Binary) bool {
	if _, ok := syntax.UnwrapParens(b.Left).(*syntax.Variable); !ok {
		return false
	}
	return semanticquery.NativeFailureExcluded(ctx, b.Left, types.Of("int", "false"), "false")
}
