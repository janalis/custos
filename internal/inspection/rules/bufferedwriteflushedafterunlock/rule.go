// Package bufferedwriteflushedafterunlock implements the native BufferedWriteFlushedAfterUnlock inspection.
package bufferedwriteflushedafterunlock

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Flush buffered writes before releasing the lock."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "BufferedWriteFlushedAfterUnlock" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fflush") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "stream")
	locked, dirty, released := false, false, false
	for _, prior := range priorCalls(ctx, c, h) {
		switch semanticquery.NativeBuiltinName(ctx, prior) {
		case "flock":
			bits, k := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(prior.Args, 1, "operation"))
			if !k {
				locked = false
				dirty = false
				released = false
				continue
			}
			switch bits & 3 {
			case 2:
				locked = true
				released = false
			case 3:
				released = locked && dirty
				locked = false
			default:
				locked = false
				dirty = false
				released = false
			}
		case "fwrite":
			if locked {
				dirty = true
			}
		case "fflush", "fclose":
			dirty = false
			released = false
		}
	}
	if released && dirty {
		ctx.ReportNode(c, message)
	}
}

func priorCalls(ctx *analysis.Context, at syntax.Node, h syntax.Expr) []*syntax.FuncCall {
	var calls []*syntax.FuncCall
	variable, ok := syntax.UnwrapParens(h).(*syntax.Variable)
	if !ok {
		return nil
	}
	for _, event := range semanticquery.ExpansionDLocalEvents(ctx, at) {
		if event.Span().Start >= at.Span().Start {
			break
		}
		if assign, ok := event.(*syntax.Assign); ok {
			if v, ok := assign.Var.(*syntax.Variable); ok && v.Name == variable.Name {
				calls = nil
			}
			if assign.ByRef || semanticquery.NativeSameValue(ctx, assign.Value, h) {
				return nil // A stream alias may be flushed outside this variable history.
			}
		}
		if _, ok := event.(*syntax.Unset); ok {
			calls = nil
		}
		if f, ok := event.(*syntax.FuncCall); ok {
			arg, vok := syntax.UnwrapParens(semanticquery.CallArgument(f.Args, 0, "stream")).(*syntax.Variable)
			if !vok || arg.Name != variable.Name {
				continue
			}
			switch semanticquery.NativeBuiltinName(ctx, f) {
			case "flock", "fwrite", "fflush", "fclose":
				if !semanticquery.NativeDominates(f, at) {
					calls = nil
					continue
				}
				calls = append(calls, f)
			default:
				calls = nil
			}
		}
		if _, ok := event.(*syntax.MethodCall); ok {
			calls = nil
		}
	}
	return calls
}
