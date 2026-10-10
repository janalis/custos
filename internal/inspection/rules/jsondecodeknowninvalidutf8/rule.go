// Package jsondecodeknowninvalidutf8 implements the native JsonDecodeKnownInvalidUtf8 inspection.
package jsondecodeknowninvalidutf8

import (
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Decode valid UTF-8 JSON input."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonDecodeKnownInvalidUtf8" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_decode") {
		return
	}
	subject, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "json"))
	if !k || utf8.ValidString(subject) {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 3, "flags")
	if flags != nil {
		bits, known := semanticquery.NativeContractInt(ctx, flags)
		if !known || bits&(1048576|2097152) != 0 {
			return
		}
	}
	ctx.ReportNode(c, message)
}
