// Package millisecondsusedasunixseconds implements the MillisecondsUsedAsUnixSeconds inspection.
package millisecondsusedasunixseconds

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MillisecondsUsedAsUnixSeconds" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall} }

const message = "Convert millisecond timestamps to seconds explicitly."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var value syntax.Expr
	if call, _ := semanticquery.GlobalCall(ctx, n, "date", "gmdate"); call != nil {
		value = semanticquery.CallArgument(call.Args, 1, "timestamp")
	}
	if call, ok := n.(*syntax.MethodCall); ok {
		name, ok := call.Name.(*syntax.Identifier)
		if ok && strings.EqualFold(name.Value, "setTimestamp") && semanticquery.NativeDateClass(ctx, call.Var) != "" {
			value = semanticquery.CallArgument(call.Args, 0, "timestamp")
		}
	}
	stamp, known := semanticquery.NativeInt(ctx, value)
	if known && stamp >= 100000000000 && stamp <= 9999999999999 {
		ctx.ReportNode(n, message)
	}
}

func (rule) Semantic() {}
