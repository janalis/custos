package passingbyreferencecorrectness

import (
	"custos/internal/php/syntax"
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
