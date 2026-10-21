// Package streamuseafterclose implements the native StreamUseAfterClose inspection.
package streamuseafterclose

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an open stream for this operation."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StreamUseAfterClose" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)

	switch name {
	case "fread", "fwrite", "fgets", "fseek", "ftell", "fflush", "ftruncate", "rewind":
	default:
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	if handle == nil {
		return
	}
	state, closed := ctx.Flow().StateBefore(c, handle, "fclose")
	if !closed {
		return
	}
	scope := syntax.EnclosingVariableScope(c)
	if streamCapturedByReference(ctx, scope, handle) || streamCloseFailed(ctx, state.Span, c, scope) {
		return
	}
	ctx.ReportNode(c, message)
}

// streamCapturedByReference reports whether a closure created in scope binds
// the handle variable by reference: calling it can store a freshly opened
// stream in the variable, which the flow state of the closed one cannot see.
func streamCapturedByReference(ctx *analysis.Context, scope syntax.Node, handle syntax.Expr) bool {
	variable, _ := syntax.UnwrapParens(handle).(*syntax.Variable)
	captured := false
	streamInspectScope(ctx, scope, func(n syntax.Node) bool {
		if n == scope {
			return true
		}
		if closure, ok := n.(*syntax.Closure); ok {
			for _, use := range closure.Uses {
				captured = captured || (variable != nil && use.ByRef && use.Var.Name == variable.Name)
			}
		}
		return !captured && !syntax.IsVariableScope(n)
	})
	return captured
}

// streamCloseFailed proves that use runs only when the fclose call spanning
// closed returned false, through the tests on that call's result guarding
// use. A refused close leaves the stream open.
func streamCloseFailed(ctx *analysis.Context, closed syntax.Span, use *syntax.FuncCall, scope syntax.Node) bool {
	if semanticquery.NativeLabelled(ctx, use) {
		return false
	}
	for _, reached := range ctx.Flow().Calls(scope) {
		call, ok := reached.Node.(*syntax.FuncCall)
		if !ok || call.Span() != closed {
			continue
		}
		if t := ctx.Types().Native().CallResultAt(call, use); !t.IsUnknown() && !semanticquery.MayHoldAtom(t, "true") {
			return true
		}
	}
	return false
}

// streamInspectScope walks the nodes of one variable scope, nested scopes
// included; nil denotes the file-level scope.
func streamInspectScope(ctx *analysis.Context, scope syntax.Node, visit func(syntax.Node) bool) {
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
		return
	}
	syntax.Inspect(scope, visit)
}
