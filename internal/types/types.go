// Package types models PHP types as sets of atoms (PhpStorm style) and
// infers expression types.
//
// Atoms are: int float string bool true false null array callable iterable
// object mixed void never resource static, class names with a leading
// backslash (\Foo\Bar), and element-typed arrays written `T[]` (\Foo[], int[]).
package types

import (
	"slices"
	"sort"
	"strings"
)

// Type is an immutable union of atoms. The zero value is "unknown" (no
// information); use Mixed for an explicit mixed.
//
// A type may also carry array facts (see shape.go) describing its array
// members: per-key types, non-emptiness. They never change the atom set, so
// atom-based checks (Has("array"), IsArrayLike, Atoms) behave as without
// them; Equal and String ignore them (ShapeString shows them).
type Type struct {
	atoms []string   // sorted, unique; nil = unknown
	arr   *arrayInfo // optional array facts; nil when none
	gen   []genEntry // generic arguments of class atoms (see gen.go); nil when none
}

// Common types.
var (
	Unknown = Type{}
	Mixed   = Of("mixed")
	Int     = Of("int")
	Float   = Of("float")
	String  = Of("string")
	Bool    = Of("bool")
	Null    = Of("null")
	Array   = Of("array")
	Void    = Of("void")
)

// Of builds a type from atoms (normalised and deduplicated).
func Of(atoms ...string) Type {
	out := make([]string, 0, len(atoms))
	for _, a := range atoms {
		if a = normalizeAtom(a); a != "" {
			out = append(out, a)
		}
	}
	// never is the bottom type: it disappears from any non-empty union.
	if len(out) > 1 {
		k := 0
		for _, a := range out {
			if a != "never" {
				out[k] = a
				k++
			}
		}
		if k > 0 {
			out = out[:k]
		}
	}
	sort.Strings(out)
	j := 0
	for i := range out {
		if i == 0 || out[i] != out[j-1] {
			out[j] = out[i]
			j++
		}
	}
	return Type{atoms: out[:j]}
}

// IsUnknown reports whether nothing is known about the type.
func (t Type) IsUnknown() bool { return len(t.atoms) == 0 }

// Atoms returns the atoms (do not modify).
func (t Type) Atoms() []string { return t.atoms }

// Has reports whether atom is part of the union.
func (t Type) Has(atom string) bool {
	atom = normalizeAtom(atom)
	i := sort.SearchStrings(t.atoms, atom)
	return i < len(t.atoms) && t.atoms[i] == atom
}

// HasAny reports whether any of atoms is present.
func (t Type) HasAny(atoms ...string) bool {
	for _, a := range atoms {
		if t.Has(a) {
			return true
		}
	}
	return false
}

// OnlyOf reports whether every atom is one of atoms (false for unknown).
func (t Type) OnlyOf(atoms ...string) bool {
	if t.IsUnknown() {
		return false
	}
	for _, a := range t.atoms {
		found := false
		for _, b := range atoms {
			if a == normalizeAtom(b) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Union merges types. Unknown absorbs nothing: unknown|int = int|unknown
// is represented as unknown (callers must treat partial knowledge as unknown).
func Union(ts ...Type) Type {
	var all []string
	info := false
	for _, t := range ts {
		if t.IsUnknown() {
			return Unknown
		}
		all = append(all, t.atoms...)
		info = info || t.arr != nil
	}
	u := Of(all...)
	if info {
		u = u.withInfo(unionInfo(ts))
	}
	return u.withGen(unionGen(u, ts))
}

// Without removes atoms.
func (t Type) Without(atoms ...string) Type {
	if t.IsUnknown() {
		return t
	}
	// Normalize the dropped atoms once (not once per member).
	var buf [8]string
	drop := buf[:0]
	for _, b := range atoms {
		drop = append(drop, normalizeAtom(b))
	}
	out := make([]string, 0, len(t.atoms))
	for _, a := range t.atoms {
		if !slices.Contains(drop, a) {
			out = append(out, a)
		}
	}
	return Type{atoms: out}.withInfo(t.arr).withGen(t.gen)
}

// Classes returns class atoms (with leading backslash, excluding T[] forms).
func (t Type) Classes() []string {
	var out []string
	for _, a := range t.atoms {
		if strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]") {
			out = append(out, a)
		}
	}
	return out
}

// IsNullable reports whether null is part of the type.
func (t Type) IsNullable() bool { return t.Has("null") }

// Elem returns the element type of array atoms (T[] → T); unknown when none.
func (t Type) Elem() Type {
	var out []string
	for _, a := range t.atoms {
		if strings.HasSuffix(a, "[]") {
			out = append(out, strings.TrimSuffix(a, "[]"))
		}
	}
	if len(out) == 0 {
		return Unknown
	}
	el := Of(out...)
	if t.arr != nil && t.arr.elem != nil {
		el = el.withInfo(t.arr.elem)
	}
	return el
}

// IsArrayLike reports whether every atom is an array form.
func (t Type) IsArrayLike() bool {
	if t.IsUnknown() {
		return false
	}
	for _, a := range t.atoms {
		if a != "array" && !strings.HasSuffix(a, "[]") {
			return false
		}
	}
	return true
}

func (t Type) String() string {
	if len(t.atoms) == 0 { // as IsUnknown: Without may leave an empty set
		return "?unknown"
	}
	return strings.Join(t.atoms, "|")
}

// Equal reports atom-set equality.
func (t Type) Equal(o Type) bool {
	if len(t.atoms) != len(o.atoms) || (t.atoms == nil) != (o.atoms == nil) {
		return false
	}
	for i := range t.atoms {
		if t.atoms[i] != o.atoms[i] {
			return false
		}
	}
	return true
}

var scalarAliases = map[string]string{
	"integer": "int", "boolean": "bool", "double": "float", "real": "float",
	"$this": "static", "callback": "callable", "void": "void",
}

func normalizeAtom(a string) string {
	a = strings.TrimSpace(a)
	if a == "" {
		return ""
	}
	if strings.HasPrefix(a, `\`) {
		if foldIn(builtinAtoms, a[1:]) {
			return strings.ToLower(a[1:])
		}
		return a
	}
	if r, ok := foldLookup(scalarAliases, a); ok {
		return r
	}
	if foldIn(builtinAtoms, strings.TrimSuffix(a, "[]")) {
		return strings.ToLower(a)
	}
	return a
}

// foldLookup looks s up in m (lower-case keys) ignoring case, without
// allocating: normalizeAtom runs for every Of/Has/Without on class names,
// and lower-casing them each time dominated the allocations of narrowing.
// Non-ASCII text goes through strings.ToLower (the Kelvin sign folds to k).
func foldLookup[V any](m map[string]V, s string) (V, bool) {
	var buf [16]byte
	if len(s) > len(buf) {
		var zero V
		return zero, false // longer than any key
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x80 {
			v, ok := m[strings.ToLower(s)]
			return v, ok
		}
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		buf[i] = c
	}
	v, ok := m[string(buf[:len(s)])]
	return v, ok
}

func foldIn(m map[string]bool, s string) bool {
	v, _ := foldLookup(m, s)
	return v
}

var builtinAtoms = map[string]bool{
	"int": true, "float": true, "string": true, "bool": true, "true": true, "false": true,
	"null": true, "array": true, "callable": true, "iterable": true, "object": true, "mixed": true,
	"void": true, "never": true, "resource": true, "static": true, "self": true, "parent": true,
}

func isBuiltinAtom(s string) bool { return builtinAtoms[s] }
