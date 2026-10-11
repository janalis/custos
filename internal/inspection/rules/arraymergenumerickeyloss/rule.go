// Package arraymergenumerickeyloss implements the native ArrayMergeNumericKeyLoss inspection.
package arraymergenumerickeyloss

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve numeric identifier keys when combining arrays."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayMergeNumericKeyLoss" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_merge") {
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok {
		return
	}
	for _, arg := range args {
		if a := semanticquery.NativeArray(ctx, arg); a != nil && renumbers(ctx, a) {
			ctx.ReportNode(call, message)
			return
		}
	}
}

// renumbers reports whether array_merge provably moves an integer key of the
// literal to another index. Integer keys are renumbered from zero in entry
// order, so string-keyed items and overwritten duplicates take no index.
func renumbers(ctx *analysis.Context, a *syntax.Array) bool {
	// Without unknown keys, the integer keys seen so far are exactly 0..index-1.
	// Each unknown key may add one more integer entry before the next item.
	index, unknown := int64(0), int64(0)
	for _, item := range a.Items {
		if item.Key == nil {
			index++
			continue
		}
		key, known := semanticquery.NativeArrayKey(ctx, item.Key)
		switch {
		case !known:
			unknown++
		case strings.HasPrefix(key, "i:"):
			n, _ := strconv.ParseInt(key[2:], 10, 64)
			switch {
			case n < 0 || n > index+unknown:
				return true
			case unknown > 0:
				unknown++
			case n == index:
				index++
			}
		}
	}
	return false
}
