package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// callbackSlot is the parameter position of the callback each builtin with a
// fixed callback contract takes.
var callbackSlot = map[string]int{
	"array_map": 0, "array_filter": 1, "array_walk": 1, "array_reduce": 1,
	"usort": 1, "uasort": 1, "uksort": 1,
}

// contextParamType types parameter i of closure or arrow function fn when fn
// is written directly as the callback of a builtin that decides what it
// passes: the elements (or keys) of the array argument for array_map(),
// array_filter() (by its mode), array_walk(), array_reduce() and the usort()
// family. Unknown otherwise.
func (e *Env) contextParamType(fn syntax.Node, i int) types.Type {
	arg, ok := fn.Parent().(*syntax.Arg)
	if !ok {
		return types.Unknown
	}
	call, ok := arg.Parent().Parent().(*syntax.FuncCall)
	if !ok {
		return types.Unknown
	}
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return types.Unknown
	}
	fqn, fb := e.Names.Function(name.Value, name.Span().Start)
	if fb != "" && e.Index.Function(fqn, e.PHP) == nil {
		if e.annotating {
			e.deps = append(e.deps, fqn)
		}
		fqn = fb
	}
	builtin := strings.ToLower(fqn)
	slot, ok := callbackSlot[builtin]
	function := e.Index.Function(fqn, e.PHP)
	if !ok || function == nil || !function.Builtin || !builtinBindingsKnown(call.Args, function.Params, e.PHP, false) {
		return types.Unknown
	}
	at := func(j int) syntax.Expr {
		x, state := bindArgument(call.Args, function.Params, j, e.PHP)
		if state != bindingBound {
			return nil
		}
		return x
	}
	if at(slot) != fn {
		return types.Unknown
	}
	switch builtin {
	case "array_map":
		// array_map(cb, $a, $b, …): parameter i receives the elements of
		// the (i+1)th array, padded with null when there are several.
		inputs := len(call.Args.Args) - 1
		if inputs == 1 {
			if i == 0 {
				return e.foreachElem(at(1))
			}
			return types.Unknown
		}
		if i >= inputs || namedArgs(call.Args) {
			return types.Unknown
		}
		if el := e.foreachElem(call.Args.Args[i+1].(*syntax.Arg).Value); !el.IsUnknown() {
			return types.Union(el, types.Null)
		}
	case "array_filter":
		arr := at(0)
		mode := ""
		if m := at(2); m != nil {
			c, ok := syntax.UnwrapParens(m).(*syntax.ConstFetch)
			if !ok {
				return types.Unknown
			}
			mode = strings.ToUpper(strings.TrimPrefix(c.Name.Value, `\`))
		}
		switch {
		case i == 0 && mode == "":
			return e.foreachElem(arr)
		case i == 0 && mode == "ARRAY_FILTER_USE_KEY":
			return e.foreachKey(arr)
		case mode == "ARRAY_FILTER_USE_BOTH":
			if i == 0 {
				return e.foreachElem(arr)
			}
			if i == 1 {
				return e.foreachKey(arr)
			}
		}
	case "array_walk":
		switch i {
		case 0:
			return e.foreachElem(at(0))
		case 1:
			return e.foreachKey(at(0))
		}
	case "array_reduce":
		if i == 1 {
			return e.foreachElem(at(0))
		}
	case "usort", "uasort":
		if i < 2 {
			return e.foreachElem(at(0))
		}
	case "uksort":
		if i < 2 {
			return e.foreachKey(at(0))
		}
	}
	return types.Unknown
}

// namedArgs reports a list with a named argument.
func namedArgs(args *syntax.ArgList) bool {
	for _, x := range args.Args {
		if x.(*syntax.Arg).Name != nil {
			return true
		}
	}
	return false
}
