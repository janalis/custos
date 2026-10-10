// Package arraycolumnduplicateindexloss implements the native ArrayColumnDuplicateIndexLoss inspection.
package arraycolumnduplicateindexloss

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use unique index values or group the duplicate rows."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayColumnDuplicateIndexLoss" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_column") {
		return
	}
	a, effective := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	column, cok := semanticquery.NativeArrayKey(ctx, semanticquery.CallArgument(call.Args, 1, "column_key"))
	index, iok := semanticquery.NativeArrayKey(ctx, semanticquery.CallArgument(call.Args, 2, "index_key"))
	if !effective || !cok || !iok {
		return
	}
	seen := map[string]bool{}
	for _, item := range a {
		row, ok := semanticquery.NativeArrayEntries(ctx, item)
		if !ok {
			return
		}
		if _, ok := row[column]; !ok {
			continue
		}
		value, ok := row[index]
		if !ok {
			continue
		}
		key, known := semanticquery.NativeArrayKey(ctx, value)
		if !known {
			return
		}
		if seen[key] {
			ctx.ReportNode(call, message)
			return
		}
		seen[key] = true
	}
}
