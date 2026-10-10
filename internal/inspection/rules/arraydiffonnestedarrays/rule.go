// Package arraydiffonnestedarrays implements the native ArrayDiffOnNestedArrays inspection.
package arraydiffonnestedarrays

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare nested arrays structurally."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayDiffOnNestedArrays" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_diff") || call.Args == nil {
		return
	}
	for i, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || (arg.Name != nil && arg.Name.Value != "array") {
			return
		}
		a, known := semanticquery.NativeArrayEntries(ctx, arg.Value)
		if !known {
			continue
		}
		for _, item := range a {
			if semanticquery.NativeArray(ctx, item) != nil {
				ctx.ReportNode(call, message)
				return
			}
		}
		_ = i
	}
}
