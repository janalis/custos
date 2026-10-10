// Package permissionmodewrittenindecimal implements the native PermissionModeWrittenInDecimal inspection.
package permissionmodewrittenindecimal

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Write the permission mask in octal notation."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PermissionModeWrittenInDecimal" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "chmod") && !semanticquery.NativeBuiltin(ctx, c, "mkdir") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 1, "permissions")
	literal, ok := syntax.UnwrapParens(arg).(*syntax.Literal)
	if !ok {
		return
	}
	text := ctx.Text(literal)
	if len(text) < 3 || len(text) > 4 || text[0] == '0' {
		return
	}
	for _, ch := range text {
		if ch < '0' || ch > '7' {
			return
		}
	}
	value, known := semanticquery.NativeInt(ctx, literal)
	if known && value > 0o777 {
		ctx.ReportNode(c, message)
	}
}
