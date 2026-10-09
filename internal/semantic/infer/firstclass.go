package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// firstClassCallableType describes acquisition of a callable, without treating
// the placeholder as call arguments. Return contracts remain useful even when
// the eventual arguments are unknown; argument-dependent refinements do not.
func (e *Env) firstClassCallableType(x syntax.Expr) types.Type {
	var ret types.Type
	switch n := x.(type) {
	case *syntax.FuncCall:
		if _, named := n.Name.(*syntax.Name); !named {
			ret = e.invokeType(e.TypeOf(n.Name))
		} else if f := e.ResolveFunction(n); f != nil {
			decls := e.Index.FunctionDecls(f.FQN, e.PHP)
			if len(decls) <= maxFuncDecls {
				call := *n
				call.Args = nil
				ts := []types.Type{e.declCallType(f, &call)}
				for _, g := range decls {
					if g != f {
						ts = append(ts, e.declCallType(g, &call))
					}
				}
				ret = types.Union(ts...)
			}
		}
	case *syntax.MethodCall:
		call := *n
		call.Args = nil
		ret = e.methodCallType(&call)
	case *syntax.StaticCall:
		call := *n
		call.Args = nil
		ret = e.staticCallType(&call)
	}
	return types.WithCallableReturn(types.Of(`\Closure`), `\Closure`, voidAsNull(ret))
}
