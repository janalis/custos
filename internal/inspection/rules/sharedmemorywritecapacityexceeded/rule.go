// Package sharedmemorywritecapacityexceeded implements the native SharedMemoryWriteCapacityExceeded inspection.
package sharedmemorywritecapacityexceeded

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep writes within shared-memory capacity."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SharedMemoryWriteCapacityExceeded" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "shmop_write") {
		return
	}
	data, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "data"))
	if !known {
		return
	}
	segment := semanticquery.CallArgument(c.Args, 0, "shmop")
	if !semanticquery.ExpansionDUnaliased(ctx, segment, n) {
		return
	}
	offset := semanticquery.CallArgument(c.Args, 2, "offset")
	if b, ok := syntax.UnwrapParens(offset).(*syntax.Binary); ok && b.Op.Kind == syntax.TMinus {
		size, ok := syntax.UnwrapParens(b.Left).(*syntax.FuncCall)
		remaining, k := semanticquery.NativeContractInt(ctx, b.Right)
		if ok && semanticquery.NativeBuiltin(ctx, size, "shmop_size") && semanticquery.NativeSameValue(ctx, segment, semanticquery.CallArgument(size.Args, 0, "shmop")) && k && remaining > 0 && int64(len(data)) > remaining {
			ctx.ReportNode(n, message)
		}
		return
	}
	start, ok := semanticquery.NativeLocalValue(ctx, segment).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, start, "shmop_open") {
		return
	}
	flags, fk := semanticquery.NativeString(ctx, semanticquery.CallArgument(start.Args, 1, "mode"))
	if !fk || flags != "n" {
		return
	}
	size, sk := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(start.Args, 3, "size"))
	position, pk := semanticquery.NativeContractInt(ctx, offset)
	if sk && pk && size > 0 && position >= 0 && position <= size && int64(len(data)) > size-position {
		ctx.ReportNode(n, message)
	}
}
