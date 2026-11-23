package infer

import (
	"slices"

	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

// pipeType invokes a callable with the left operand. Detached shallow call
// copies retain lexical context without changing the source tree or its cache.
func (e *Env) pipeType(input, callback syntax.Expr) types.Type {
	switch c := syntax.UnwrapParens(callback).(type) {
	case *syntax.FuncCall:
		if args := pipeArgs(c.Args, input); args != nil {
			call := *c
			call.Args = args
			return voidAsNull(e.funcCallType(&call))
		}
	case *syntax.MethodCall:
		if args := pipeArgs(c.Args, input); args != nil {
			call := *c
			call.Args = args
			return voidAsNull(e.methodCallType(&call))
		}
	case *syntax.StaticCall:
		if args := pipeArgs(c.Args, input); args != nil {
			call := *c
			call.Args = args
			return voidAsNull(e.staticCallType(&call))
		}
	case *syntax.Clone:
		// `$o |> clone(?)`, `$o |> clone(?, $with)`: the input is cloned.
		if pipeArgs(c.Args, input) != nil && clonesInput(c.Args) {
			return e.TypeOf(input)
		}
	}
	return voidAsNull(e.callbackReturn(callback))
}

// clonesInput reports whether the input piped into clone() is the object
// it clones: through `clone(...)`, or a `?` in the first position or named
// object.
func clonesInput(list *syntax.ArgList) bool {
	i := slices.IndexFunc(list.Args, func(x syntax.Expr) bool { _, ok := x.(*syntax.ArgPlaceholder); return ok })
	if i < 0 {
		return true // clone(...)
	}
	hole := list.Args[i].(*syntax.ArgPlaceholder)
	return hole.Name == nil && i == 0 || hole.Name != nil && hole.Name.Value == "object"
}

// pipeArgs returns the arguments a callback written as a call receives from
// a pipe: the input alone for `f(...)`, or, for a partial application with
// one `?` (`f(?, 2)`, `f(2, x: ?)`, PHP 8.6), the written arguments with the
// input in its place. Any other list is nil.
func pipeArgs(list *syntax.ArgList, input syntax.Expr) *syntax.ArgList {
	if list.FirstClassCallable() {
		return &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: input}}}
	}
	if !list.Partial() {
		return nil
	}
	args := make([]syntax.Expr, 0, len(list.Args))
	holes := 0
	for _, x := range list.Args {
		switch a := x.(type) {
		case *syntax.ArgPlaceholder:
			holes++
			args = append(args, &syntax.Arg{Name: a.Name, Value: input})
		case *syntax.Arg:
			args = append(args, a)
		}
	}
	if holes != 1 {
		return nil
	}
	return &syntax.ArgList{Args: args}
}
