package flow

import (
	"reflect"
	"testing"
	"unsafe"

	"custos/internal/php/syntax"
)

// extField sets one field of a value's extension and reads it back.
type extField struct {
	name       string
	set, clear func(*Value)
	held       func(Value) bool
}

func extFields() []extField {
	record, literal, returned := &resourceReturn{}, &constantScalar{}, &callableValue{}
	targets, interval, elements := &callableSet{}, &intInterval{}, &elementSet{}
	origins, tested, unless := &originSet{}, &syntax.Binary{}, &syntax.Binary{}
	return []extField{
		{"record", func(v *Value) { v.setRecord(record) }, func(v *Value) { v.setRecord(nil) }, func(v Value) bool { return v.record() == record }},
		{"literal", func(v *Value) { v.setLiteral(literal) }, func(v *Value) { v.setLiteral(nil) }, func(v Value) bool { return v.literal() == literal }},
		{"returned", func(v *Value) { v.setReturned(returned) }, func(v *Value) { v.setReturned(nil) }, func(v Value) bool { return v.returned() == returned }},
		{"targets", func(v *Value) { v.setTargets(targets) }, func(v *Value) { v.setTargets(nil) }, func(v Value) bool { return v.targets() == targets }},
		{"interval", func(v *Value) { v.setInterval(interval) }, func(v *Value) { v.setInterval(nil) }, func(v Value) bool { return v.interval() == interval }},
		{"elements", func(v *Value) { v.setElements(elements) }, func(v *Value) { v.setElements(nil) }, func(v Value) bool { return v.elements() == elements }},
		{"origins", func(v *Value) { v.setOrigins(origins) }, func(v *Value) { v.setOrigins(nil) }, func(v Value) bool { return v.origins() == origins }},
		{"tested", func(v *Value) { v.setTested(tested) }, func(v *Value) { v.setTested(nil) }, func(v Value) bool { return v.tested() == tested }},
		{"unless", func(v *Value) { v.setUnlessFacts(unless, true, true, true, true) }, func(v *Value) { v.setUnlessFacts(nil, false, false, false, false) }, func(v Value) bool {
			return v.unless() == unless && v.unlessFalse() && v.unlessNull() && v.skippedFalse() && v.skippedNull()
		}},
		{"length", func(v *Value) { v.edit(func(x valueExt) valueExt { x.length, x.lengthKnown = 3, true; return x }) }, func(v *Value) { v.setLengthKnown(false) }, func(v Value) bool { return v.length() == 3 }},
		{"storedIdentity", func(v *Value) { v.setStoredIdentity(3) }, func(v *Value) { v.setStoredIdentity(0) }, func(v Value) bool { return v.storedIdentity() == 3 }},
		{"lengthKnown", func(v *Value) { v.setLengthKnown(true) }, func(v *Value) { v.setLengthKnown(false) }, func(v Value) bool { return v.lengthKnown() }},
		{"headerMask", func(v *Value) { v.setHeaderMask(3) }, func(v *Value) { v.setHeaderMask(0) }, func(v Value) bool { return v.headerMask() == 3 }},
		{"upload", func(v *Value) { v.setUpload(3) }, func(v *Value) { v.setUpload(0) }, func(v Value) bool { return v.upload() == 3 }},
		{"scalars", func(v *Value) { v.setScalars(true) }, func(v *Value) { v.setScalars(false) }, func(v Value) bool { return v.scalars() }},
		{"matched", func(v *Value) { v.setMatched(true) }, func(v *Value) { v.setMatched(false) }, func(v Value) bool { return v.matched() }},
		{"replaced", func(v *Value) { v.setReplaced(true) }, func(v *Value) { v.setReplaced(false) }, func(v Value) bool { return v.replaced() }},
		{"entry", func(v *Value) { v.setEntry(true) }, func(v *Value) { v.setEntry(false) }, func(v Value) bool { return v.entry() }},
		{"truth", func(v *Value) { v.setTruth(3) }, func(v *Value) { v.setTruth(0) }, func(v Value) bool { return v.truth() == 3 }},
	}
}

