// Package pregreplacefailureunchecked implements the native PregReplaceFailureUnchecked inspection.
package pregreplacefailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/phpunit"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle regex replacement failure before using its result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "PregReplaceFailureUnchecked" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KStaticCall, syntax.KMethodCall}
}

func (rule) Check(ctx *analysis.Context, call syntax.Node) {
	var args *syntax.ArgList
	switch x := call.(type) {
	case *syntax.FuncCall:
		args = x.Args
	case *syntax.StaticCall:
		args = x.Args
	case *syntax.MethodCall:
		args = x.Args
	}
	for _, input := range semanticquery.NativeStringInputs(ctx, call, args) {
		inner, ok := semanticquery.NativeLocalBuiltinResult(ctx, input, "null", "preg_replace", "preg_replace_callback").(*syntax.FuncCall)
		if !ok {
			continue
		}
		switch semanticquery.NativeBuiltinName(ctx, inner) {
		case "preg_replace", "preg_replace_callback":
		default:
			continue
		}
		// In a scope with a `goto` label only the flow says what holds here.
		if ctx.Flow().Excludes(input, "null") || semanticquery.NativeStringNonNullGuard(ctx, input) || !semanticquery.NativeFailureFlowOnly(ctx, input) && (ctx.TypeOf(input).OnlyOf("string") || nullGuarded(ctx, input)) {
			continue
		}
		// The parameter must be `string` in whichever declaration runs, and a
		// PHPUnit assertion is the check of what it is given.
		if !phpunit.PuAssertion(ctx, call) && semanticquery.NativeStringConsumer(ctx, call, args, input) {
			ctx.ReportNode(call, message)
			return
		}
	}
}
