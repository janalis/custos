package index

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// maxConstRefs bounds the `self::NAME` references followed while typing one
// initialiser (reference cycles stop there).
const maxConstRefs = 64

// constTyper types array literal constant initialisers at extraction time,
// when their source is at hand. local returns the initialiser of a constant
// of the class being extracted that a class constant fetch names (`self::`
// or the class's own name); nil outside a class.
type constTyper struct {
	local func(f *syntax.ClassConstFetch) syntax.Expr
	refs  int
}

// constArrayDoc returns the documented type of initialiser v when it is an
// array literal, "" otherwise: a sealed shape when every key is known, else
// the union of its values as element type with its emptiness. A value it
// cannot type is mixed.
func constArrayDoc(v syntax.Expr, local func(f *syntax.ClassConstFetch) syntax.Expr) string {
	a, ok := syntax.UnwrapParens(v).(*syntax.Array)
	if !ok {
		return ""
	}
	t := &constTyper{local: local}
	return t.array(a).DocString()
}

func (t *constTyper) ref(f *syntax.ClassConstFetch) syntax.Expr {
	if t.local == nil || t.refs >= maxConstRefs {
		return nil
	}
	t.refs++
	return t.local(f)
}

func (t *constTyper) of(v syntax.Expr) types.Type {
	switch n := syntax.UnwrapParens(v).(type) {
	case *syntax.Literal:
		switch n.LitKind {
		case syntax.LitInt:
			return types.Int
		case syntax.LitFloat:
			return types.Float
		}
		return types.String
	case *syntax.MagicConst:
		if n.Token.Kind == syntax.TLine {
			return types.Int
		}
		return types.String
	case *syntax.ConstFetch:
		switch v := strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)); v {
		case "true", "false":
			return types.Of(v)
		case "null":
			return types.Null
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TMinus || n.Op.Kind == syntax.TPlus {
			if o := t.of(n.Expr); o.OnlyOf("int") || o.OnlyOf("float") {
				return o
			}
		}
	case *syntax.Array:
		return t.array(n)
	case *syntax.ClassConstFetch:
		if x := t.ref(n); x != nil {
			return t.of(x)
		}
	}
	return types.Unknown
}

func (t *constTyper) key(k syntax.Expr) (string, bool) {
	if name, ok := syntax.LiteralArrayKey(k); ok {
		return name, true
	}
	if f, ok := syntax.UnwrapParens(k).(*syntax.ClassConstFetch); ok {
		if x := t.ref(f); x != nil {
			return t.key(x)
		}
	}
	return "", false
}

// array follows PHP's key assignment in source order; a spread or a key it
// cannot name keeps only the element type and emptiness.
func (t *constTyper) array(n *syntax.Array) types.Type {
	var keys []types.ShapeKey
	elems := []types.Type{types.Of("never")}
	precise, negative := true, false
	next := int64(0)
	for _, it := range n.Items {
		if it == nil || it.Value == nil || it.Unpack {
			precise = false
			elems = append(elems, types.Mixed)
			continue
		}
		v := t.of(it.Value)
		if v.IsUnknown() {
			v = types.Mixed
		}
		elems = append(elems, v)
		name := ""
		if it.Key == nil {
			// Before PHP 8.3 a key after negative ones restarts at 0.
			precise = precise && !negative
			name = strconv.FormatInt(next, 10)
		} else {
			var ok bool
			if name, ok = t.key(it.Key); !ok {
				precise = false
				continue
			}
		}
		if i, err := strconv.ParseInt(name, 10, 64); err == nil && strconv.FormatInt(i, 10) == name {
			negative = negative || i < 0
			next = max(next, i+1)
		}
		found := false
		for j := range keys {
			if keys[j].Name == name {
				keys[j].Type, found = v, true
			}
		}
		if !found {
			keys = append(keys, types.ShapeKey{Name: name, Type: v})
		}
	}
	if precise && len(keys) <= types.MaxShapeKeys {
		return types.Array.WithShape(keys, true)
	}
	// Only a literal with items gets here, so the union has an element atom.
	var atoms []string
	for _, a := range types.Union(elems...).Atoms() {
		if a != "never" {
			atoms = append(atoms, a+"[]")
		}
	}
	return types.Of(atoms...).WithNonEmpty(true)
}
