package infer

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// arrayBuiltins are the array-returning builtins arrayBuiltinType types
// from their arguments.
var arrayBuiltins = map[string]bool{
	"array_merge": true, "array_replace": true, "array_replace_recursive": true, "array_chunk": true, "array_combine": true,
	"array_column": true, "array_fill": true, "array_fill_keys": true, "array_keys": true,
	"array_flip": true, "array_pad": true, "array_count_values": true, "iterator_to_array": true,
	"array_intersect_key": true, "array_diff_key": true, "array_intersect": true, "array_diff": true,
	"array_intersect_assoc": true, "array_diff_assoc": true, "range": true,
	"fgetcsv": true, "str_getcsv": true, "file": true, "glob": true, "preg_split": true,
}

// shapeless is array t without its shape: a sealed shape's values become
// the element type, as its keys are renumbered or dropped.
func shapeless(t types.Type) types.Type {
	if !t.IsSealedShape() {
		return t.WithoutShape()
	}
	el := arrayElements(t)
	if el.OnlyOf("never") {
		return t.WithoutShape() // an empty literal
	}
	return arrayOf(el).WithNonEmpty(t.IsNonEmptyArray()).WithArrayKey(t.ArrayKey())
}

// listOf is the list of el elements.
func listOf(el types.Type, nonEmpty bool) types.Type {
	return arrayOf(el).WithNonEmpty(nonEmpty).WithArrayKey(types.Int)
}

// literalFalse reports whether x is absent or the constant false.
func literalFalse(x syntax.Expr) bool { return x == nil || constLiteral(x) == "false" }

