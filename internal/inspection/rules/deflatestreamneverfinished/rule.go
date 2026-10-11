// Package deflatestreamneverfinished implements the native DeflateStreamNeverFinished inspection.
package deflatestreamneverfinished

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Finish the compressed stream before publishing it."

// fileAppend is the value of PHP's FILE_APPEND flag.
const fileAppend = 8

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DeflateStreamNeverFinished" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "file_put_contents") {
		return
	}
	output := semanticquery.CallArgument(c.Args, 1, "data")
	chunk, ok := semanticquery.NativeValue(ctx, output).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, chunk, "deflate_add") {
		return
	}
	mode := semanticquery.CallArgument(chunk.Args, 2, "flush_mode")
	if !hasConstant(ctx, mode, "ZLIB_SYNC_FLUSH") && !hasConstant(ctx, mode, "ZLIB_NO_FLUSH") && !hasConstant(ctx, mode, "ZLIB_FULL_FLUSH") {
		return
	}
	// An appending write publishes one piece of a file that a later write may
	// finish, so only a provably truncating write is a standalone stream.
	if flags := semanticquery.CallArgument(c.Args, 2, "flags"); flags != nil {
		if bits, known := semanticquery.NativeContractInt(ctx, flags); !known || bits&fileAppend != 0 {
			return
		}
	}
	ctx.ReportNode(c, message)
}

func hasConstant(ctx *analysis.Context, e syntax.Expr, name string) bool {
	value := semanticquery.NativeValue(ctx, e)
	c, ok := value.(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(c.Name.Value, name) && semanticquery.BareReachesGlobal(ctx, name, c.Span().Start)
}
