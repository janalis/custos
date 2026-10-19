// Package connecttimeoutassumedtoboundreads implements the native ConnectTimeoutAssumedToBoundReads inspection.
package connecttimeoutassumedtoboundreads

import (
	"strconv"
	"strings"

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
	if !positiveTimeout(ctx, semanticquery.CallArgument(origin.Args, 4, "timeout")) {
		return
	}

	if !timeoutSet(ctx, c, handle, origin) {
		ctx.ReportNode(c, message)
	}
}

// timeoutSet reports whether a stream_set_timeout call on the handle opened
// by origin, written before read in its scope, is known to have succeeded
// there.
func timeoutSet(ctx *analysis.Context, read *syntax.FuncCall, handle syntax.Expr, origin *syntax.FuncCall) bool {
	set := false
	visit := func(n syntax.Node) bool {
		if set || n.Span().Start >= read.Span().Start {
			return false
		}
		if n != syntax.EnclosingVariableScope(read) && syntax.IsVariableScope(n) {
			return false
		}
		if call, ok := n.(*syntax.FuncCall); ok && semanticquery.NativeBuiltin(ctx, call, "stream_set_timeout") {
			stream := semanticquery.CallArgument(call.Args, 0, "stream")
			set = semanticquery.ExpansionCSame(ctx, handle, stream) && semanticquery.NativeValue(ctx, stream) == syntax.Expr(origin) && semanticquery.NativeCallSucceeded(ctx, call, read)
		}
		return !set
	}
	if scope := syntax.EnclosingVariableScope(read); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return set
}

// positiveTimeout accepts a known integer or float literal above zero.
func positiveTimeout(ctx *analysis.Context, e syntax.Expr) bool {
	if n, ok := semanticquery.NativeInt(ctx, e); ok {
		return n > 0
	}
	lit, ok := semanticquery.NativeValue(ctx, e).(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitFloat {
		return false
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
	return err == nil && v > 0
}
