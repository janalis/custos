// Package callbackrequiredargumentmismatch implements the native CallbackRequiredArgumentMismatch inspection.
package callbackrequiredargumentmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the callback parameters to the supplied arrays."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CallbackRequiredArgumentMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_map") {
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) < 2 {
		return
	}
	params, body := semanticquery.NativeCallback(ctx, args[0])
	if body == nil {
		return
	}
	required := 0
	for _, param := range params {
		if param.Default == nil && !param.Variadic {
			required++
		}
	}
	if required > len(args)-1 {
		ctx.ReportNode(call, message)
	}
}
