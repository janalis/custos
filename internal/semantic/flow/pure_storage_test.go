package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/semantic/types"
)

func TestPureBindingStorage(t *testing.T) {
	// More than eight declarations exercise lookup and reassignment in both
	// the small binding set and its overflow map.
	var params, args strings.Builder
	for i := range 16 {
		if i != 0 {
			params.WriteByte(',')
			args.WriteByte(',')
		}
		fmt.Fprintf(&params, "$p%d", i)
		fmt.Fprintf(&args, "%d", i)
	}
	source := fmt.Sprintf(`function many(%s){$p12=$p15;$p0=$p12;return $p0+$p12;}probe(many(%s));`, params.String(), args.String())
	e, f := summarized(t, source)
	got, known := knownScalar(e.Value(argument(callsNamed(f, "probe")[0])))
	if !known || got.k != 'i' || got.i != 30 {
		t.Fatalf("overflow bindings lost: %+v known=%v", got, known)
	}
}

func TestPureCoercionsDoNotMutateBindings(t *testing.T) {
	e, _ := summarized(t, `probe(1);`)
	args := []Value{scalarValue(constantScalar{k: 's', s: "2"}), scalarValue(constantScalar{k: 's', s: "3"})}
	contract := &pureContract{params: []pureParam{{}, {typ: types.FromDoc("int", nil)}}}
	bound, known := e.pureBind(contract, args, true)
	if !known {
		t.Fatal("valid scalar binding rejected")
	}
	second, ok := knownScalar(bound[1])
	if !ok || second.k != 'i' || second.i != 3 {
		t.Fatal("coercion lost")
	}
	for _, arg := range args {
		v, ok := knownScalar(arg)
		if !ok || v.k != 's' {
			t.Fatal("binding modified caller argument")
		}
	}
	// Coercion of the first slot also leaves an unchanged later slot intact.
	contract.params[0].typ = types.FromDoc("int", nil)
	bound, known = e.pureBind(contract, args, true)
	first, ok := knownScalar(bound[0])
	if !known || !ok || first.k != 'i' || first.i != 2 {
		t.Fatal("multiple coercions lost")
	}
}

func TestPureOuterAndInvocationSlotBoundaries(t *testing.T) {
	var outer, inner, args strings.Builder
	for i := range 64 {
		if i != 0 {
			outer.WriteByte(',')
			inner.WriteByte(',')
			args.WriteByte(',')
		}
		fmt.Fprintf(&outer, "int $p%d", i)
		fmt.Fprintf(&inner, "$q%d", i)
		fmt.Fprintf(&args, "%d", i)
	}
	source := fmt.Sprintf(`function factory(%s){return fn(%s)=>$p63+$q63;}$cb=factory(%s);probe($cb(%s));`, outer.String(), inner.String(), args.String(), args.String())
	e, f := summarized(t, source)
	got, known := knownScalar(e.Value(argument(callsNamed(f, "probe")[0])))
	if !known || got.k != 'i' || got.i != 126 {
		t.Fatalf("outer and invocation slots overlapped: %+v known=%v", got, known)
	}
}
