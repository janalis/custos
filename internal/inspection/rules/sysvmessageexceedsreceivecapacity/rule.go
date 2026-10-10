// Package sysvmessageexceedsreceivecapacity implements the native SysvMessageExceedsReceiveCapacity inspection.
package sysvmessageexceedsreceivecapacity

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Allow enough receive capacity for the message."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SysvMessageExceedsReceiveCapacity" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "msg_receive") {
		return
	}
	send := semanticquery.ExpansionDMessage(ctx, c)
	if send == nil {
		return
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
