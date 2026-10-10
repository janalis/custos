// Package umasksetterreturnmisread implements the native UmaskSetterReturnMisread inspection.
package umasksetterreturnmisread

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the new mask with a separate umask call."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UmaskSetterReturnMisread" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	c, ok := syntax.UnwrapParens(b.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "umask") {
		return
	}
	set, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 0, "mask"))
	expected, ek := semanticquery.NativeInt(ctx, b.Right)
	if !k || !ek || set == 0 || set != expected {
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if ok && semanticquery.ExpansionDFinalOutcome(ctx, branch.Body, false) {
		ctx.ReportNode(b, message)
	}
}
