// Package regexreplacementcallbackfallsthrough implements the native RegexReplacementCallbackFallsThrough inspection.
package regexreplacementcallbackfallsthrough

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the replacement from the callback."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RegexReplacementCallbackFallsThrough" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "preg_replace_callback") {
		return
	}
	value := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "callback"))
	closure, ok := value.(*syntax.Closure)
	if !ok || closure.Body == nil || semanticquery.NativeGeneratorBody(closure.Body) {
		return
	}
	if !semanticquery.ExpansionDCallbackTerminates(ctx, closure.Body) && supported(closure.Body) {
		ctx.ReportNode(c, message)
	}
}

func supported(body syntax.Node) bool {
	ok := true
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) || n.Kind() == syntax.KClassLike {
			return false
		}
		switch n.(type) {
		case *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile, *syntax.Try, *syntax.Switch, *syntax.Goto:
			ok = false
		}
		return ok
	})
	return ok
}
