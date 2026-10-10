// Package waitstatususedasexitcode implements the native WaitStatusUsedAsExitCode inspection.
package waitstatususedasexitcode

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Decode the child wait status."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "WaitStatusUsedAsExitCode" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
	default:
		return
	}
	value, ok := semanticquery.NativeInt(ctx, b.Right)
	if !ok || value < 1 || value > 255 {
		return
	}
	wait := semanticquery.ExpansionCOutputCall(ctx, b, b.Left, []string{"pcntl_waitpid"}, 1, "status")
	if wait == nil {
		wait = semanticquery.ExpansionCOutputCall(ctx, b, b.Left, []string{"pcntl_wait"}, 0, "status")
	}
	if wait != nil {
		ctx.ReportNode(b, message)
	}
}