// A setter never writes an extension another copy shares, leaves it alone
// when the field already holds what it sets, and drops an extension left
// with no field set.
func TestValueExtensionCopyOnWrite(t *testing.T) {
	fields := extFields()
	for i, field := range fields {
		var v Value
		if field.clear(&v); v.ext != nil || field.held(v) {
			t.Fatalf("%s: clearing a value without an extension made one", field.name)
		}
		field.set(&v)
		ext := v.ext
		if field.set(&v); !field.held(v) || v.ext != ext {
			t.Fatalf("%s: setting the held field again replaced the extension", field.name)
		}
		// A copy shares the extension until one of them is set.
		other := fields[(i+1)%len(fields)]
		w := v
		if other.set(&w); v.ext != ext || other.held(v) || !other.held(w) || !field.held(w) {
			t.Fatalf("%s: setting %s on a copy wrote the shared extension", field.name, other.name)
		}
		if field.clear(&v); v.ext != nil || field.held(v) || w.ext == nil || !field.held(w) {
			t.Fatalf("%s: clearing the only field kept an extension or wrote the copy's", field.name)
		}
		if field.clear(&w); w.ext == nil || field.held(w) || !other.held(w) {
			t.Fatalf("%s: clearing one of two fields lost the other", field.name)
		}
	}
}

// Every field of valueExt has a case above: one added without its accessors
// fails here.
func TestValueExtensionFieldsCovered(t *testing.T) {
	covered := map[string]bool{"unlessFalse": true, "unlessNull": true, "skippedFalse": true, "skippedNull": true}
	for _, field := range extFields() {
		covered[field.name] = true
	}
	ext := reflect.TypeFor[valueExt]()
	for i := range ext.NumField() {
		if !covered[ext.Field(i).Name] {
			t.Errorf("valueExt.%s has no accessor case", ext.Field(i).Name)
		}
	}
}

// sameExt tells apart two extensions that differ in any one field, but for
// a length that is not known.
func TestSameExtComparesEveryField(t *testing.T) {
	kind := reflect.TypeFor[valueExt]()
	for i := range kind.NumField() {
		field := kind.Field(i)
		a := valueExt{lengthKnown: true}
		b := a
		cell := reflect.NewAt(field.Type, unsafe.Pointer(reflect.ValueOf(&b).Elem().Field(i).UnsafeAddr())).Elem()
		switch field.Type.Kind() {
		case reflect.Bool:
			cell.SetBool(!cell.Bool())
		case reflect.Int64:
			cell.SetInt(1)
		case reflect.Uint8, reflect.Uint32:
			cell.SetUint(1)
		case reflect.Pointer:
			cell.Set(reflect.New(field.Type.Elem()))
		default:
			t.Fatalf("valueExt.%s has a kind this test does not set", field.Name)
		}
		if sameExt(&a, &b) || sameExt(&b, &a) {
			t.Errorf("sameExt ignores valueExt.%s", field.Name)
		}
	}
	a, b := valueExt{length: 1}, valueExt{length: 2}
	if !sameExt(&a, &b) || !sameExt(&a, nil) {
		t.Error("sameExt compares a length that is not known")
	}
}

// A length that is not known is dropped, so an extension holding nothing
// else goes; a join keeps a length only where both paths know the same one.
func TestUnknownLengthIsZero(t *testing.T) {
	var v Value
	v.edit(func(x valueExt) valueExt { x.length, x.lengthKnown = 4, true; return x })
	if v.setLengthKnown(false); v.ext != nil || v.length() != 0 {
		t.Fatal("an unknown length kept its extension")
	}
	if rare(valueExt{length: 4}) != nil {
		t.Fatal("rare kept an unknown length")
	}
	known, other := v, v
	known.edit(func(x valueExt) valueExt { x.length, x.lengthKnown = 4, true; return x })
	other.edit(func(x valueExt) valueExt { x.length, x.lengthKnown = 5, true; return x })
	if m := merge(known, other); m.ext != nil || !reflect.DeepEqual(m, merge(v, v)) {
		t.Fatal("a join of different lengths kept one")
	}
	if m := merge(known, known); m.ext != known.ext || m.length() != 4 {
		t.Fatal("a join of the same length lost it")
	}
}

