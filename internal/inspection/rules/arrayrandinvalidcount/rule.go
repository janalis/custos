// Package arrayrandinvalidcount implements the native ArrayRandInvalidCount inspection.
package arrayrandinvalidcount

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Request a count within the array size."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayRandInvalidCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "array_rand") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 1, "num")
	count, k := int64(1), true
	if arg != nil {
		count, k = semanticquery.NativeInt(ctx, arg)
		k = k && semanticquery.StableLocalValue(ctx, arg, semanticquery.LocalStability{}) && !property(arg)
	}
	array := semanticquery.CallArgument(c.Args, 0, "array")
	entries, known := semanticquery.NativeArrayEntries(ctx, array)
	known = known && semanticquery.StableLocalValue(ctx, array, semanticquery.LocalStability{}) && !property(array)
	if k && (count < 1 || known && count > int64(len(entries))) {
		ctx.ReportNode(c, message)
	}
}

// property reports whether e reads a property: other methods may have
// changed it since its default or the constructor's write (D2).
func property(e syntax.Expr) bool {
	switch syntax.UnwrapParens(e).(type) {
	case *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
		return true
	}
	return false
}
