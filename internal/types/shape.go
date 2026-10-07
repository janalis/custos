package types

import (
	"strconv"
	"strings"
)

// MaxShapeKeys caps the number of keys a shape tracks; a larger array keeps
// its atoms (and emptiness) but no per-key types.
const MaxShapeKeys = 32

// arrayInfo refines the array members (`array`, `T[]`) of a type with
// facts the atom set cannot express: per-key types (shapes), emptiness, and
// the same facts about the elements of `T[]` members. It is immutable and
// shared between types.
type arrayInfo struct {
	shape    bool       // keys describe the array's keys
	sealed   bool       // the array has no keys beyond keys (literal arrays, doc shapes)
	nonEmpty bool       // the array has at least one element
	keys     []ShapeKey // in declaration order; at most MaxShapeKeys
	elem     *arrayInfo // array facts of the elements of `T[]` members
}

// ShapeKey is one key of an array shape. Name is the key as PHP stores it:
// integer-like keys in canonical decimal form ("0", "-3"), strings as is.
type ShapeKey struct {
	Name     string
	Type     Type
	Optional bool
}

func (a *arrayInfo) empty() bool {
	return a == nil || (!a.shape && !a.nonEmpty && a.elem == nil)
}

func (a *arrayInfo) isNonEmpty() bool {
	if a == nil {
		return false
	}
	if a.nonEmpty {
		return true
	}
	if a.shape {
		for _, k := range a.keys {
			if !k.Optional {
				return true
			}
		}
	}
	return false
}

func norm(a *arrayInfo) *arrayInfo {
	if a.empty() {
		return nil
	}
	return a
}

// hasArrayAtom reports whether t has an `array` or `T[]` member.
func (t Type) hasArrayAtom() bool {
	for _, a := range t.atoms {
		if a == "array" || strings.HasSuffix(a, "[]") {
			return true
		}
	}
	return false
}

// withInfo attaches array facts to t (dropped when t has no array member).
func (t Type) withInfo(a *arrayInfo) Type {
	a = norm(a)
	if a == nil || !t.hasArrayAtom() {
		t.arr = nil
		return t
	}
	t.arr = a
	return t
}

// WithShape returns t (which must have an array member) with per-key types.
// sealed states that the array has no other keys (a literal array). More
// than MaxShapeKeys keys keep only the emptiness fact.
func (t Type) WithShape(keys []ShapeKey, sealed bool) Type {
	a := t.cloneInfo()
	if len(keys) > MaxShapeKeys {
		a.shape, a.sealed, a.keys = false, false, nil
		for _, k := range keys {
			if !k.Optional {
				a.nonEmpty = true
				break
			}
		}
		return t.withInfo(a)
	}
	a.shape, a.sealed, a.keys = true, sealed, keys
	return t.withInfo(a)
}

func (t Type) cloneInfo() *arrayInfo {
	if t.arr == nil {
		return &arrayInfo{}
	}
	c := *t.arr
	return &c
}

// HasShape reports whether per-key types are known for t's array members.
func (t Type) HasShape() bool { return t.arr != nil && t.arr.shape }

// IsSealedShape reports a shape listing every key of the array.
func (t Type) IsSealedShape() bool { return t.arr != nil && t.arr.shape && t.arr.sealed }

// ShapeKeys returns the shape's keys (do not modify); nil without a shape.
func (t Type) ShapeKeys() []ShapeKey {
	if t.arr == nil || !t.arr.shape {
		return nil
	}
	return t.arr.keys
}

// ShapeKey returns the type of key name in t's shape (ok false when t has no
// shape or the shape does not list the key).
func (t Type) ShapeKey(name string) (Type, bool) {
	if t.arr == nil || !t.arr.shape {
		return Unknown, false
	}
	for _, k := range t.arr.keys {
		if k.Name == name {
			return k.Type, true
		}
	}
	return Unknown, false
}

// MapShape rewrites every shape key with fn (t unchanged without a shape).
func (t Type) MapShape(fn func(ShapeKey) ShapeKey) Type {
	if !t.HasShape() {
		return t
	}
	keys := make([]ShapeKey, len(t.arr.keys))
	for i, k := range t.arr.keys {
		keys[i] = fn(k)
	}
	a := t.cloneInfo()
	a.keys = keys
	return t.withInfo(a)
}

