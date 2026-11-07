package index

import (
	"reflect"

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
