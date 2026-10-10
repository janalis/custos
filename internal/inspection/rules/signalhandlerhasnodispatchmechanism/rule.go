// Package signalhandlerhasnodispatchmechanism implements the native SignalHandlerHasNoDispatchMechanism inspection.
package signalhandlerhasnodispatchmechanism

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Dispatch signals when asynchronous handling is disabled."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SignalHandlerHasNoDispatchMechanism" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KWhile} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.While)
	disabled := false
	handler := false
	for _, fact := range ctx.Flow().Calls(syntax.EnclosingVariableScope(loop)) {
		call, ok := fact.Node.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeDominates(call, loop) {
			continue
		}
		switch semanticquery.NativeBuiltinName(ctx, call) {
		case "pcntl_async_signals":
			v, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(call.Args, 0, "enable"))
			if known {
				disabled = !v
			}
		case "pcntl_signal":
			arg := semanticquery.CallArgument(call.Args, 1, "handler")
			_, handler = arg.(*syntax.Closure)
		}
	}
	if !disabled || !handler {
		return
	}
	safe := false
	syntax.Inspect(loop.Body, func(node syntax.Node) bool {
		if call, ok := node.(*syntax.FuncCall); ok {
			name := semanticquery.NativeBuiltinName(ctx, call)
			if name == "pcntl_signal_dispatch" || (name != "usleep" && name != "sleep") {
				safe = true
			}
		}
		return true
	})
	for p := loop.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.Declare); ok {
			safe = true
		}
	}
	if !safe {
		ctx.ReportNode(loop, message)
	}
}
