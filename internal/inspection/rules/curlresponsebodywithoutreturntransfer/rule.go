// Package curlresponsebodywithoutreturntransfer implements the native CurlResponseBodyWithoutReturnTransfer inspection.
package curlresponsebodywithoutreturntransfer

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable response-body return before decoding curl output."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlResponseBodyWithoutReturnTransfer" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	for _, input := range semanticquery.NativeStringInputs(ctx, call, call.Args) {
		producer, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, producer, "curl_exec") {
			continue
		}
		handle := semanticquery.CallArgument(producer.Args, 0, "handle")
		origin, known := ctx.Flow().Resolve(handle)
		if !known {
			continue
		}
		init, ok := origin.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, init, "curl_init") {
			continue
		}
		if !semanticquery.NativeCurlDefault(ctx, producer, handle, "CURLOPT_RETURNTRANSFER") {
			continue
		}
		if sharedOutsideScope(ctx, producer, handle) {
			continue
		}
		ctx.ReportNode(call, message)
		return
	}
}

// sharedOutsideScope reports whether another variable scope can reach the
// handle: a closure capture, an arrow-function reference or a global binding.
// Code in that scope may configure the handle without passing it to a call
// the local proof sees, so the option state is unknown.
func sharedOutsideScope(ctx *analysis.Context, at syntax.Node, handle syntax.Expr) bool {
	id := ctx.Flow().Value(handle).Identity
	scope := syntax.EnclosingVariableScope(at)
	names := map[string]bool{}
	visit := func(fn func(syntax.Node) bool) {
		if scope != nil {
			syntax.Inspect(scope, fn)
			return
		}
		syntax.InspectFile(ctx.File, fn)
	}
	// Every local name holding the handle, so captured aliases count too.
	visit(func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		switch n := n.(type) {
		case *syntax.Variable:
			if n.NameExpr == nil && ctx.Flow().Value(n).Identity == id {
				names[n.Name] = true
			}
		case *syntax.Assign:
			if v, ok := n.Var.(*syntax.Variable); ok && ctx.Flow().Value(n.Value).Identity == id {
				names[v.Name] = true
			}
		}
		return true
	})
	shared := false
	visit(func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Closure:
			for _, use := range n.Uses {
				shared = shared || n != scope && names[use.Var.Name]
			}
		case *syntax.ArrowFunction:
			for name := range names {
				shared = shared || n != scope && astquery.MentionsVariable(n.Expr, name)
			}
		case *syntax.Global:
			for _, v := range n.Vars {
				if v, ok := v.(*syntax.Variable); ok && names[v.Name] {
					shared = true
				}
			}
		}
		return !shared
	})
	return shared
}
