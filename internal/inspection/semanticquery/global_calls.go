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
	for _, n := range names {
		if g == n {
			return call, n
		}
	}
	return nil, ""
}
