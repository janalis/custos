package flow

import (
	"reflect"
	"testing"
	"unsafe"

	"custos/internal/php/syntax"
)

// leanValue must keep every field of Value but the references leanOf
// checks: a field added to Value is either one of them or kept by both
// conversions.
func TestLeanValueKeepsEveryOtherField(t *testing.T) {
	left := map[string]bool{"record": true, "literal": true, "returned": true, "targets": true, "interval": true, "elements": true, "origins": true, "tested": true, "unless": true}
	full, lean := reflect.TypeFor[Value](), reflect.TypeFor[leanValue]()
	kept := 0
	for i := range full.NumField() {
		field := full.Field(i)
		if left[field.Name] {
			continue
		}
		other, ok := lean.FieldByName(field.Name)
		if !ok || other.Type != field.Type {
			t.Fatalf("leanValue lacks %s %s", field.Name, field.Type)
		}
		kept++
	}
	if kept != lean.NumField() || kept+len(left) != full.NumField() {
		t.Fatalf("leanValue has %d fields, want %d", lean.NumField(), kept)
	}
	// Every kept field survives the round trip.
	var v Value
	fields := reflect.ValueOf(&v).Elem()
	for i := range full.NumField() {
		if left[full.Field(i).Name] {
			continue
		}
		f := settable(fields.Field(i))
		switch f.Kind() {
		case reflect.Bool:
			f.SetBool(true)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32:
			f.SetUint(uint64(i + 1))
		case reflect.Int64:
			f.SetInt(int64(i + 1))
		case reflect.Slice:
			f.Set(reflect.MakeSlice(f.Type(), 1, 1))
		default:
			f.Set(reflect.ValueOf(&syntax.Literal{}))
		}
	}
	l, ok := leanOf(&v)
	if back := l.value(); !ok || !reflect.DeepEqual(back, v) {
		t.Fatalf("round trip changed the value: %+v", back)
	}
	for name := range left {
		w := v
		f := settable(reflect.ValueOf(&w).Elem().FieldByName(name))
		f.Set(reflect.New(f.Type().Elem()))
		if _, ok := leanOf(&w); ok {
			t.Fatalf("a value with %s became lean", name)
		}
	}
}

// settable returns the addressable field f, exported or not, for writing.
func settable(f reflect.Value) reflect.Value {
	return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
}

func TestValueStoreMovesBetweenLeanAndFull(t *testing.T) {
	store := newValueStore(8)
	exprs := make([]syntax.Expr, 8)
	for i := range exprs {
		exprs[i] = &syntax.Literal{}
	}
	full := func(id uint32) Value {
		return Value{Identity: id, interval: &intInterval{lo: int64(id), hi: int64(id)}}
	}
	for i, expr := range exprs {
		if i%2 == 0 {
			store.set(expr, full(uint32(i+1)))
		} else {
			store.set(expr, Value{Identity: uint32(i + 1)})
		}
	}
	// Full blocks grow from one entry.
	if len(store.fulls) != 3 || len(store.fulls[0]) != 1 || len(store.fulls[1]) != 2 || len(store.fulls[2]) != 4 || len(store.leans) != 1 {
		t.Fatalf("blocks: %d full, %d lean", len(store.fulls), len(store.leans))
	}
	for i, expr := range exprs {
		v := store.at(expr)
		if v.Identity != uint32(i+1) || (v.interval != nil) != (i%2 == 0) {
			t.Fatalf("entry %d: %+v", i, v)
		}
	}
	// A lean entry that becomes full moves; a full entry stays full.
	store.set(exprs[1], full(100))
	store.set(exprs[0], Value{Identity: 200})
	store.set(exprs[3], Value{Identity: 300})
	if v := store.at(exprs[1]); v.Identity != 100 || v.interval == nil || v.interval.lo != 100 {
		t.Fatalf("moved entry: %+v", v)
	}
	if v := store.at(exprs[0]); v.Identity != 200 || v.interval != nil {
		t.Fatalf("full entry: %+v", v)
	}
	if store.at(exprs[3]).Identity != 300 || store.at(exprs[5]).Identity != 6 || store.at(exprs[4]).interval.lo != 5 {
		t.Fatal("updates disturbed other entries")
	}
	// Past the first blocks, full blocks keep the limit width.
	many := newValueStore(0)
	exprs = make([]syntax.Expr, 300)
	for i := range exprs {
		exprs[i] = &syntax.Literal{}
		many.set(exprs[i], full(uint32(i+1)))
	}
	for i, expr := range exprs {
		if v := many.at(expr); v.Identity != uint32(i+1) || v.interval.lo != int64(i+1) {
			t.Fatalf("full entry %d: %+v", i, v)
		}
	}
	if last := many.fulls[len(many.fulls)-1]; len(last) != valueBlockLimit || len(many.fulls) != 2+(300-3+valueBlockLimit-1)/valueBlockLimit {
		t.Fatalf("%d full blocks", len(many.fulls))
	}
}

// BenchmarkMixedValueStore stores what a scope records: mostly lean values,
// some full ones.
func BenchmarkMixedValueStore(b *testing.B) {
	expressions := make([]syntax.Expr, 128)
	for i := range expressions {
		expressions[i] = &syntax.Literal{LitKind: syntax.LitInt}
	}
	interval := &intInterval{lo: 1, hi: 2}
	b.ReportAllocs()
	for b.Loop() {
		store := newValueStore(len(expressions))
		for i, expr := range expressions {
			v := Value{Complete: true, Identity: uint32(i + 1), Expr: expr}
			if i%16 == 0 {
				v.interval = interval
			}
			store.set(expr, v)
		}
		for i, expr := range expressions {
			if store.at(expr).Identity != uint32(i+1) {
				b.Fatal("lost value")
			}
		}
	}
}

