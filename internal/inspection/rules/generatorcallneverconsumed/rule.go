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
	body := semanticquery.NativeFunctionBody(ctx, call)
	if !semanticquery.NativeGeneratorBody(body) {
		return
	}
	effects := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch x.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.Assign, *syntax.IncDec:
			effects = true
		}
		return !effects
	})
	if effects {
		ctx.ReportNode(call, message)
	}
}
