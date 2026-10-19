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
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
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
	if released && dirty && !escapes(ctx, c, syntax.UnwrapParens(h).(*syntax.Variable)) {
		ctx.ReportNode(c, message)
	}
}

// escapes reports whether code outside the straight-line stream calls can
// reach the handle: a closure or arrow function of the scope captures it, or
// it is handed to a call before the flush in any slot but a stream one. Such
// code may flush or close the stream without a visible builtin call.
func escapes(ctx *analysis.Context, at *syntax.FuncCall, variable *syntax.Variable) bool {
	scope := syntax.EnclosingVariableScope(at)
	escaped := false
	visit := func(n syntax.Node) bool {
		if escaped {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			escaped = captures(n, variable.Name)
			return false
		}
		v, ok := n.(*syntax.Variable)
		if !ok || v.Name != variable.Name || v.Span().Start >= at.Span().Start {
			return true
		}
		var value syntax.Node = v
		for paren, ok := value.Parent().(*syntax.Paren); ok; paren, ok = value.Parent().(*syntax.Paren) {
			value = paren
		}
		arg, ok := value.Parent().(*syntax.Arg)
		if !ok {
			return true
		}
		call, direct := arg.Parent().Parent().(*syntax.FuncCall)
		escaped = !direct || syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 0, "stream")) != syntax.Expr(v)
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return escaped
}

// captures reports whether a nested scope shares the named variable: a
// closure through its use list, an arrow function through any reference.
func captures(scope syntax.Node, name string) bool {
	switch fn := scope.(type) {
	case *syntax.Closure:
		for _, use := range fn.Uses {
			if use.Var.Name == name {
				return true
			}
		}
	case *syntax.ArrowFunction:
		found := false
		syntax.Inspect(fn.Expr, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.Name == name {
				found = true
			}
			return !found
		})
		return found
	}
	return false
}

func priorCalls(ctx *analysis.Context, at syntax.Node, h syntax.Expr) []*syntax.FuncCall {
	var calls []*syntax.FuncCall
	variable, ok := syntax.UnwrapParens(h).(*syntax.Variable)
	if !ok || variable.Name == "" {
		return nil // A variable variable names no single handle.
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
