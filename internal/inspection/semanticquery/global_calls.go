package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// GlobalCall returns e as a plain function call that resolves to one of the
// global functions names (lower-case; compared case-insensitively, as PHP
// does, and only when the call reaches the global function rather than a
// same-named namespaced or imported one), plus that lower-case name.
func GlobalCall(ctx *analysis.Context, e syntax.Node, names ...string) (*syntax.FuncCall, string) {
	call, ok := e.(*syntax.FuncCall)
	if !ok {
		return nil, ""
	}
	g := ctx.GlobalFunctionName(call)
	if g == "" {
		return nil, ""
	}
	if declaration := ctx.Types().ResolveFunction(call); declaration != nil && !declaration.Builtin {
		return nil, ""
	}
	for _, n := range names {
		if g == n {
			return call, n
		}
	}
	return nil, ""
}

// NativeBuiltinName returns a resolved available builtin's canonical name.
// Unlike lexical name matching, this excludes application declarations and
// APIs unavailable at the selected language version.
func NativeBuiltinName(ctx *analysis.Context, call *syntax.FuncCall) string {
	name := ctx.GlobalFunctionName(call)
	declaration := ctx.Types().ResolveFunction(call)
	if declaration == nil || !declaration.Builtin || !declaration.Avail.In(ctx.PHP) {
		return ""
	}
	return name
}

// NativeBuiltin proves that a call targets the available named builtin.
func NativeBuiltin(ctx *analysis.Context, call *syntax.FuncCall, name string) bool {
	return NativeBuiltinName(ctx, call) == name
}
