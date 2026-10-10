// Package sortresultusedasarray implements the native SortResultUsedAsArray inspection.
package sortresultusedasarray

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the sorted input array instead of the success flag."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SortResultUsedAsArray" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach, syntax.KFuncCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var input syntax.Expr
	switch x := n.(type) {
	case *syntax.Foreach:
		input = x.Expr
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, x) {
		case "array_values", "array_keys", "array_sum", "array_reverse", "array_filter", "array_merge":
			input = semanticquery.CallArgument(x.Args, 0, "array")
		default:
			return
		}
	}
	call, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok {
		return
	}
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "sort", "rsort", "asort", "arsort", "ksort", "krsort", "usort", "uasort", "uksort":
		ctx.ReportNode(n, message)
	}
}
