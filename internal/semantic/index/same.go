package index

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"reflect"
	"slices"

	"custos/internal/php/syntax"
)

var spanType = reflect.TypeFor[syntax.Span]()

// SameDeclarations reports whether a and b, two extractions of one file,
// declare the same symbols with the same facts, wherever they sit: they
// differ at most in source positions, which give no other file a type, and
// in the digest of their bytes.
func SameDeclarations(a, b *FileSymbols) bool {
	if a == nil || b == nil {
		return a == b
	}
	x, y := *a, *b
	x.Digest, y.Digest = 0, 0
	return sameValue(reflect.ValueOf(&x), reflect.ValueOf(&y))
}

// sameValue is reflect.DeepEqual for the acyclic symbol trees, with every
// syntax.Span equal.
func sameValue(a, b reflect.Value) bool {
	if a.Type() == spanType {
		return true
	}
	switch a.Kind() {
	case reflect.Pointer:
		if a.IsNil() || b.IsNil() {
			return a.IsNil() == b.IsNil()
		}
		return sameValue(a.Elem(), b.Elem())
	case reflect.Struct:
		for i := range a.NumField() {
			if !sameValue(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if a.Kind() == reflect.Slice && a.IsNil() != b.IsNil() || a.Len() != b.Len() {
			return false
		}
		for i := range a.Len() {
			if !sameValue(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if a.IsNil() != b.IsNil() || a.Len() != b.Len() {
			return false
		}
		for _, k := range a.MapKeys() {
			v := b.MapIndex(k)
			if !v.IsValid() || !sameValue(a.MapIndex(k), v) {
				return false
			}
		}
		return true
	}
	return a.Equal(b) // strings, numbers and booleans
}

// Fingerprint hashes fs as SameDeclarations compares it: two extractions
// SameDeclarations calls the same have the same fingerprint, and any other
// difference changes it but for a SHA-256 collision. Unlike Digest, it is
// stable across processes, so a cache kept between runs can compare
// declarations it no longer holds.
func Fingerprint(fs *FileSymbols) [sha256.Size]byte {
	var buf []byte
	if fs != nil {
		x := *fs
		x.Digest = 0
		buf = appendValue(buf, reflect.ValueOf(&x))
	}
	return sha256.Sum256(buf)
}

// appendValue appends the encoding of a to buf, as sameValue compares it:
// every syntax.Span the same, nil distinct from empty, maps in key order.
func appendValue(buf []byte, a reflect.Value) []byte {
	if a.Type() == spanType {
		return buf
	}
	switch a.Kind() {
	case reflect.Pointer:
		if a.IsNil() {
			return append(buf, 0)
		}
		return appendValue(append(buf, 1), a.Elem())
	case reflect.Struct:
		for i := range a.NumField() {
			buf = appendValue(buf, a.Field(i))
		}
		return buf
	case reflect.Slice, reflect.Array:
		if a.Kind() == reflect.Slice && a.IsNil() {
			return append(buf, 0)
		}
		buf = binary.AppendUvarint(buf, uint64(a.Len())+1)
		for i := range a.Len() {
			buf = appendValue(buf, a.Index(i))
		}
		return buf
	case reflect.Map:
		if a.IsNil() {
			return append(buf, 0)
		}
		buf = binary.AppendUvarint(buf, uint64(a.Len())+1)
		type entry struct {
			key   []byte
			value reflect.Value
		}
		entries := make([]entry, 0, a.Len())
		for it := a.MapRange(); it.Next(); {
			entries = append(entries, entry{appendValue(nil, it.Key()), it.Value()})
		}
		slices.SortFunc(entries, func(x, y entry) int { return bytes.Compare(x.key, y.key) })
		for _, e := range entries {
			buf = appendValue(append(buf, e.key...), e.value)
		}
		return buf
	case reflect.String:
		buf = binary.AppendUvarint(buf, uint64(a.Len()))
		return append(buf, a.String()...)
	case reflect.Bool:
		if a.Bool() {
			return append(buf, 1)
		}
		return append(buf, 0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return binary.AppendVarint(buf, a.Int())
	}
	return binary.AppendUvarint(buf, a.Uint()) // the other kinds symbols hold
}
