// Package arraycolumnmissingfield implements the native ArrayColumnMissingField inspection.
package arraycolumnmissingfield

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Handle rows missing the selected field explicitly."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayColumnMissingField" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_column") || ctx.PHP < phpversion.PHP55 {
		return
	}
	rows, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	if !known {
		return
	}
	key, known := semanticquery.NativeArrayKey(ctx, semanticquery.CallArgument(call.Args, 1, "column_key"))
	if !known {
		return
	}
	present, missing := false, false
	for _, row := range rows {
		entries, ok := semanticquery.NativeArrayEntries(ctx, row)
		if !ok {
			return
		}
		if _, exists := entries[key]; exists {
			present = true
		} else {
			missing = true
		}
	}
	if present && missing {
		ctx.ReportNode(call, message)
	}
}
