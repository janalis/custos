// Package hexdecodeoddlengthliteral implements the native HexDecodeOddLengthLiteral inspection.
package hexdecodeoddlengthliteral

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Provide an even number of hexadecimal digits."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HexDecodeOddLengthLiteral" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP54 {
		return
	}
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "hex2bin") {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 0, "string")
	s, ok := semanticquery.NativeString(ctx, arg)
	if !ok || len(s)%2 == 0 {
		return
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return
		}
	}
	ctx.ReportNode(arg, message)
}
