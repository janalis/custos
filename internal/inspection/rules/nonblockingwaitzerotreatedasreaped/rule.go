// Package nonblockingwaitzerotreatedasreaped implements the native NonblockingWaitZeroTreatedAsReaped inspection.
package nonblockingwaitzerotreatedasreaped

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Require a positive PID from nonblocking wait."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonblockingWaitZeroTreatedAsReaped" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsNotIdentical && b.Op.Kind != syntax.TIsNotEqual && b.Op.Kind != syntax.TIsGreaterOrEqual {
		return
	}
	value, ok := semanticquery.NativeInt(ctx, b.Right)
	if !ok || !((b.Op.Kind == syntax.TIsGreaterOrEqual && value == 0) || (b.Op.Kind != syntax.TIsGreaterOrEqual && value == -1)) {
		return
	}
	wait, ok := semanticquery.NativeValue(ctx, b.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, wait, "pcntl_waitpid") {
		return
	}
	flag := semanticquery.CallArgument(wait.Args, 2, "flags")
	if hasConstant(ctx, flag, "WNOHANG") {
		ctx.ReportNode(b, message)
	}
}

func hasConstant(ctx *analysis.Context, e syntax.Expr, name string) bool {
	value := semanticquery.NativeValue(ctx, e)
	c, ok := value.(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(c.Name.Value, name) && semanticquery.BareReachesGlobal(ctx, name, c.Span().Start)
}
