// Package socketreadzerorejected implements SocketReadZeroRejected.
package socketreadzerorejected

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare socket read failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SocketReadZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "socket_read") && socketReadBinaryMode(ctx, c) && semanticquery.FailureBranch(c) != nil {
		ctx.ReportNode(c, message)
	}
}

// socketReadBinaryMode proves the binary read mode: either the builtin constant
// is passed, or the mode is left to its default, which is PHP_BINARY_READ. A
// spread or a first-class callable placeholder may bind anything.
func socketReadBinaryMode(ctx *analysis.Context, c *syntax.FuncCall) bool {
	positional := 0
	for _, node := range c.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return false
		}
		if arg.Name == nil {
			positional++
		} else if arg.Name.Value == "mode" {
			positional = 3
		}
	}
	if positional < 3 {
		return true
	}
	return semanticquery.IsBuiltinConst(ctx, semanticquery.CallArgument(c.Args, 2, "mode"), "PHP_BINARY_READ")
}
