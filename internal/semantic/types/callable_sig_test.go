package types

import (
	"reflect"
	"testing"
)

func TestCallableSignatureShape(t *testing.T) {
	for _, c := range []struct {
		in       string
		required int
		variadic bool
		byRef    uint64
		doc      string
	}{
		{"callable(int, string): bool", 2, false, 0, "callable(int, string): (bool)"},
		{"callable(int, string=): bool", 1, false, 0, "callable(int, string=): (bool)"},
		{"callable(int $a, string ...$rest): void", 1, true, 0, "callable(int, string...): (void)"},
		{"callable(int &$a): void", 1, false, 1, "callable(int&): (void)"},
		{"callable(string &...$rest): void", 0, true, 1, "callable(string&...): (void)"},
		{`Closure(int=): int`, 0, false, 0, `\Closure(int=): (int)`},
		{"callable(mixed=)", 0, false, 0, "callable(mixed=): (mixed)"},
	} {
		got := FromDoc(c.in, nil)
		atom := got.atoms[0]
		sig := got.callSigOf(atom)
		if sig == nil || sig.required != c.required || (sig.rest != nil) != c.variadic || sig.byRef != c.byRef {
			t.Errorf("%s: shape = %+v", c.in, sig)
		}
		if s := got.DocString(); s != c.doc {
			t.Errorf("%s: doc = %s, want %s", c.in, s, c.doc)
		}
		again := FromDoc(got.DocString(), nil)
		if !sameSig(sig, again.callSigOf(atom)) || again.DocString() != c.doc {
			t.Errorf("%s: round trip = %s %+v", c.in, again.DocString(), again.callSigOf(atom))
		}
	}
}

// A variadic parameter takes no position of its own in the signature's types.
func TestCallableVariadicParamType(t *testing.T) {
	got := FromDoc("callable(int, string ...$rest): void", nil)
	if p := CallableParam(got, 1); !p.IsUnknown() {
		t.Errorf("variadic parameter = %s", p.DocString())
	}
	if r := got.callSigOf("callable").rest; r == nil || r.DocString() != "string" {
		t.Errorf("rest = %v", r)
	}
}

func TestCallableWithoutShape(t *testing.T) {
	for _, in := range []string{"callable", "callable()", "callable(mixed): mixed", "callable(int"} {
		if sig := FromDoc(in, nil).callSigOf("callable"); sig != nil {
			t.Errorf("%s: shape = %+v", in, sig)
		}
	}
	if sig := (Type{}).callSigOf("callable"); sig != nil {
		t.Errorf("unknown type: shape = %+v", sig)
	}
	rebuilt := FromDoc("callable(int=): void", nil).WithTypeArgs("callable", []Type{Int})
	if sig := rebuilt.callSigOf("callable"); sig != nil {
		t.Errorf("rebuilt from types alone: shape = %+v", sig)
	}
	if s := rebuilt.DocString(); s != "callable(): (int)" {
		t.Errorf("rebuilt = %s", s)
	}
	noShape := Of("callable").WithTypeArgs("callable", []Type{Int, String})
	if s := noShape.DocString(); s != "callable(string): (int)" {
		t.Errorf("no shape = %s", s)
	}
}

func TestCallableShapeInUnions(t *testing.T) {
	a := FromDoc("callable(int=): void", nil)
	if sig := Union(a, FromDoc("callable(int=): void", nil)).callSigOf("callable"); sig == nil || sig.required != 0 {
		t.Errorf("same shapes: %+v", sig)
	}
	for _, other := range []string{"callable(int): void", "callable(string=): void"} {
		if got := Union(a, FromDoc(other, nil)); got.TypeArgs("callable") != nil {
			t.Errorf("with %s: kept %s", other, got.DocString())
		}
	}
	if got := FromDoc("callable(int=): void|callable(int): void", nil); got.TypeArgs("callable") != nil {
		t.Errorf("doc union kept %s", got.DocString())
	}
}

func TestCallableShapeWire(t *testing.T) {
	want := FromDoc("callable(int &$a, string=, float ...$r): void", nil)
	if got := roundTrip(t, want); !reflect.DeepEqual(got, want) {
		t.Errorf("decoded %s, want %s", got.DocString(), want.DocString())
	}
}

func TestSameSig(t *testing.T) {
	a, b := &callSig{required: 1}, &callSig{required: 1}
	if !sameSig(a, b) || !sameSig(nil, nil) || sameSig(a, nil) || sameSig(nil, b) || sameSig(a, &callSig{}) {
		t.Error("sameSig")
	}
	i, s := Int, String
	withInt, withInt2, withString := &callSig{rest: &i}, &callSig{rest: &i}, &callSig{rest: &s}
	if !sameSig(withInt, withInt2) || sameSig(withInt, withString) || sameSig(withInt, &callSig{}) {
		t.Error("sameSig on the variadic type")
	}
}

func TestParamSuffixBeyondMask(t *testing.T) {
	s := &callSig{required: 70, byRef: 1}
	if got := s.paramSuffix(64); got != "" {
		t.Errorf("parameter 64 = %q", got)
	}
}