// A Value is copied everywhere and stored by the hundred thousand: it
// stays within one 64-byte allocation class on 64-bit platforms.
func TestValueSize(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 8 && unsafe.Sizeof(Value{}) > 64 {
		t.Fatalf("Value is %d bytes", unsafe.Sizeof(Value{}))
	}
}

// Equal values compare equal whether or not they share an extension.
func TestEqualComparesExtensions(t *testing.T) {
	a := Value{Complete: true}
	a.setTruth(1)
	b := Value{Complete: true}
	b.setTruth(1)
	if a.ext == b.ext || !equal(a, b) || !reflect.DeepEqual(a, b) {
		t.Fatal("equal extensions compare unequal")
	}
	if b.setTruth(2); equal(a, b) || equal(a, Value{Complete: true}) || equal(Value{Complete: true}, a) {
		t.Fatal("different extensions compare equal")
	}
	if !equal(Value{Complete: true}, Value{Complete: true}) {
		t.Fatal("values without an extension compare unequal")
	}
}

// BenchmarkMergeValues joins what two paths hold of a binding, with and
// without an extension, as frame joins do.
func BenchmarkMergeValues(b *testing.B) {
	sources := []Source{{Kind: "get"}}
	plain := Value{Complete: true, Sources: sources, Safe: HTML, plain: true, deep: true}
	ranged := plain
	ranged.setInterval(&intInterval{lo: 1, hi: 2})
	ranged.setTruth(1)
	for _, bench := range []struct {
		name string
		a, b Value
	}{{"plain", plain, plain}, {"extended", ranged, ranged}, {"mixed", plain, ranged}} {
		b.Run(bench.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if m := merge(bench.a, bench.b); !m.Complete {
					b.Fatal("lost fact")
				}
			}
		})
	}
}

// BenchmarkFrameJoin joins a branch that changed some bindings back into
// the frame it split from.
func BenchmarkFrameJoin(b *testing.B) {
	f := frame{}
	f.resetVars()
	names := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	for i, name := range names {
		v := Value{Complete: true, Identity: uint32(i + 1), Sources: []Source{{Kind: "get", Parameter: i}}}
		if i%2 == 0 {
			v.setInterval(&intInterval{lo: int64(i), hi: int64(i)})
		}
		f.setVar(name, v)
	}
	b.ReportAllocs()
	for b.Loop() {
		joined := clone(&f)
		branch := clone(&f)
		for i, name := range names[:4] {
			v := branch.vars.value(name)
			v.Identity = uint32(100 + i)
			branch.setVar(name, v)
		}
		if changed, _ := joined.absorbAt(branch, true); !changed {
			b.Fatal("join lost the branch")
		}
	}
}

// BenchmarkEraseRecordResources drops the resources of a record whose
// entries each hold one, as a summary that cannot describe them does.
func BenchmarkEraseRecordResources(b *testing.B) {
	set := &elementSet{whole: true, items: map[string]Value{}}
	for i := range 8 {
		entry := Value{Complete: true}
		entry.setRecord(&resourceReturn{parameter: i})
		entry.setStoredIdentity(uint32(i + 1))
		set.put(string(rune('a'+i)), entry)
	}
	var value Value
	value.setElements(set)
	value.setStoredIdentity(9)
	b.ReportAllocs()
	for b.Loop() {
		if erased := eraseRecordResources(value); erased.elements().hasResources {
			b.Fatal("kept a resource")
		}
	}
}
