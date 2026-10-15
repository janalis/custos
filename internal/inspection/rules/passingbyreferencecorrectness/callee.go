package passingbyreferencecorrectness

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// callArgs returns the argument list of a function, method or static call.
func callArgs(call syntax.Node) *syntax.ArgList {
	switch c := call.(type) {
	case *syntax.FuncCall:
		return c.Args
	case *syntax.MethodCall:
		return c.Args
	}
	return call.(*syntax.StaticCall).Args
}

// declaredParams collapses consecutive duplicate names in permissive stub
// signatures. ResolveCallee already selects parameter availability by version.
func declaredParams(params []index.Param) []index.Param {
	out := make([]index.Param, 0, len(params))
	for _, p := range params {
		if n := len(out); n > 0 && out[n-1].Name == p.Name {
			continue
		}
		out = append(out, p)
	}
	return out
}