func TestValueStoreGrowthAndCopies(t *testing.T) {
	var store valueStore
	if value, known := store.get(nil); known || value.Complete {
		t.Fatal("missing entry became known")
	}
	expressions := make([]syntax.Expr, 1000)
	for i := range expressions {
		expressions[i] = &syntax.Literal{LitKind: syntax.LitInt}
		store.set(expressions[i], Value{Complete: true, Identity: uint32(i + 1)})
	}
	for i, expr := range expressions {
		if value, known := store.get(expr); !known || !value.Complete || value.Identity != uint32(i+1) {
			t.Fatalf("entry %d corrupted by growth", i)
		}
	}
	old := store.at(expressions[0])
	store.set(expressions[0], Value{Complete: true, Identity: 2000})
	if old.Identity != 1 || store.at(expressions[0]).Identity != 2000 {
		t.Fatal("updating storage changed a retained copy")
	}
}

func TestFrameRepeatedFactsRemainIsolated(t *testing.T) {
	original := frame{}
	original.resetVars()
	value := Value{Complete: true, Safe: HTML}
	original.setVar("value", value)
	branch := clone(&original)
	branch.setVar("value", value)
	branch.setVar("value", Value{Complete: true, Sources: []Source{{Kind: "request", Parameter: -1}}})
	if original.vars.value("value").Safe != HTML || tainted(original.vars.value("value")) || !tainted(branch.vars.value("value")) {
		t.Fatal("a repeated fact broke branch isolation")
	}
	original.setVar("other", value)
	if _, known := branch.vars["other"]; known {
		t.Fatal("new binding leaked across branches")
	}
}

func BenchmarkSharedFrameRepeatedFact(b *testing.B) {
	f := frame{}
	f.resetVars()
	value := Value{Complete: true, Safe: HTML}
	f.setVar("value", value)
	b.ReportAllocs()
	for b.Loop() {
		branch := clone(&f)
		branch.setVar("value", value)
		if !branch.vars.value("value").Complete {
			b.Fatal("lost fact")
		}
	}
}

func TestValueStoreReservesOnlyReachedPayloads(t *testing.T) {
	for _, capacity := range []int{0, 1, 10000} {
		store := newValueStore(capacity)
		if len(store.leans) != 0 || len(store.fulls) != 0 {
			t.Fatal("unreached expressions allocated payloads")
		}
		store.set(nil, Value{Complete: true})
		if len(store.leans) != 1 || len(store.fulls) != 0 || len(store.leans[0]) > valueBlockLimit || !store.at(nil).Complete {
			t.Fatal("payload allocation exceeded one block")
		}
	}
}

func BenchmarkExpressionValueStore(b *testing.B) {
	expressions := make([]syntax.Expr, 128)
	for i := range expressions {
		expressions[i] = &syntax.Literal{LitKind: syntax.LitInt}
	}
	b.ReportAllocs()
	for b.Loop() {
		store := newValueStore(len(expressions))
		for i, expr := range expressions {
			store.set(expr, Value{Complete: true, Identity: uint32(i + 1)})
		}
		for i, expr := range expressions {
			if store.at(expr).Identity != uint32(i+1) {
				b.Fatal("lost value")
			}
		}
	}
}

func BenchmarkFlowAllocationHints(b *testing.B) {
	e, file := environment(b, `class R{function f($x){if($x){return $this->g($x);}return 1;}function g($v){return $v;}}`, nil)
	var scope syntax.Node
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if method, ok := node.(*syntax.Method); ok && method.Name.Value == "f" {
			scope = method
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		values, before := e.expressionSizes(scope)
		if values == 0 || before == 0 {
			b.Fatal("missing runtime nodes")
		}
	}
}

func TestValueStoreDistinctEqualSpans(t *testing.T) {
	first, second := &syntax.Literal{}, &syntax.Literal{}
	store := newValueStore(2)
	store.set(first, Value{Identity: 1})
	store.set(second, Value{Identity: 2})
	if store.at(first).Identity != 1 || store.at(second).Identity != 2 {
		t.Fatal("equal spans aliased")
	}
}

func TestValueStoreNarrowBlocks(t *testing.T) {
	for _, capacity := range []int{1, 2, 3, 4} {
		store := newValueStore(capacity)
		expressions := make([]syntax.Expr, 19)
		for i := range expressions {
			expressions[i] = &syntax.Literal{}
			store.set(expressions[i], Value{Identity: uint32(i + 1)})
		}
		for i, expr := range expressions {
			if store.at(expr).Identity != uint32(i+1) {
				t.Fatalf("capacity %d entry %d", capacity, i)
			}
			store.set(expr, Value{Identity: uint32(i + 20)})
			if store.at(expr).Identity != uint32(i+20) {
				t.Fatal("update failed")
			}
		}
	}
}

func TestValueStoreTracksSimpleVariablesOnce(t *testing.T) {
	store := newValueStore(4)
	variable := &syntax.Variable{Name: "parameter"}
	store.set(variable, Value{Identity: 1})
	store.set(variable, Value{Identity: 2})
	store.set(&syntax.Variable{NameExpr: &syntax.Literal{}}, Value{})
	store.set(&syntax.Literal{}, Value{})
	store.set(nil, Value{})
	if len(store.variables) != 1 || store.variables[0] != variable || store.at(variable).Identity != 2 {
		t.Fatal("variable iteration changed")
	}
}
