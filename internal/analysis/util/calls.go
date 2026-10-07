package util

import (
	"strings"

	"custos/internal/index"
	"custos/internal/syntax"
)

// CallName splits the name of a plain function call as written into its
// namespace qualifier (“, `\`, `Ns\`, `namespace\`) and last segment.
// ok is false for dynamic calls (`$f()`, `(expr)()`).
func CallName(call *syntax.FuncCall) (qualifier, name string, ok bool) {
	n, isName := call.Name.(*syntax.Name)
	if !isName {
		return "", "", false
	}
	v := n.Value
	i := strings.LastIndexByte(v, '\\')
	return v[:i+1], v[i+1:], true
}

// CallLastName returns the last segment of a plain function call's name as
// written (case preserved), or "" for dynamic calls and non-calls.
func CallLastName(e syntax.Expr) string {
	call, ok := e.(*syntax.FuncCall)
	if !ok {
		return ""
	}
	_, name, _ := CallName(call)
	return name
}

// CallArgValues returns the argument values of call, in order. ok is false when
// the call has no argument list or uses spreads, named arguments or the
// first-class callable placeholder `f(...)`.
func CallArgValues(call *syntax.FuncCall) (args []syntax.Expr, ok bool) {
	if call.Args == nil {
		return nil, false
	}
	return ArgValues(call.Args)
}

// ArgValues returns the values of an argument list (see CallArgValues).
func ArgValues(list *syntax.ArgList) ([]syntax.Expr, bool) {
	out := make([]syntax.Expr, 0, len(list.Args))
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Name != nil || arg.Value == nil {
			return nil, false
		}
		out = append(out, arg.Value)
	}
	return out, true
}

// ParentFuncCall returns the plain function call that has e directly (no
// parentheses in between) as one of its arguments, or nil.
func ParentFuncCall(e syntax.Node) *syntax.FuncCall {
	arg, ok := e.Parent().(*syntax.Arg)
	if !ok || arg.Value != e {
		return nil
	}
	list, ok := arg.Parent().(*syntax.ArgList)
	if !ok {
		return nil
	}
	call, _ := list.Parent().(*syntax.FuncCall)
	return call
}

// BoolConst reports whether e is the constant `true` or `false`
// (case-insensitive, optionally `\`-qualified) and which one.
func BoolConst(e syntax.Node) (value, ok bool) {
	c, isConst := e.(*syntax.ConstFetch)
	if !isConst || c.Name == nil {
		return false, false
	}
	v := strings.TrimPrefix(c.Name.Value, `\`)
	switch {
	case strings.EqualFold(v, "true"):
		return true, true
	case strings.EqualFold(v, "false"):
		return false, true
	}
	return false, false
}

// ArgBindsByRef reports whether an argument of list binds to a by-reference
// parameter of params (positional, variadic tail or named). With
// callTimeRef, a call-time `&$x` argument counts too.
func ArgBindsByRef(list *syntax.ArgList, params []index.Param, callTimeRef bool) bool {
	if list == nil {
		return false
	}
	pos := 0
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			continue
		}
		if callTimeRef && arg.ByRef {
			return true
		}
		if arg.Name != nil {
			for _, p := range params {
				if p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
					return true
				}
			}
			continue
		}
		if pos < len(params) && params[pos].ByRef {
			return true
		}
		if n := len(params); pos >= n && n > 0 && params[n-1].Variadic && params[n-1].ByRef {
			return true
		}
		pos++
	}
	return false
}
