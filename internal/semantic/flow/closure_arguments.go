package flow

import "custos/internal/php/syntax"

func (e *Env) closureDefault(expr syntax.Expr, owner *scopeResult) (Value, bool) {
	fuel := 256
	if !primitiveDefault(expr, 8, &fuel) {
		return Value{}, false
	}
	r := &scopeResult{sharing: e.sharingArena(), complete: true, values: newValueStore(256 - fuel), before: map[syntax.NodeIdentity]map[uint32]State{}}
	f := frame{definite: true}
	f.resetVars()
	value := e.eval(expr, &f, r, 0)
	value.elements = e.elements(expr, &f, r)
	owner.operations += r.operations
	return value, value.Complete && owner.operations <= MaxTransfers
}

func primitiveDefault(expr syntax.Expr, depth int, fuel *int) bool {
	*fuel--
	if depth == 0 || *fuel < 0 {
		return false
	}
	switch n := syntax.UnwrapParens(expr).(type) {
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return constantName(n) == "null" || constantName(n) == "true" || constantName(n) == "false"
	case *syntax.Unary:
		_, known := literalInteger(n)
		return known
	case *syntax.Array:
		for _, item := range n.Items {
			if item == nil || item.ByRef || item.Unpack || !primitiveDefault(item.Value, depth-1, fuel) {
				return false
			}
			if item.Key != nil {
				if _, known := syntax.LiteralArrayKey(item.Key); !known {
					return false
				}
			}
		}
		return true
	}
	return false
}