// arrayBuiltinType types a call of an arrayBuiltins function from its
// arguments (arg(i): the argument bound to parameter i, nil when absent);
// ok is false when its arguments say nothing more than the declaration.
func (e *Env) arrayBuiltinType(builtin string, n *syntax.FuncCall, arg func(int) syntax.Expr) (types.Type, bool) {
	// (The argument binding proof guarantees the required arguments.)
	array := func(i int) (types.Type, bool) {
		t := e.TypeOf(arg(i))
		return t, t.IsArrayLike()
	}
	switch builtin {
	case "array_merge", "array_replace":
		return e.mergeType(n, builtin == "array_merge")
	case "array_replace_recursive":
		// Arrays at the same key merge too: the union of the elements
		// (keys only one has optional) holds what they merge into.
		var elems, keys []types.Type
		nonEmpty := false
		for _, x := range n.Args.Args {
			a, ok := x.(*syntax.Arg)
			if !ok || a.Unpack || a.Name != nil {
				return types.Unknown, false
			}
			t := e.TypeOf(a.Value)
			if !t.IsArrayLike() {
				return types.Unknown, false
			}
			elems, keys = append(elems, arrayElements(t)), append(keys, t.ArrayKey())
			nonEmpty = nonEmpty || t.IsNonEmptyArray()
		}
		el := types.Union(elems...)
		if el.IsUnknown() {
			return types.Array.WithNonEmpty(nonEmpty), true
		}
		return arrayOf(el).WithNonEmpty(nonEmpty).WithArrayKey(types.Union(keys...)), true
	case "array_chunk":
		t, ok := array(0)
		if !ok {
			return types.Unknown, false
		}
		chunk := arrayOf(arrayElements(t))
		if arrayElements(t).IsUnknown() {
			chunk = types.Array
		}
		chunk = chunk.WithNonEmpty(true)
		if literalFalse(arg(2)) {
			chunk = chunk.WithArrayKey(types.Int)
		} else {
			chunk = chunk.WithArrayKey(t.ArrayKey())
		}
		return arrayOf(chunk).WithNonEmpty(t.IsNonEmptyArray()).WithArrayKey(types.Int), true
	case "array_combine":
		k, kok := array(0)
		v, vok := array(1)
		if !kok || !vok {
			return types.Unknown, false
		}
		r := shapeless(v).WithArrayKey(keyDomain(arrayElements(k)))
		if r.Has("array") {
			r = types.Array.WithNonEmpty(r.IsNonEmptyArray()).WithArrayKey(r.ArrayKey())
		}
		if e.PHP.Below(phpversion.PHP80) {
			return types.Union(r, types.Of("false")), true // a count mismatch
		}
		return r, true
	case "array_column":
		t, ok := array(0)
		el := arrayElements(t)
		if !ok || el.IsUnknown() || !el.IsArrayLike() {
			return types.Unknown, false
		}
		var col types.Type
		switch c := arg(1); {
		case constLiteral(c) == "null":
			col = el
		default:
			name, lit := syntax.LiteralArrayKey(c)
			kt, known := el.ShapeKey(name)
			if !lit || !known || !el.IsSealedShape() {
				return types.Unknown, false
			}
			col = kt
		}
		r := arrayOf(col)
		if arg(2) == nil {
			r = r.WithArrayKey(types.Int)
		}
		return r, true
	case "array_fill":
		if v := arg(2); v != nil {
			if vt := e.TypeOf(v); !vt.IsUnknown() {
				return arrayOf(vt).WithArrayKey(types.Int), true
			}
		}
	case "array_fill_keys":
		k, ok := array(0)
		if v := arg(1); ok && v != nil {
			if vt := e.TypeOf(v); !vt.IsUnknown() {
				return arrayOf(vt).WithNonEmpty(k.IsNonEmptyArray()).WithArrayKey(keyDomain(arrayElements(k))), true
			}
		}
	case "array_pad":
		t, ok := array(0)
		el := arrayElements(t)
		if v := arg(2); ok && v != nil && !el.IsUnknown() {
			if vt := e.TypeOf(v); !vt.IsUnknown() {
				return arrayOf(types.Union(el, vt)).WithArrayKey(t.ArrayKey()), true
			}
		}
	case "array_keys", "array_flip":
		t, ok := array(0)
		if !ok {
			return types.Unknown, false
		}
		k := t.ArrayKey()
		if k.IsUnknown() {
			k = types.Of("int", "string")
		}
		if builtin == "array_keys" {
			return listOf(k, t.IsNonEmptyArray() && arg(1) == nil), true
		}
		return arrayOf(k).WithNonEmpty(t.IsNonEmptyArray()).WithArrayKey(keyDomain(arrayElements(t))), true
	case "array_count_values":
		t, ok := array(0)
		if !ok {
			return types.Unknown, false
		}
		return arrayOf(types.Int).WithNonEmpty(t.IsNonEmptyArray()).WithArrayKey(keyDomain(arrayElements(t))), true
	case "array_intersect_key", "array_diff_key", "array_intersect", "array_diff", "array_intersect_assoc", "array_diff_assoc":
		// A subset of the first array's entries, unchanged.
		t, ok := array(0)
		if !ok {
			return types.Unknown, false
		}
		if t.IsSealedShape() {
			keys := append([]types.ShapeKey(nil), t.ShapeKeys()...)
			for i := range keys {
				keys[i].Optional = true
			}
			return t.WithNonEmpty(false).WithShape(keys, true), true
		}
		return t.WithNonEmpty(false), true
	case "iterator_to_array":
		_, v := e.iterTypes(e.TypeOf(arg(0)))
		if v.IsUnknown() {
			return types.Unknown, false
		}
		r := arrayOf(v)
		if constLiteral(arg(1)) == "false" {
			r = r.WithArrayKey(types.Int)
		}
		return r, true
	case "range":
		lo, hi := arg(0), arg(1)
		step := types.Int
		if s := arg(2); s != nil {
			step = e.TypeOf(s)
		}
		if e.TypeOf(lo).OnlyOf("int") && e.TypeOf(hi).OnlyOf("int") && step.OnlyOf("int") {
			return listOf(types.Int, true), true
		}
	case "fgetcsv", "str_getcsv":
		// A blank line reads as [null].
		r := listOf(types.Of("string", "null"), true)
		if builtin == "str_getcsv" {
			return r, true
		}
		if e.PHP.Below(phpversion.PHP80) {
			return types.Union(r, types.Of("false", "null")), true
		}
		return types.Union(r, types.Of("false")), true
	case "file", "glob":
		return types.Union(listOf(types.String, false), types.Of("false")), true
	case "preg_split":
		flags := arg(3)
		if flags == nil {
			return types.Unknown, false
		}
		pair := types.Array.WithShape([]types.ShapeKey{{Name: "0", Type: types.String}, {Name: "1", Type: types.Int}}, true)
		capture, known := e.pregSplitOffsetCapture(flags)
		switch {
		case !known:
			return types.Union(listOf(types.Union(types.String, pair), false), types.Of("false")), true
		case capture:
			return types.Union(listOf(pair, false), types.Of("false")), true
		}
		return types.Union(listOf(types.String, false), types.Of("false")), true
	}
	return types.Unknown, false
}

