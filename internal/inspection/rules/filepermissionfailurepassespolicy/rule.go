// Package filepermissionfailurepassespolicy implements the native FilePermissionFailurePassesPolicy inspection.
package filepermissionfailurepassespolicy

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
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
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
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
	if !semanticquery.UnaliasedLocal(ctx, mask.Left, n) {
		return
	}
	c, ck := semanticquery.NativeLocalBuiltinResult(ctx, mask.Left, "false", "fileperms").(*syntax.FuncCall)
	if !ck || !semanticquery.NativeBuiltin(ctx, c, "fileperms") || semanticquery.NativeFailureExcludedSingle(ctx, mask.Left, types.Of("int", "false"), "false") {
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if ok && semanticquery.TerminalOutcome(ctx, branch.Body, true) {
		ctx.ReportNode(b, message)
	}
}
