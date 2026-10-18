package semanticquery

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

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
		byRef, _ := referenceBinding(arg, pos, params)
		if byRef {
			return true
		}
		pos = nextArgumentPosition(arg, pos)
	}
	return false
}

// ReferenceArgument is a written argument that may bind by reference.
// Known distinguishes an exact binding from an uncertain unpacked position.
type ReferenceArgument struct {
	Argument *syntax.Arg
	Known    bool
}

// ReferenceArguments collects possible reference bindings in one pass over the
// written arguments. Callers must not use uncertain bindings as write proofs.
func ReferenceArguments(list *syntax.ArgList, params []index.Param, callTimeRef bool) []ReferenceArgument {
	if list == nil {
		return nil
	}
	var out []ReferenceArgument
	pos := 0
	for _, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok {
			continue
		}
		byRef, known := referenceBinding(arg, pos, params)
		if callTimeRef && arg.ByRef {
			byRef, known = true, true
		}
		if byRef {
			out = append(out, ReferenceArgument{Argument: arg, Known: known})
		}
		pos = nextArgumentPosition(arg, pos)
	}
	return out
}

// ArgumentBindsByRef reports whether one written argument may bind by
// reference. known=false means unpacking obscures its parameter; byRef then
// conservatively includes every parameter it may bind. A node not in list
// has no established binding.
func ArgumentBindsByRef(list *syntax.ArgList, target *syntax.Arg, params []index.Param, callTimeRef bool) (byRef, known bool) {
	if list != nil && target != nil {
		pos := 0
		for _, node := range list.Args {
			arg, ok := node.(*syntax.Arg)
			if !ok {
				continue
			}
			if arg == target {
				if callTimeRef && arg.ByRef {
					return true, true
				}
				return referenceBinding(arg, pos, params)
			}
			pos = nextArgumentPosition(arg, pos)
		}
	}
	return false, false
}

func nextArgumentPosition(arg *syntax.Arg, pos int) int {
	if arg.Unpack {
		return -1
	}
	if arg.Name == nil && pos >= 0 {
		return pos + 1
	}
	return pos
}

func referenceBinding(arg *syntax.Arg, pos int, params []index.Param) (bool, bool) {
	if arg.Unpack || arg.Name == nil && pos < 0 {
		for i, p := range params {
			if p.ByRef && (pos < 0 || i >= pos || p.Variadic) {
				return true, false
			}
		}
		return false, false
	}
	if arg.Name != nil {
		for _, p := range params {
			if strings.TrimPrefix(p.Name, "$") == arg.Name.Value {
				return p.ByRef, true
			}
		}
	} else if pos < len(params) {
		return params[pos].ByRef, true
	}
	if len(params) > 0 && params[len(params)-1].Variadic {
		return params[len(params)-1].ByRef, true
	}
	return false, arg.Name == nil
}