// IsNonEmptyArray reports whether t's array members are known to hold at
// least one element (literal non-empty arrays, `non-empty-array`, shapes
// with a required key, or narrowing such as `!empty($x)`).
func (t Type) IsNonEmptyArray() bool { return t.hasArrayAtom() && t.arr.isNonEmpty() }

// WithNonEmpty sets (or clears) the non-empty fact of t's array members.
// Clearing also drops required shape keys' guarantee by making them optional.
func (t Type) WithNonEmpty(nonEmpty bool) Type {
	if !t.hasArrayAtom() || t.arr.isNonEmpty() == nonEmpty {
		return t
	}
	a := t.cloneInfo()
	a.nonEmpty = nonEmpty
	if !nonEmpty && a.shape {
		keys := make([]ShapeKey, len(a.keys))
		for i, k := range a.keys {
			k.Optional = true
			keys[i] = k
		}
		a.keys = keys
	}
	return t.withInfo(a)
}

// WithoutShape drops per-key types, keeping emptiness and element facts
// (for operations that renumber or drop keys).
func (t Type) WithoutShape() Type {
	if !t.HasShape() {
		return t
	}
	a := t.cloneInfo()
	a.nonEmpty = a.isNonEmpty()
	a.shape, a.sealed, a.keys = false, false, nil
	return t.withInfo(a)
}

// WithoutArrayInfo drops every array fact beyond the atoms.
func (t Type) WithoutArrayInfo() Type {
	t.arr = nil
	return t
}

// WithElem attaches the array facts of elem (the element type) to t's `T[]`
// members.
func (t Type) WithElem(elem Type) Type {
	if elem.arr == nil && (t.arr == nil || t.arr.elem == nil) {
		return t
	}
	a := t.cloneInfo()
	a.elem = elem.arr
	return t.withInfo(a)
}

// mergeInfo combines the array facts of two union members.
func mergeInfo(a, b *arrayInfo) *arrayInfo {
	if a == nil || b == nil {
		return nil
	}
	if a == b {
		return a
	}
	out := &arrayInfo{nonEmpty: a.isNonEmpty() && b.isNonEmpty(), elem: mergeInfo(a.elem, b.elem)}
	if a.shape && b.shape {
		out.shape, out.sealed = true, a.sealed && b.sealed
		keys := make([]ShapeKey, 0, len(a.keys)+len(b.keys))
		for _, k := range a.keys {
			if bt, ok := findKey(b.keys, k.Name); ok {
				keys = append(keys, ShapeKey{Name: k.Name, Type: Union(k.Type, bt.Type), Optional: k.Optional || bt.Optional})
			} else if b.sealed {
				k.Optional = true
				keys = append(keys, k)
			}
		}
		for _, k := range b.keys {
			if _, ok := findKey(a.keys, k.Name); !ok && a.sealed {
				k.Optional = true
				keys = append(keys, k)
			}
		}
		if len(keys) > MaxShapeKeys {
			out.shape, out.sealed = false, false
		} else {
			out.keys = keys
		}
	}
	return norm(out)
}

func findKey(keys []ShapeKey, name string) (ShapeKey, bool) {
	for _, k := range keys {
		if k.Name == name {
			return k, true
		}
	}
	return ShapeKey{}, false
}

// unionInfo computes the array facts of a union: every member with an
// array atom must carry facts, members without array atoms do not count.
func unionInfo(ts []Type) *arrayInfo {
	var out *arrayInfo
	seen := false
	for _, t := range ts {
		if !t.hasArrayAtom() {
			continue
		}
		if t.arr == nil {
			return nil
		}
		if !seen {
			out, seen = t.arr, true
			continue
		}
		if out = mergeInfo(out, t.arr); out == nil {
			return nil
		}
	}
	return out
}

