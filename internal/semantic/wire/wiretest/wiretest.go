// Package wiretest helps test the codecs semantic packages write with wire:
// Fill sets every field of a value, so that a codec forgetting one fails a
// round trip, and Fields lists the fields of every struct a value reaches,
// so that a test can name those it encodes and fail when one is added.
package wiretest

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Fill sets every field reachable from ptr to a non-zero value, each number
// and string distinct, slices and maps to two elements, down to depth
// levels of pointers, slices and maps; deeper ones stay nil. Interfaces
// stay nil, as do the fields skip names ("pkg.Type.field"). A pointer
// field inline names ("pkg.Type.field") stands for fields its parent would
// otherwise hold itself: it is set at every depth, 0 included, without
// counting as a level, so that the fields it points to are set at the
// parent's level.
func Fill(ptr any, depth int, skip map[string]bool, inline ...string) {
	f := &filler{skip: skip, inline: map[string]bool{}}
	for _, name := range inline {
		f.inline[name] = true
	}
	f.fill(reflect.ValueOf(ptr).Elem(), depth)
}

type filler struct {
	skip, inline map[string]bool
	n            int
}

func (f *filler) next() int {
	f.n++
	return f.n
}

func (f *filler) fill(v reflect.Value, depth int) {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(f.next()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(uint64(f.next()))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(f.next()) + 0.5)
	case reflect.String:
		v.SetString(fmt.Sprintf("s%d", f.next()))
	case reflect.Struct:
		for i := range v.NumField() {
			name := v.Type().String() + "." + v.Type().Field(i).Name
			if f.skip[name] {
				continue
			}
			field := v.Field(i)
			// Unexported fields are set through their address.
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			if f.inline[name] {
				field.Set(reflect.New(field.Type().Elem()))
				f.fill(field.Elem(), depth)
				continue
			}
			f.fill(field, depth)
		}
	case reflect.Pointer:
		if depth > 0 {
			v.Set(reflect.New(v.Type().Elem()))
			f.fill(v.Elem(), depth-1)
		}
	case reflect.Slice:
		if depth > 0 {
			v.Set(reflect.MakeSlice(v.Type(), 2, 2))
			for i := range 2 {
				f.fill(v.Index(i), depth-1)
			}
		}
	case reflect.Map:
		if depth > 0 {
			v.Set(reflect.MakeMap(v.Type()))
			for range 2 {
				key, elem := reflect.New(v.Type().Key()).Elem(), reflect.New(v.Type().Elem()).Elem()
				f.fill(key, depth-1)
				f.fill(elem, depth-1)
				v.SetMapIndex(key, elem)
			}
		}
	}
}

// Fields returns, by type name, the field names of every struct type t
// reaches through its fields, elements and pointers, but for the fields skip
// names and the types stop accepts (whose own test lists them).
func Fields(t reflect.Type, skip map[string]bool, stop func(reflect.Type) bool) map[string][]string {
	out := map[string][]string{}
	var walk func(t reflect.Type)
	walk = func(t reflect.Type) {
		switch t.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array:
			walk(t.Elem())
		case reflect.Map:
			walk(t.Key())
			walk(t.Elem())
		case reflect.Struct:
			name := t.String()
			if _, seen := out[name]; seen || stop(t) {
				return
			}
			out[name] = []string{}
			for i := range t.NumField() {
				field := t.Field(i)
				out[name] = append(out[name], field.Name)
				if !skip[name+"."+field.Name] {
					walk(field.Type)
				}
			}
		}
	}
	walk(t)
	return out
}
