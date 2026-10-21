package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// GlobalStateCall returns the call that last set the global state domain
// (such as "response" or "session") before at, in the scope of at; nil
// when that is not known or not a call there.
func GlobalStateCall(ctx *analysis.Context, at syntax.Node, domain string) *syntax.FuncCall {
	state, known := ctx.Flow().GlobalStateBefore(at, domain)
	if !known {
		return nil
	}
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		if call, ok := c.Node.(*syntax.FuncCall); ok && call.Span() == state.Span {
			return call
		}
	}
	return nil
}

// CallsAnnotatedFunction reports whether c calls a project function whose
// doc comment carries annotation (such as @custos-credential-store).
func CallsAnnotatedFunction(ctx *analysis.Context, c *syntax.FuncCall, annotation string) bool {
	function := ctx.Types().ResolveFunction(c)
	if function == nil || function.Builtin {
		return false
	}
	decl := FunctionDecl(ctx.File, function)
	if decl == nil {
		return false
	}
	for _, field := range strings.Fields(index.DocComment(ctx.File, decl)) {
		if strings.Trim(field, "/*") == annotation {
			return true
		}
	}
	return false
}

// OnlyFalse accepts a narrowed result that can only be false.
func OnlyFalse(t types.Type) bool { return t.OnlyOf("false") }

// ArgumentsRemovable reports whether every argument of call can be
// removed without losing an observable evaluation: none is unpacked, passed
// by reference or has side effects.
func ArgumentsRemovable(call *syntax.FuncCall) bool {
	for _, node := range call.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.ByRef || flowquery.MayHaveSideEffects(arg.Value) {
			return false
		}
	}
	return true
}
