// Package negativebitshiftcount implements the native NegativeBitShiftCount inspection.
package negativebitshiftcount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a nonnegative shift count."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NegativeBitShiftCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TSl && b.Op.Kind != syntax.TSr {
		return
	}
	v, k := semanticquery.NativeInt(ctx, b.Right)
	if k && v < 0 {
		ctx.ReportNode(b, message)
	}
}
