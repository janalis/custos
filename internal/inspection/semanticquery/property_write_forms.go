package semanticquery

import (
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/types"
)

// Property writes other than a plain or compound assignment: destructuring,
// foreach targets, ++/-- and by-reference assignments. Each yields
// propertyStore values judged by judgeStore.

// storeTargets calls visit with every target node n writes: the target of
// an assignment, each destructured item, the key and value of a foreach
// loop, the operand of ++/--.
func storeTargets(n syntax.Node, visit func(syntax.Expr)) {
	switch n := n.(type) {
	case *syntax.Assign:
		destructured(n.Var, visit)
	case *syntax.Foreach:
		if n.Key != nil {
			visit(n.Key)
		}
		destructured(n.Value, visit)
	case *syntax.IncDec:
		visit(n.Var)
	}
}

// destructured calls visit with x, or with each item target of the
// destructuring pattern x, nested patterns included.
func destructured(x syntax.Expr, visit func(syntax.Expr)) {
	items, ok := patternItems(x)
	if !ok {
		visit(x)
		return
	}
	for _, it := range items {
		if it != nil && it.Value != nil {
			destructured(it.Value, visit)
		}
	}
}

// patternItems returns the items of destructuring pattern x (`[…]` or
// `list(…)`).
func patternItems(x syntax.Expr) ([]*syntax.ArrayItem, bool) {
	switch p := x.(type) {
	case *syntax.Array:
		return p.Items, true
	case *syntax.List:
		return p.Items, true
	}
	return nil, false
}

// destructuringMismatches reports the properties `[$this->a, $this->b] =
// $pair` stores into.
func destructuringMismatches(ctx *analysis.Context, a *syntax.Assign) []PropertyMismatch {
	if _, ok := patternItems(a.Var); !ok || a.Value == nil || reboundThis(a) {
		return nil
	}
	return patternMismatches(ctx, a, a.Var, func(env *infer.Env) types.Type { return env.TypeOf(a.Value) })
}

// foreachMismatches reports the properties loop fe binds its key and value
// to, the value possibly destructured.
func foreachMismatches(ctx *analysis.Context, fe *syntax.Foreach) []PropertyMismatch {
	if reboundThis(fe.Expr) {
		return nil
	}
	var out []PropertyMismatch
	if fe.Key != nil && propertyTarget(fe.Key) && !reboundThis(fe.Key) {
		w := propertyStore{target: fe.Key, at: fe, report: fe.Key}
		if m, ok := judgeStore(ctx, w, func(env *infer.Env) types.Type { return env.ForeachKeyType(fe) }); ok {
			out = append(out, m)
		}
	}
	if fe.ByRef {
		return out
	}
	return append(out, patternMismatches(ctx, fe, fe.Value, func(env *infer.Env) types.Type { return env.ForeachValueType(fe) })...)
}

// patternMismatches reports the property targets of x, a target or a
// destructuring pattern, that writing node at stores into from a value
// typed by src.
func patternMismatches(ctx *analysis.Context, at syntax.Node, x syntax.Expr, src func(*infer.Env) types.Type) []PropertyMismatch {
	items, ok := patternItems(x)
	if !ok {
		if !propertyTarget(x) || reboundThis(x) {
			return nil
		}
		if m, ok := judgeStore(ctx, propertyStore{target: x, at: at, report: x}, src); ok {
			return []PropertyMismatch{m}
		}
		return nil
	}
	var out []PropertyMismatch
	for i, it := range items {
		if it == nil || it.Value == nil || it.ByRef {
			continue
		}
		key, ok := strconv.Itoa(i), it.Key == nil
		if it.Key != nil {
			key, ok = syntax.LiteralArrayKey(it.Key)
		}
		if !ok {
			continue
		}
		item := func(env *infer.Env) types.Type { return destructuredItem(src(env), key) }
		out = append(out, patternMismatches(ctx, at, it.Value, item)...)
	}
	return out
}

// destructuredItem is the value destructuring stores from key of a value of
// type src: the shape's type for the key, the element type of `T[]`
// members, null from a scalar or null (PHP reads no key from them) and,
// unless a required shape key, null for a missing key. Unknown for objects
// (ArrayAccess runs code), plain `array` and keys a shape does not list.
func destructuredItem(src types.Type, key string) types.Type {
	if src.IsUnknown() {
		return types.Unknown
	}
	var arr, scalar []string
	for _, a := range src.Atoms() {
		switch a {
		case "null", "bool", "true", "false", "int", "float", "string":
			scalar = append(scalar, a)
		default:
			arr = append(arr, a)
		}
	}
	if len(arr) == 0 {
		return types.Null
	}
	t := src.Without(scalar...)
	if !t.IsArrayLike() {
		return types.Unknown
	}
	var elem types.Type
	required := false
	switch {
	case t.HasShape():
		k, ok := shapeKey(t, key)
		if !ok {
			return types.Unknown
		}
		elem, required = k.Type, !k.Optional
	case !t.Has("array"):
		elem = t.Elem()
	default:
		return types.Unknown
	}
	if elem.IsUnknown() || required && len(scalar) == 0 {
		return elem
	}
	return types.Union(elem, types.Null)
}

// shapeKey returns key name of t's shape.
func shapeKey(t types.Type, name string) (types.ShapeKey, bool) {
	for _, k := range t.ShapeKeys() {
		if k.Name == name {
			return k, true
		}
	}
	return types.ShapeKey{}, false
}

// incDecMismatches reports ++/-- storing into a property a value its type
// rejects: `null` incremented to an integer.
func incDecMismatches(ctx *analysis.Context, n *syntax.IncDec) []PropertyMismatch {
	if !propertyTarget(n.Var) || reboundThis(n) {
		return nil
	}
	w := propertyStore{target: n.Var, at: n, report: n, modifies: true}
	if m, ok := judgeStore(ctx, w, func(env *infer.Env) types.Type { return env.IncDecStoredType(n) }); ok {
		return []PropertyMismatch{m}
	}
	return nil
}

// referenceMismatch reports `$this->p = &$local` when the local's current
// value is one the property rejects: PHP checks it when binding, as for a
// plain assignment. Other referenced values (elements, properties, calls)
// are left alone: an absent element is created as null.
func referenceMismatch(ctx *analysis.Context, a *syntax.Assign) (PropertyMismatch, bool) {
	v, ok := syntax.UnwrapParens(a.Value).(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return PropertyMismatch{}, false
	}
	w := propertyStore{target: a.Var, at: a, report: a.Value, modifies: true}
	return judgeStore(ctx, w, func(env *infer.Env) types.Type { return env.TypeOf(v) })
}
