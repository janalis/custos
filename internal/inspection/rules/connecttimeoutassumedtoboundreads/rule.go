// Package connecttimeoutassumedtoboundreads implements the native ConnectTimeoutAssumedToBoundReads inspection.
package connecttimeoutassumedtoboundreads

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Set a separate timeout for stream reads."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ConnectTimeoutAssumedToBoundReads" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "fgets", "fread", "stream_get_contents", "fgetcsv":
	default:
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	origin, ok := semanticquery.NativeValue(ctx, handle).(*syntax.FuncCall)
	if !ok {
		return
	}
	name := semanticquery.NativeBuiltinName(ctx, origin)
	if name != "fsockopen" && name != "pfsockopen" {
		return
	}
	timeout, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(origin.Args, 4, "timeout"))
	if !known || timeout <= 0 {
		return
	}

	for parent := c.Parent(); parent != nil; parent = parent.Parent() {
		if branch, ok := parent.(*syntax.If); ok {
			configured := false
			syntax.Inspect(branch.Cond, func(n syntax.Node) bool {
				if call, ok := n.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, call, "stream_set_timeout") && semanticquery.ExpansionCSame(ctx, handle, semanticquery.CallArgument(call.Args, 0, "stream")) {
					configured = true
				}
				return true
			})
			if configured {
				return
			}
		}
	}

	ctx.ReportNode(c, message)
}
