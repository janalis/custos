// Package arraymultisortlengthmismatch implements the native ArrayMultisortLengthMismatch inspection.
package arraymultisortlengthmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Provide parallel arrays of equal length."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayMultisortLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_multisort") || call.Args == nil {
		return
	}
	length := -1
	for _, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack {
			return
		}
		a, known := semanticquery.NativeArrayEntries(ctx, arg.Value)
		if !known {
			continue
		}
		if length < 0 {
			length = len(a)
		} else if length != len(a) {
			ctx.ReportNode(call, message)
			return
		}
	}
}
