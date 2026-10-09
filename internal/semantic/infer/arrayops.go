package infer

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// requiredShape admits only arrays whose keys are all known and present.
func requiredShape(t types.Type) bool {
	if !t.IsArrayLike() || !t.IsSealedShape() {
		return false
	}
	for _, k := range t.ShapeKeys() {
		if k.Optional {
			return false
		}
	}
	return true
}

// arrayElements includes shape values even when the array has mixed atoms.
func arrayElements(t types.Type) types.Type {
	if t.IsSealedShape() {
		ts := []types.Type{types.Of("never")}
		for _, k := range t.ShapeKeys() {
			ts = append(ts, k.Type)
		}
		return types.Union(ts...)
	}
	if !t.IsArrayLike() || t.Has("array") {
		return types.Unknown
	}
	return t.Elem()
}

// elementArray retains the existing literal-array atom representation.
func elementArray(elem types.Type, nonEmpty bool) types.Type {
	t := types.Array
	if !elem.IsUnknown() && !elem.OnlyOf("never") && len(elem.Atoms()) == 1 && !strings.HasSuffix(elem.Atoms()[0], "[]") {
		t = types.Of(elem.Atoms()[0] + "[]").WithElem(elem)
	}
	return t.WithNonEmpty(nonEmpty)
}

// arrayUnion implements PHP's left-key precedence without reindexing.
func arrayUnion(a, b types.Type) types.Type {
	elem := types.Union(arrayElements(a), arrayElements(b))
	nonEmpty := a.IsNonEmptyArray() || b.IsNonEmptyArray()
	key := types.Unknown
	if ak, bk := a.ArrayKey(), b.ArrayKey(); !ak.IsUnknown() && !bk.IsUnknown() {
		key = types.Union(ak, bk)
	}
	if !requiredShape(a) || !requiredShape(b) {
		return elementArray(elem, nonEmpty).WithArrayKey(key)
	}
	keys := append([]types.ShapeKey(nil), a.ShapeKeys()...)
	for _, k := range b.ShapeKeys() {
		found := false
		for _, left := range keys {
			if left.Name == k.Name {
				found = true
				break
			}
		}
		if !found {
			if len(keys) == types.MaxShapeKeys {
				return elementArray(elem, nonEmpty).WithArrayKey(key)
			}
			keys = append(keys, k)
		}
	}
	ts := []types.Type{types.Of("never")}
	for _, k := range keys {
		ts = append(ts, k.Type)
	}
	return elementArray(types.Union(ts...), nonEmpty).WithShape(keys, true)
}

// arrayType processes entries in source order. Shapes stop at the existing
// cap; uncertain spreads keep only element and emptiness facts.
func (e *Env) arrayType(n *syntax.Array) types.Type {
	keys := make([]types.ShapeKey, 0, min(len(n.Items), types.MaxShapeKeys))
	elems := []types.Type{types.Of("never")}
	precise, nonEmpty := true, false
	next, hasInt := int64(0), false
	put := func(name string, ty types.Type) {
		if !precise {
			return
		}
		if v, err := strconv.ParseInt(name, 10, 64); err == nil && strconv.FormatInt(v, 10) == name {
			switch {
			case !hasInt && v < 0 && e.PHP.Below(phpversion.PHP83):
				next = 0
			case !hasInt || v >= next:
				next = v + 1
			}
			hasInt = true
		}
		for i := range keys {
			if keys[i].Name == name {
				keys[i].Type = ty
				return
			}
		}
		if len(keys) == types.MaxShapeKeys {
			precise = false
			return
		}
		keys = append(keys, types.ShapeKey{Name: name, Type: ty})
	}
	for _, it := range n.Items {
		if it == nil || it.Value == nil {
			precise = false
			elems = append(elems, types.Unknown)
			continue
		}
		t := e.TypeOf(it.Value)
		if it.Unpack {
			if e.PHP.Below(phpversion.PHP74) || !t.IsArrayLike() {
				precise = false
				elems = append(elems, types.Unknown)
				continue
			}
			nonEmpty = nonEmpty || t.IsNonEmptyArray()
			elems = append(elems, arrayElements(t))
			if !requiredShape(t) {
				precise = false
				continue
			}
			for _, k := range t.ShapeKeys() {
				name := k.Name
				if v, err := strconv.ParseInt(name, 10, 64); err == nil && strconv.FormatInt(v, 10) == name {
					name = strconv.FormatInt(next, 10)
				} else if e.PHP.Below(phpversion.PHP81) {
					precise = false
					elems = append(elems, types.Unknown)
					break
				}
				put(name, k.Type)
			}
			continue
		}
		nonEmpty = true
		elems = append(elems, t)
		name := strconv.FormatInt(next, 10)
		if it.Key != nil {
			var ok bool
			name, ok = literalKey(it.Key)
			if !ok {
				precise = false
				continue
			}
		}
		put(name, t)
	}
	t := elementArray(types.Union(elems...), nonEmpty)
	if precise {
		return t.WithShape(keys, true)
	}
	return t
}
