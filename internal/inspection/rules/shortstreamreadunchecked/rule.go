// Package shortstreamreadunchecked implements the native ShortStreamReadUnchecked inspection.
package shortstreamreadunchecked

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the complete record before decoding it."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ShortStreamReadUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "unpack") {
		return
	}
	format, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "format"))
	if !known {
		return
	}
	input := semanticquery.CallArgument(call.Args, 1, "string")
	inner, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, inner, "fread") {
		return
	}
	length, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(inner.Args, 1, "length"))
	if !known || length < 1 {
		return
	}
	// A single fixed-width network integer is independent of machine word size.
	size := int64(0)
	if format != "" {
		switch format[0] {
		case 'N', 'V':
			size = 4
		case 'n', 'v':
			size = 2
		case 'C', 'c':
			size = 1
		}
	}
	if size == 0 || length != size || strings.ContainsAny(format, "/*0123456789") {
		return
	}
	if semanticquery.NativeLengthGuard(ctx, input, size) {
		return
	}
	ctx.ReportNode(call, message)
}
