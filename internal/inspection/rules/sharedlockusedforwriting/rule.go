// Package sharedlockusedforwriting implements the native SharedLockUsedForWriting inspection.
package sharedlockusedforwriting

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Acquire an exclusive lock before writing."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SharedLockUsedForWriting" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "fwrite" && name != "fputs" && name != "ftruncate" {
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	history := semanticquery.NativeCallHistory(ctx, c, handle, "flock")
	if !history.Complete {
		return
	}
	var previousNode syntax.Node
	for _, event := range history.Events {
		if event.Must {
			previousNode = event.Node
		}
	}
	previous, ok := previousNode.(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltin(ctx, previous, "flock") {
		mode := semanticquery.CallArgument(previous.Args, 1, "operation")
		if shared(ctx, mode) && !relocked(ctx, previous, history) {
			ctx.ReportNode(c, message)
		}
	}
}

// relocked reports a possible lock-mode change after the last mandatory lock.
func relocked(ctx *analysis.Context, lock *syntax.FuncCall, history semanticquery.CallHistory) bool {
	for _, event := range history.Events {
		later, ok := event.Node.(*syntax.FuncCall)
		if !ok || later.Span().Start < lock.Span().End || !semanticquery.NativeBuiltin(ctx, later, "flock") {
			continue
		}
		if !shared(ctx, semanticquery.CallArgument(later.Args, 1, "operation")) {
			return true
		}
	}
	return false
}

func shared(ctx *analysis.Context, e syntax.Expr) bool {
	if c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch); ok {
		return semanticquery.GlobalConstName(ctx, c) == "LOCK_SH"
	}
	if b, ok := syntax.UnwrapParens(e).(*syntax.Binary); ok && b.Op.Kind == syntax.TBar {
		left, lk := syntax.UnwrapParens(b.Left).(*syntax.ConstFetch)
		right, rk := syntax.UnwrapParens(b.Right).(*syntax.ConstFetch)
		return lk && rk && (semanticquery.GlobalConstName(ctx, left) == "LOCK_SH" && semanticquery.GlobalConstName(ctx, right) == "LOCK_NB" || semanticquery.GlobalConstName(ctx, left) == "LOCK_NB" && semanticquery.GlobalConstName(ctx, right) == "LOCK_SH")
	}
	return false
}
