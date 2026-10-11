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
	calls := semanticquery.NativePriorCalls(ctx, c, handle, "flock")
	if len(calls) == 0 {
		return
	}
	previous, ok := calls[len(calls)-1].(*syntax.FuncCall)
	if ok && semanticquery.NativeBuiltin(ctx, previous, "flock") {
		mode := semanticquery.CallArgument(previous.Args, 1, "operation")
		if shared(ctx, mode) && !relocked(ctx, previous, c, handle) {
			ctx.ReportNode(c, message)
		}
	}
}

// relocked reports whether a flock call between the proven shared lock and
// the write may have changed the lock mode. Checked upgrades sit in conditions
// or guarded blocks, so every reached call counts, not only dominating ones.
func relocked(ctx *analysis.Context, lock, write *syntax.FuncCall, handle syntax.Expr) bool {
	id := ctx.Flow().Value(handle).Identity
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(write)) {
		later, ok := call.Node.(*syntax.FuncCall)
		if !ok || later.Span().Start < lock.Span().End || later.Span().End > write.Span().Start || !semanticquery.NativeBuiltin(ctx, later, "flock") {
			continue
		}
		// Only a handle proven to be another stream leaves the lock untouched.
		other := ctx.Flow().Value(semanticquery.CallArgument(later.Args, 0, "stream"))
		if other.Complete && other.Identity != 0 && other.Identity != id {
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
