// Package filepermissionfailurepassespolicy implements the native FilePermissionFailurePassesPolicy inspection.
package filepermissionfailurepassespolicy

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject permission lookup failure before checking bits."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FilePermissionFailurePassesPolicy" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsEqual {
		return
	}
	zero, k := semanticquery.NativeInt(ctx, b.Right)
	mask, ok := syntax.UnwrapParens(b.Left).(*syntax.Binary)
	if !k || zero != 0 || !ok || mask.Op.Kind != syntax.TAmpersand {
		return
	}
	bits, known := semanticquery.NativeInt(ctx, mask.Right)
	if !known || bits == 0 {
		return
	}
	if !semanticquery.ExpansionDUnaliased(ctx, mask.Left, n) {
		return
	}
	c, ck := semanticquery.NativeLocalValue(ctx, mask.Left).(*syntax.FuncCall)
	if !ck || !semanticquery.NativeBuiltin(ctx, c, "fileperms") || semanticquery.NativeLocalSentinelGuard(ctx, mask.Left, "false") || ctx.Flow().Excludes(mask.Left, "false") {
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if ok && semanticquery.ExpansionDFinalOutcome(ctx, branch.Body, true) {
		ctx.ReportNode(b, message)
	}
}