// ShapeString renders t with its array facts, for tests and debugging:
// the atoms (as String) prefixed by "non-empty " when the array members are
// known non-empty without a shape saying so, followed by the shape
// `{a: int, b?: string}` (`...` when unsealed) and the element facts `<…>`.
func (t Type) ShapeString() string {
	base := t.String()
	if len(t.gen) > 0 {
		parts := make([]string, len(t.atoms))
		for i, a := range t.atoms {
			parts[i] = a
			if args := t.TypeArgs(a); args != nil {
				parts[i] = genString(a, args)
			}
		}
		base = strings.Join(parts, "|")
	}
	if t.arr == nil {
		return base
	}
	return infoString(t.arr, base)
}

func infoString(a *arrayInfo, base string) string {
	var b strings.Builder
	if a.nonEmpty && !(a.shape && a.isNonEmptyByKeys()) {
		b.WriteString("non-empty ")
	}
	b.WriteString(base)
	if a.shape {
		b.WriteByte('{')
		for j, k := range a.keys {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(k.Name)
			if k.Optional {
				b.WriteByte('?')
			}
			b.WriteString(": ")
			b.WriteString(k.Type.ShapeString())
		}
		if !a.sealed {
			if len(a.keys) > 0 {
				b.WriteString(", ")
			}
			b.WriteString("...")
		}
		b.WriteByte('}')
	}
	if a.elem != nil {
		b.WriteString("<" + infoString(a.elem, "array") + ">")
	}
	return b.String()
}

func (a *arrayInfo) isNonEmptyByKeys() bool {
	for _, k := range a.keys {
		if !k.Optional {
			return true
		}
	}
	return false
}

// DocString renders t as a PHPDoc type that FromDoc parses back to the same
// atoms and array facts (shapes on the `array` member, `non-empty-array`,
// element shapes). Used to store doc types in the symbol index.
func (t Type) DocString() string {
	if t.arr == nil && t.gen == nil {
		return t.String()
	}
	parts := make([]string, len(t.atoms))
	for i, a := range t.atoms {
		if args := t.TypeArgs(a); args != nil {
			parts[i] = genString(a, args)
			continue
		}
		parts[i] = docAtom(a, t.arr)
	}
	return strings.Join(parts, "|")
}

func docAtom(atom string, a *arrayInfo) string {
	if a == nil {
		return atom
	}
	if atom == "array" {
		if !a.shape {
			if a.nonEmpty {
				return "non-empty-array"
			}
			return atom
		}
		var b strings.Builder
		if a.nonEmpty && !a.isNonEmptyByKeys() {
			b.WriteString("non-empty-")
		}
		b.WriteString("array{")
		for j, k := range a.keys {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteString(docKey(k.Name))
			if k.Optional {
				b.WriteByte('?')
			}
			b.WriteByte(':')
			if k.Type.IsUnknown() {
				b.WriteString("mixed")
			} else {
				b.WriteString(k.Type.DocString())
			}
		}
		if !a.sealed {
			if len(a.keys) > 0 {
				b.WriteByte(',')
			}
			b.WriteString("...")
		}
		b.WriteByte('}')
		return b.String()
	}
	if !strings.HasSuffix(atom, "[]") {
		return atom
	}
	el := strings.TrimSuffix(atom, "[]")
	if a.elem != nil && (el == "array" || strings.HasSuffix(el, "[]")) {
		el = docAtom(el, a.elem)
	}
	if a.nonEmpty {
		return "non-empty-array<" + el + ">"
	}
	if strings.ContainsAny(el, "{<") {
		return "array<" + el + ">"
	}
	return el + "[]"
}

// docKey quotes a shape key unless it is a plain identifier or integer.
func docKey(k string) string {
	plain := k != ""
	for i := 0; i < len(k); i++ {
		c := k[i]
		if !(c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			plain = false
			break
		}
	}
	if plain {
		return k
	}
	if !strings.Contains(k, "'") {
		return "'" + k + "'"
	}
	return `"` + k + `"`
}

// IsIntKey reports whether a canonical key is an integer key.
func IsIntKey(s string) bool {
	n, err := strconv.ParseInt(s, 10, 64)
	return err == nil && strconv.FormatInt(n, 10) == s
}
