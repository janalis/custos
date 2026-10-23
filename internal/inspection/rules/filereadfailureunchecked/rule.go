// Package filereadfailureunchecked implements the native FileReadFailureUnchecked inspection.
package filereadfailureunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/phpunit"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Handle file reading failure before using its result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "FileReadFailureUnchecked" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KStaticCall, syntax.KMethodCall, syntax.KBinary}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var inputs []syntax.Expr
	var args *syntax.ArgList
	switch x := n.(type) {
	case *syntax.FuncCall:
		args = x.Args
	case *syntax.StaticCall:
		args = x.Args
	case *syntax.MethodCall:
		args = x.Args
	case *syntax.Binary:
		if x.Op.Kind != syntax.TDot {
			return
		}
		inputs = []syntax.Expr{x.Left, x.Right}
	}
	assertion := args != nil && phpunit.PuAssertion(ctx, n)
	library, checked := phpunit.LibraryFailurePassing(ctx, n, args, true)
	switch {
	case assertion:
		inputs = phpunit.FailurePassing(ctx, n, args, true)
	case checked:
		inputs, assertion = library, true
	case args != nil:
		inputs = semanticquery.NativeStringInputs(ctx, n, args)
	}
	for _, input := range inputs {
		read, ok := semanticquery.NativeLocalBuiltinResult(ctx, input, "false", "file_get_contents", "stream_get_contents").(*syntax.FuncCall)
		if !ok || infallibleRead(ctx, read) {
			continue
		}

		// A narrowed type describes a static local as its last test left it:
		// only the flow knows the calls that may have assigned it since. In
		// a scope with a `goto` label only the flow says what holds here.
		jumps := semanticquery.NativeFailureFlowOnly(ctx, input)
		if !jumps && !semanticquery.NativeStaticLocal(ctx, input) && ctx.TypeOf(input).OnlyOf("string") || ctx.Flow().Excludes(input, "false") || !jumps && semanticquery.NativeNarrowedExcludes(ctx, input, types.Of("string", "false"), "false") {
			continue
		}
		// A parameter must be `string` in whichever declaration runs; a
		// PHPUnit assertion checks what it is given, unless it lets the
		// failure pass as text.
		if args == nil || assertion || semanticquery.NativeStringConsumer(ctx, n, args, input) {
			ctx.ReportNode(n, message)
			return
		}
	}
}

// infallibleRead reports whether read is stream_get_contents() of a handle
// opened on one of PHP's own memory or standard streams (E7), which
// cannot fail.
func infallibleRead(ctx *analysis.Context, read *syntax.FuncCall) bool {
	if semanticquery.NativeBuiltinName(ctx, read) != "stream_get_contents" {
		return false
	}
	open, ok := semanticquery.NativeLocalValue(ctx, semanticquery.CallArgument(read.Args, 0, "stream")).(*syntax.FuncCall)
	return ok && semanticquery.NativeBuiltinName(ctx, open) == "fopen" && semanticquery.InfallibleOpen(open)
}
