package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// arrayAccessDimType resolves an object's offset read through ArrayAccess's
// offsetGet contract. As for array and string reads, null/false alternatives
// are ignored; every remaining union alternative must support offset reads.
func (e *Env) arrayAccessDimType(n *syntax.ArrayDimFetch, receiver types.Type) types.Type {
	if n.Dim == nil {
		return types.Unknown
	}
	receiver = receiver.Without("null", "false")
	classes := receiver.Classes()
	if len(classes) == 0 || len(classes) != len(receiver.Atoms()) {
		return types.Unknown
	}
	classes = e.memberClasses(receiver, func(cls string) bool {
		return e.Index.IsSubtype(cls, "ArrayAccess", e.PHP)
	})
	if len(classes) == 0 {
		return types.Unknown
	}
	call := &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: n.Dim}}}
	var results []types.Type
	for _, cls := range classes {
		if !e.Index.IsSubtype(cls, "ArrayAccess", e.PHP) {
			return types.Unknown
		}
		name := strings.TrimPrefix(cls, `\`)
		result := e.methodReturn(name, "offsetGet", true, name, receiver.TypeArgs(cls), call)
		if result.IsUnknown() {
			return types.Unknown
		}
		results = append(results, result)
	}
	return types.Union(results...)
}
