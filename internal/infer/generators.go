package infer

import (
	"custos/internal/syntax"
	"custos/internal/types"
)

// generatorType keeps generator identity even when yielded types are unknown.
// Returns inside a generator describe its completion value, not its call result.
func (e *Env) generatorType(body syntax.Node) (types.Type, bool) {
	var yields []syntax.Node
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Yield, *syntax.YieldFrom:
			yields = append(yields, n)
		}
		return true
	})
	if len(yields) == 0 {
		return types.Unknown, false
	}
	var keys, values []types.Type
	for _, n := range yields {
		var k, v types.Type
		switch n := n.(type) {
		case *syntax.Yield:
			k, v = types.Int, types.Null
			if n.Key != nil {
				k = e.TypeOf(n.Key)
			}
			if n.Value != nil {
				v = e.TypeOf(n.Value)
			}
		case *syntax.YieldFrom:
			t := e.TypeOf(n.Expr)
			if t.IsArrayLike() {
				k, v = types.Of("int", "string"), t.Elem()
				if t.IsSealedShape() {
					if len(t.ShapeKeys()) == 0 {
						continue // an empty delegation contributes no keys or values
					}
					var ks, vs []types.Type
					for _, key := range t.ShapeKeys() {
						kt := types.String
						if types.IsIntKey(key.Name) {
							kt = types.Int
						}
						ks, vs = append(ks, kt), append(vs, key.Type)
					}
					k, v = types.Union(ks...), types.Union(vs...)
				}
			} else {
				k, v = e.iterTypes(t)
			}
		}
		keys, values = append(keys, k), append(values, v)
	}
	join := func(ts []types.Type) types.Type {
		for _, t := range ts {
			if t.IsUnknown() || t.Has("mixed") {
				return types.Mixed
			}
		}
		if len(ts) == 0 {
			return types.Mixed
		}
		return types.Union(ts...)
	}
	return types.Of(`\Generator`).WithTypeArgs(`\Generator`, []types.Type{join(keys), join(values)}), true
}
