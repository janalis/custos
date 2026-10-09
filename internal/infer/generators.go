package infer

import (
	"strings"

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
	return types.Of(`\Generator`).WithTypeArgs(`\Generator`, []types.Type{join(keys), join(values), types.Mixed, e.generatorCompletion(body)}), true
}

// generatorCompletion describes the value made available after iteration ends.
func (e *Env) generatorCompletion(body syntax.Node) types.Type {
	if expr, ok := body.(syntax.Expr); ok {
		t := e.TypeOf(expr)
		if t.IsUnknown() {
			return types.Mixed
		}
		return t
	}
	block := body.(*syntax.Block)
	var ts []types.Type
	unknown := false
	var walk func(syntax.Node)
	walk = func(n syntax.Node) {
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return
		case *syntax.Block:
			for _, stmt := range x.Stmts {
				walk(stmt)
				if syntax.Terminates(stmt) {
					break
				}
			}
			return
		case *syntax.Return:
			t := types.Null
			if x.Expr != nil {
				t = e.TypeOf(x.Expr)
			}
			if t.IsUnknown() || t.Has("mixed") {
				unknown = true
			}
			ts = append(ts, t)
			return
		}
		syntax.Children(n, walk)
	}
	walk(block)
	if unknown {
		return types.Mixed
	}
	if !syntax.Terminates(block) {
		ts = append(ts, types.Null)
	}
	if len(ts) == 0 {
		return types.Of("never")
	}
	return types.Union(ts...)
}

// yieldType uses the documented send contract of the nearest function scope.
func (e *Env) yieldType(n *syntax.Yield) types.Type {
	if e.native {
		return types.Unknown
	}
	scope := syntax.EnclosingVariableScope(n)
	if scope == nil {
		return types.Unknown
	}
	d := e.DocOf(scope)
	if d == nil {
		return types.Unknown
	}
	t := types.FromDoc(d.ReturnType(), e.resolverFor(scope, scope.Span().Start))
	args := t.TypeArgs(`\Generator`)
	if len(args) < 3 || args[2].IsUnknown() || args[2].Has("mixed") {
		return types.Unknown
	}
	return types.Union(args[2], types.Null)
}

// yieldFromType is delegation's completion value, rather than its yielded value.
func (e *Env) yieldFromType(n *syntax.YieldFrom) types.Type {
	t := e.TypeOf(n.Expr)
	if t.IsUnknown() {
		return types.Unknown
	}
	var ts []types.Type
	for _, a := range t.Atoms() {
		switch {
		case strings.EqualFold(a, `\Generator`):
			args := t.TypeArgs(a)
			if len(args) < 4 || args[3].IsUnknown() || args[3].Has("mixed") {
				return types.Unknown
			}
			ts = append(ts, args[3])
		case a == "array" || strings.HasSuffix(a, "[]"):
			ts = append(ts, types.Null)
		case strings.HasPrefix(a, `\`):
			cls := strings.TrimPrefix(a, `\`)
			if strings.EqualFold(cls, "Traversable") || strings.EqualFold(cls, "Iterator") || strings.EqualFold(cls, "IteratorAggregate") {
				// Interface values can also contain generators whose completion is unknown.
				return types.Unknown
			}
			traversable := false
			for _, c := range e.Index.Ancestors(cls, e.PHP) {
				if strings.EqualFold(c.FQN, "Traversable") {
					traversable = true
				}
			}
			if !traversable {
				return types.Unknown
			}
			ts = append(ts, types.Null)
		default:
			return types.Unknown
		}
	}
	return types.Union(ts...)
}
