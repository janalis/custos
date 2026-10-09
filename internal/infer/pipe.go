package infer

import (
	"custos/internal/syntax"
	"custos/internal/types"
)

// pipeType invokes a callable with the left operand. Detached shallow call
// copies retain lexical context without changing the source tree or its cache.
func (e *Env) pipeType(input, callback syntax.Expr) types.Type {
	args := &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: input}}}
	switch c := syntax.UnwrapParens(callback).(type) {
	case *syntax.FuncCall:
		if isFirstClassCallable(c.Args) {
			call := *c
			call.Args = args
			return voidAsNull(e.funcCallType(&call))
		}
	case *syntax.MethodCall:
		if isFirstClassCallable(c.Args) {
			call := *c
			call.Args = args
			return voidAsNull(e.methodCallType(&call))
		}
	case *syntax.StaticCall:
		if isFirstClassCallable(c.Args) {
			call := *c
			call.Args = args
			return voidAsNull(e.staticCallType(&call))
		}
	}
	return voidAsNull(e.callbackReturn(callback))
}
