// Package signalhandlerhasnodispatchmechanism implements the native SignalHandlerHasNoDispatchMechanism inspection.
package signalhandlerhasnodispatchmechanism

import (
	"strings"

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
	if mayDispatch(ctx, loop) || ticksEnabled(ctx, loop) {
		return
	}
	ctx.ReportNode(loop, message)
}

// mayDispatch reports whether the loop condition or body runs anything other
// than a builtin sleep: an explicit dispatch, or code that cannot be proven
// not to dispatch (application calls, methods, constructors, included files,
// evaluated code and generator suspension).
func mayDispatch(ctx *analysis.Context, loop *syntax.While) bool {
	found := false
	syntax.Inspect(loop, func(node syntax.Node) bool {
		switch node := node.(type) {
		case *syntax.FuncCall:
			name := semanticquery.NativeBuiltinName(ctx, node)
			if name != "usleep" && name != "sleep" {
				found = true
			}
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Eval, *syntax.Yield, *syntax.YieldFrom:
			found = true
		}
		return !found
	})
	return found
}

// ticksEnabled reports whether a ticks directive governs the loop: a block
// declare around it, or a statement-form declare earlier in the file. The
// latest such directive wins, and a literal zero switches ticks off.
func ticksEnabled(ctx *analysis.Context, loop *syntax.While) bool {
	enabled := false
	syntax.InspectFile(ctx.File, func(node syntax.Node) bool {
		if node.Span().Start >= loop.Span().Start {
			return false
		}
		directive, ok := node.(*syntax.Declare)
		if !ok {
			return true
		}
		// A block form ending before the loop no longer applies to it.
		if directive.Body != nil && directive.Span().End < loop.Span().End {
			return false
		}
		for _, item := range directive.Items {
			if strings.EqualFold(item.Key.Value, "ticks") {
				n, known := semanticquery.NativeInt(ctx, item.Value)
				enabled = !known || n != 0
			}
		}
		return true
	})
	return enabled
}
