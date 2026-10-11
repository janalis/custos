package passingbyreferencecorrectness

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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

// declaredParams returns the parameter list the callee has at the analysed
// PHP level. Stub declarations spell a parameter whose signature changed
// between versions once per variant: only the variant available at ver is
// kept, and variants carrying no version range (same name repeated in a
// row) collapse into one, so list positions match call positions.
func declaredParams(params []index.Param, ver phpversion.Version) []index.Param {
	out := make([]index.Param, 0, len(params))
	for _, p := range params {
		if !p.Avail.In(ver) {
			continue
		}
		if n := len(out); n > 0 && out[n-1].Name == p.Name {
			continue
		}
		out = append(out, p)
	}
	return out
}
