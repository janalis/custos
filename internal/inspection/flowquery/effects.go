package flowquery

import (
	"custos/internal/php/syntax"
)

// MayHaveSideEffects reports whether evaluating e may do more than compute a
// value: calls (function, method, static, `|>`), object creation, clone,
// assignments, increments, include/eval/exit/print/throw/yield and shell
// exec. Bodies of closures and arrow functions are not evaluated where they
// are written, so they are not inspected.
func MayHaveSideEffects(e syntax.Node) bool {
	if e == nil {
		return false
	}
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		if found {
			return false
		}
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New,
			*syntax.Clone, *syntax.Assign, *syntax.IncDec, *syntax.Include, *syntax.Eval,
			*syntax.Exit, *syntax.Print, *syntax.Throw, *syntax.Yield, *syntax.YieldFrom:
			found = true
		case *syntax.Binary:
			found = x.Op.Kind == syntax.TPipe
		case *syntax.InterpolatedString:
			found = x.Backtick
		}
		return !found
	})
	return found
}
