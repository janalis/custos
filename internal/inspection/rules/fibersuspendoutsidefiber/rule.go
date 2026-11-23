// Package fibersuspendoutsidefiber implements the native
// FiberSuspendOutsideFiber inspection.
package fibersuspendoutsidefiber

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Suspend only inside a running fiber."

type rule struct{}

// New returns the FiberSuspendOutsideFiber inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberSuspendOutsideFiber" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KStaticCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || ctx.PHP.Below(phpversion.PHP81) || !semanticquery.NativeCallbackReachable(ctx, n) || !semanticquery.BuiltinStaticCall(ctx, n.(syntax.Expr), "Fiber", "suspend") {
		return
	}
	for p := n.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(n.Span()) {
			continue
		}
		b, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TIsIdentical {
			continue
		}
		if (syntax.IsNullConst(syntax.UnwrapParens(b.Right)) && semanticquery.BuiltinStaticCall(ctx, syntax.UnwrapParens(b.Left), "Fiber", "getCurrent")) || (syntax.IsNullConst(syntax.UnwrapParens(b.Left)) && semanticquery.BuiltinStaticCall(ctx, syntax.UnwrapParens(b.Right), "Fiber", "getCurrent")) {
			ctx.ReportNode(n, message)
			return
		}
	}
}
