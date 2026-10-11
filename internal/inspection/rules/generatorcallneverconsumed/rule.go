// Package generatorcallneverconsumed implements the native GeneratorCallNeverConsumed inspection.
package generatorcallneverconsumed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Consume the generator to execute its body."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GeneratorCallNeverConsumed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	if _, ok := call.Parent().(*syntax.ExprStmt); !ok {
		return
	}
	if firstClassCallable(call) {
		return
	}
	body := semanticquery.NativeFunctionBody(ctx, call)
	if !semanticquery.NativeGeneratorBody(body) {
		return
	}
	if deferredEffects(body) {
		ctx.ReportNode(call, message)
	}
}

// firstClassCallable recognizes `f(...)`, which builds a Closure without
// calling the generator function.
func firstClassCallable(call *syntax.FuncCall) bool {
	for _, arg := range call.Args.Args {
		if _, ok := arg.(*syntax.VariadicPlaceholder); ok {
			return true
		}
	}
	return false
}

// deferredEffects reports whether the generator body owns an operation that
// is observable outside the generator. Writes to plain local variables only
// count when the body can alias a local to outer state.
func deferredEffects(body *syntax.Block) bool {
	effects, localWrite, aliased := false, false, false
	if fn, ok := body.Parent().(*syntax.Function); ok {
		for _, p := range fn.Params {
			aliased = aliased || p.ByRef
		}
	}
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New,
			*syntax.Echo, *syntax.Print, *syntax.InlineHTML, *syntax.Include, *syntax.Eval,
			*syntax.Exit, *syntax.Throw:
			effects = true
		case *syntax.Global, *syntax.StaticStmt:
			aliased = true
		case *syntax.Foreach:
			aliased = aliased || x.ByRef
		case *syntax.Assign:
			aliased = aliased || x.ByRef
			if localTarget(x.Var) {
				localWrite = true
			} else {
				effects = true
			}
		case *syntax.IncDec:
			if localTarget(x.Var) {
				localWrite = true
			} else {
				effects = true
			}
		}
		return !effects
	})
	return effects || localWrite && aliased
}

// localTarget recognizes a write whose root is a named function-local
// variable: `$x` or `$x[...]`. Property, static and dynamic targets are not.
func localTarget(target syntax.Expr) bool {
	for {
		dim, ok := target.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		target = dim.Var
	}
	v, ok := target.(*syntax.Variable)
	return ok && v.Name != "" && v.Name != "this" && v.Name != "GLOBALS"
}
