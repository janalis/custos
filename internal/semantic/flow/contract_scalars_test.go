package flow

import (
	"math"
	"testing"
)

func TestPrimitiveConversions(t *testing.T) {
	for _, tc := range []struct {
		value  constantScalar
		target string
		weak   bool
		want   constantScalar
		known  bool
	}{
		{constantScalar{k: 'i', i: 1}, "float", false, constantScalar{k: 'f', f: 1}, true},
		{constantScalar{k: 's', s: "00"}, "int", false, constantScalar{}, false},
		{constantScalar{k: 's', s: "00"}, "int", true, constantScalar{k: 'i'}, true},
		{constantScalar{k: 's', s: "-01"}, "int", true, constantScalar{k: 'i', i: -1}, true},
		{constantScalar{k: 's', s: "+"}, "int", true, constantScalar{}, false},
		{constantScalar{k: 's', s: "1x"}, "int", true, constantScalar{}, false},
		{constantScalar{k: 's', s: "9223372036854775808"}, "int", true, constantScalar{}, false},
		{constantScalar{k: 'b', b: true}, "int", true, constantScalar{k: 'i', i: 1}, true},
		{constantScalar{k: 'b'}, "int", true, constantScalar{k: 'i'}, true},
		{constantScalar{k: 'f', f: 2}, "int", true, constantScalar{k: 'i', i: 2}, true},
		{constantScalar{k: 'f', f: 2.5}, "int", true, constantScalar{}, false},
		{constantScalar{k: 'b', b: true}, "float", true, constantScalar{k: 'f', f: 1}, true},
		{constantScalar{k: 'b'}, "float", true, constantScalar{k: 'f'}, true},
		{constantScalar{k: 'i', i: 1 << 54}, "float", true, constantScalar{}, false},
		{constantScalar{k: 'i', i: -12}, "string", true, constantScalar{k: 's', s: "-12"}, true},
		{constantScalar{k: 'b'}, "string", true, constantScalar{k: 's'}, true},
		{constantScalar{k: 'b', b: true}, "string", true, constantScalar{k: 's', s: "1"}, true},
		{constantScalar{k: 'f'}, "string", true, constantScalar{k: 's', s: "0"}, true},
		{constantScalar{k: 'f', f: math.Copysign(0, -1)}, "string", true, constantScalar{k: 's', s: "-0"}, true},
		{constantScalar{k: 'f', f: 1.1}, "string", true, constantScalar{}, false},
		{constantScalar{k: 's', s: "0"}, "bool", true, constantScalar{k: 'b'}, true},
		{constantScalar{k: 'n'}, "bool", true, constantScalar{}, false},
		{constantScalar{k: 'i'}, "object", true, constantScalar{}, false},
	} {
		got, known := convertScalar(tc.value, tc.target, tc.weak)
		if known != tc.known || known && !sameScalar(&got, &tc.want) {
			t.Fatalf("%+v -> %s: %+v %v", tc.value, tc.target, got, known)
		}
	}
	for _, value := range []constantScalar{{k: 'i', i: 3}, {k: 'f', f: 2.5}, {k: 's', s: "a'\\b"}, {k: 'b'}, {k: 'b', b: true}, {k: 'n'}} {
		got, known := knownScalar(Value{Complete: true, Expr: scalarExpression(value)})
		if !known || !sameScalar(&got, &value) {
			t.Fatalf("round trip %+v: %+v", value, got)
		}
	}
	a, b := constantScalar{k: 'f'}, constantScalar{k: 'f', f: math.Copysign(0, -1)}
	if sameScalar(&a, &b) || !sameScalar(&a, &a) || sameScalar(&a, nil) {
		t.Fatal("scalar equality lost signed zero or nil")
	}
}

func TestCallbackCoercionContext(t *testing.T) {
	for _, tc := range []struct {
		source string
		empty  bool
	}{
		{`$xs=array_filter(['00'],fn(int $v)=>$v);probe($xs);`, true},
		{`$xs=array_filter(['00'],fn(int $v)=>!$v);probe($xs);`, false},
		{`declare(strict_types=1);$xs=array_filter(['00'],fn(int $v)=>$v);probe($xs);`, true},
		{`$xs=array_filter([1],fn($v):int=>'00');probe($xs);`, true},
		{`declare(strict_types=1);$xs=array_filter([1],fn($v):int=>'00');probe($xs);`, false},
		{`$xs=array_filter([1],fn($v):string=>-0.0);probe($xs);`, false},
		{`$xs=array_filter([1],fn($v):string=>0);probe($xs);`, true},
		{`$xs=array_filter([1],fn($v):bool=>'0');probe($xs);`, true},
		{`$xs=array_filter([1],fn($v):float=>false);probe($xs);`, true},
		{`$xs=array_filter([1],fn($v):int=>0.5);probe($xs);`, false},
		{`function reject($v){if($v){return false;}return false;}$xs=array_filter([1],'reject');probe($xs);`, true},
		{`function factory(){return function($v){if($v){return false;}return false;};}$xs=array_filter([1],factory());probe($xs);`, true},
		{`function reject($v):int{return '00';}$xs=array_filter([1],'reject');probe($xs);`, true},
		{`function factory(){return fn($v):int=>'00';}$xs=array_filter([1],factory());probe($xs);`, true},
	} {
		e, f := summarized(t, tc.source)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		empty := v.elements != nil && v.elements.whole && !v.elements.optional && len(v.elements.items) == 0
		if empty != tc.empty {
			t.Fatalf("%s: %+v", tc.source, v)
		}
	}
}