// pregSplitOffsetCapture reports whether preg_split() flags x include
// PREG_SPLIT_OFFSET_CAPTURE; known is false unless x is a `|` combination
// of named constants (or 0).
func (e *Env) pregSplitOffsetCapture(x syntax.Expr) (capture, known bool) {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.ConstFetch:
		return strings.EqualFold(strings.TrimPrefix(n.Name.Value, `\`), "PREG_SPLIT_OFFSET_CAPTURE"), true
	case *syntax.Literal:
		return false, n.LitKind == syntax.LitInt && n.Raw == "0"
	case *syntax.Binary:
		if n.Op.Kind != syntax.TBar {
			return false, false
		}
		l, lk := e.pregSplitOffsetCapture(n.Left)
		r, rk := e.pregSplitOffsetCapture(n.Right)
		return l || r, lk && rk
	}
	return false, false
}

// keyDomain is the key domain of an array whose keys are the values el:
// int for integers, else unknown (a numeric string becomes an int key).
func keyDomain(el types.Type) types.Type {
	if el.OnlyOf("int") {
		return types.Int
	}
	return types.Unknown
}

// mergeType types array_merge() (merge: integer keys renumbered) and
// array_replace() from their written arguments: a merged sealed shape when
// every argument is one with known keys, else the union of their elements.
func (e *Env) mergeType(n *syntax.FuncCall, merge bool) (types.Type, bool) {
	var args []types.Type
	for _, node := range n.Args.Args {
		t := e.TypeOf(node.(*syntax.Arg).Value) // no spread: see builtinBindingsKnown
		if !t.IsArrayLike() {
			return types.Unknown, false
		}
		args = append(args, t)
	}
	if len(args) == 0 {
		return types.Array.WithShape(nil, true), true // array_merge() from PHP 7.4
	}
	if t, ok := mergedShape(args, merge); ok {
		return t, true
	}
	ts := []types.Type{types.Of("never")}
	nonEmpty := false
	for _, t := range args {
		ts = append(ts, arrayElements(t))
		nonEmpty = nonEmpty || t.IsNonEmptyArray()
	}
	el := types.Union(ts...)
	if el.IsUnknown() || el.OnlyOf("never") {
		return types.Array.WithNonEmpty(nonEmpty), true
	}
	return arrayOf(el).WithNonEmpty(nonEmpty), true
}

// mergedShape merges sealed shapes with only required keys in order: a
// later string key (array_replace: any key) replaces an earlier one, and
// array_merge appends integer keys after the integer keys so far.
func mergedShape(args []types.Type, merge bool) (types.Type, bool) {
	var keys []types.ShapeKey
	next := int64(0)
	for _, t := range args {
		if !requiredShape(t) {
			return types.Unknown, false
		}
		for _, k := range t.ShapeKeys() {
			name := k.Name
			if i, err := strconv.ParseInt(name, 10, 64); err == nil && types.IsIntKey(name) {
				if merge {
					name = strconv.FormatInt(next, 10)
					next++
				} else {
					next = max(next, i+1)
				}
			}
			found := false
			for j := range keys {
				if keys[j].Name == name {
					keys[j].Type, found = k.Type, true
				}
			}
			if !found {
				keys = append(keys, types.ShapeKey{Name: name, Type: k.Type})
			}
		}
	}
	if len(keys) > types.MaxShapeKeys {
		return types.Unknown, false
	}
	return types.Array.WithShape(keys, true), true
}
