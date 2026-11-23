// Package sysvmessageexceedsreceivecapacity implements the native SysvMessageExceedsReceiveCapacity inspection.
package sysvmessageexceedsreceivecapacity

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Allow enough receive capacity for the message."

type rule struct{}

// New returns the SysvMessageExceedsReceiveCapacity inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SysvMessageExceedsReceiveCapacity" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "msg_receive") {
		return
	}
	send := semanticquery.MatchingQueueSend(ctx, c)
	if send == nil {
		return
	}
	for _, arg := range []syntax.Expr{
		semanticquery.CallArgument(c.Args, 1, "desired_message_type"),
		semanticquery.CallArgument(c.Args, 3, "max_message_size"),
		semanticquery.CallArgument(c.Args, 6, "flags"),
		semanticquery.CallArgument(send.Args, 1, "message_type"),
		semanticquery.CallArgument(send.Args, 2, "message"),
		semanticquery.CallArgument(send.Args, 3, "serialize"),
	} {
		if arg != nil && stepped(ctx, arg) {
			return
		}
	}
	serialize, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(send.Args, 3, "serialize"))
	if !known || serialize {
		return
	}
	data, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(send.Args, 2, "message"))
	if !known {
		return
	}
	capacity, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 3, "max_message_size"))
	if !known || capacity <= 0 || int64(len(data)) <= capacity {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 6, "flags")
	if flags != nil {
		v, k := semanticquery.NativeContractInt(ctx, flags)
		if !k || v&2 != 0 {
			return
		}
	}
	ctx.ReportNode(n, message)
}

// stepped reports whether a variable read by e may observe an increment, a
// decrement or a write through a by-reference closure capture. Local flow
// facts keep the assigned literal across those constructs, so the value is no
// longer proven. A later change still counts when a loop repeats the read.
func stepped(ctx *analysis.Context, e syntax.Expr) bool {
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		if use, ok := n.(*syntax.Variable); ok && !found {
			found = flowquery.SteppedBefore(ctx.File, use)
		}
		return !found
	})
	return found
}
