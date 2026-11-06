package flow

import (
	"testing"

	phpversion "custos/internal/php/version"
)

var textPaddingCases = []struct {
	name, source string
	want         constantScalar
}{
	{"lower", `function f($s){return strtolower($s);}probe(f('AbC-Z'));`, constantScalar{k: 's', s: "abc-z"}},
	{"upper", `function f($s){return strtoupper(string:$s);}probe(f("aÿz"));`, constantScalar{k: 's', s: "AÿZ"}},
	{"ucfirst", `function f($s){return ucfirst($s);}probe(f('abc'));`, constantScalar{k: 's', s: "Abc"}},
	{"lcfirst", `function f($s){return lcfirst(...[$s]);}probe(f('ABC'));`, constantScalar{k: 's', s: "aBC"}},
	{"ucfirst_empty", `function f($s){return ucfirst($s);}probe(f(''));`, constantScalar{k: 's'}},
	{"ltrim", `function f($s){return ltrim($s);}probe(f(" \t a "));`, constantScalar{k: 's', s: "a "}},
	{"rtrim_chars", `function f($s){return rtrim($s,'/');}probe(f('a//'));`, constantScalar{k: 's', s: "a"}},
	{"trim_chars", `function f($s){return trim(characters:'x',string:$s);}probe(f('xxaxx'));`, constantScalar{k: 's', s: "a"}},
	{"trim_bytes", `function f($s){return trim($s,"\xff");}probe(f("\xffa\xff"));`, constantScalar{k: 's', s: "a"}},
	{"trim_all", `function f($s){return rtrim($s,'ab');}probe(f('abab'));`, constantScalar{k: 's'}},
	{"pad_right", `function f($s){return str_pad($s,5,'-');}probe(f('ab'));`, constantScalar{k: 's', s: "ab---"}},
	{"pad_default", `function f($s){return str_pad($s,4);}probe(f('ab'));`, constantScalar{k: 's', s: "ab  "}},
	{"pad_left", `function f($s){return str_pad($s,5,'-',STR_PAD_LEFT);}probe(f('ab'));`, constantScalar{k: 's', s: "---ab"}},
	{"pad_both", `function f($s){return str_pad($s,7,'xy',STR_PAD_BOTH);}probe(f('ab'));`, constantScalar{k: 's', s: "xyabxyx"}},
	{"pad_short", `function f($s){return str_pad($s,3,'-');}probe(f('abcdef'));`, constantScalar{k: 's', s: "abcdef"}},
	{"array_pad", `function f($xs){return array_pad($xs,3,'z')[2];}probe(f(['a']));`, constantScalar{k: 's', s: "z"}},
	{"array_pad_left", `function f($xs){return array_pad($xs,-3,'z')[2];}probe(f(['a']));`, constantScalar{k: 's', s: "a"}},
	{"array_pad_renumber", `function f($xs){return array_keys(array_pad($xs,-4,'z'))[3];}probe(f(['k'=>'a',5=>'b']));`, constantScalar{k: 'i', i: 2}},
	{"array_pad_string_key", `function f($xs){return array_keys(array_pad($xs,4,'z'))[0];}probe(f(['k'=>'a',5=>'b']));`, constantScalar{k: 's', s: "k"}},
	{"array_pad_unchanged", `function f($xs){return array_keys(array_pad($xs,1,'z'))[0];}probe(f([5=>'a']));`, constantScalar{k: 'i', i: 5}},
	{"array_pad_empty", `function f($xs){return count(array_pad($xs,2,'z'));}probe(f([]));`, constantScalar{k: 'i', i: 2}},
	{"range", `function f($n){return range(1,$n)[2];}probe(f(4));`, constantScalar{k: 'i', i: 3}},
	{"range_step", `function f($n){return count(range(0,$n,2));}probe(f(6));`, constantScalar{k: 'i', i: 4}},
	{"range_down", `function f($n){return range(5,$n,2)[2];}probe(f(1));`, constantScalar{k: 'i', i: 1}},
	{"range_single", `function f($n){return count(range($n,$n));}probe(f(3));`, constantScalar{k: 'i', i: 1}},
	{"range_named", `function f($n){return range(end:$n,start:-2)[0];}probe(f(0));`, constantScalar{k: 'i', i: -2}},
}

func TestPureTextAndPadding(t *testing.T) {
	for _, tc := range textPaddingCases {
		t.Run(tc.name, func(t *testing.T) {
			e, file := summarized(t, tc.source)
			contract, _ := e.contract("f")
			if contract.pure == nil {
				t.Fatal("missing pure contract")
			}
			for _, env := range []*Env{e, NewLocal(file, e.types, e.snapshot, Options{}, nil)} {
				got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
				if !known || got != tc.want {
					t.Fatalf("got=%+v known=%v want=%+v", got, known, tc.want)
				}
			}
		})
	}
}

func TestPureTextAndPaddingRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($s){return strtolower($s);}probe(f(1));`,
		`function f($s){return strtolower($s);}probe(f(str_repeat('a',4097)));`,
		`function f($s){return rtrim($s,'a..c');}probe(f('abc'));`,
		`function f($s){return rtrim($s,1);}probe(f('abc'));`,
		`function f($s){return rtrim($s,$s.str_repeat('x',300));}probe(f('abc'));`,
		`function f($s){return str_pad($s,5,'');}probe(f('ab'));`,
		`function f($s){return str_pad($s,5,'-',3);}probe(f('ab'));`,
		`function f($s){return str_pad($s,5,'-',-1);}probe(f('ab'));`,
		`function f($s){return str_pad($s,'5');}probe(f('ab'));`,
		`function f($s){return str_pad($s,5000);}probe(f('ab'));`,
		`function f($s,$n){return str_pad($s,$n);}probe(f('ab',$_GET['n']));`,
		`function f($s,$p){return str_pad($s,5,$p['missing']);}probe(f('ab',[]));`,
		`function f($s,$p){return str_pad($s,5,'-',$p['missing']);}probe(f('ab',[]));`,
		`function f($s,$p){return str_pad($s,$p['missing']);}probe(f('ab',[]));`,
		`function f($s){return str_pad($s);}probe(f('ab'));`,
		`function f($xs){return array_pad($xs,'3','z')[2];}probe(f(['a']));`,
		`function f($xs){return array_pad($xs,65,'z')[2];}probe(f(['a']));`,
		`function f($xs){return array_pad($xs,3,new stdClass)[2];}probe(f(['a']));`,
		`function f($xs,$v){return array_pad($xs,3,$v['missing'])[2];}probe(f(['a'],[]));`,
		`function f($xs){return array_pad($xs,3,'z')[2];}probe(f(opaque()));`,
		`function f($xs){return array_pad($xs,3)[2];}probe(f(['a']));`,
		`function f($n){return range(1,$n,0)[0];}probe(f(4));`,
		`function f($n){return range(1,$n,-1)[0];}probe(f(4));`,
		`function f($n){return range(1,$n,5)[0];}probe(f(4));`,
		`function f($n){return range($n,$n,1)[0];}probe(f(4));`,
		`function f($n){return range(1,$n)[0];}probe(f(1.5));`,
		`function f($n){return range(1,$n)[0];}probe(f('4'));`,
		`function f($n){return range(1,$n)[0];}probe(f(100));`,
		`function f($n){return range($n,1)[0];}probe(f(100));`,
		`function f($n){return range($n,0)[0];}probe(f(-9223372036854775807));`,
		`function f($n){return range(0,$n)[0];}probe(f(9223372036854775807));`,
		`function f($n){return range(1,$n,$n['missing'])[0];}probe(f([]));`,
		`function f($n){return range(1,$n['missing'])[0];}probe(f([]));`,
		`function f($n){return range($n)[0];}probe(f(1));`,
	} {
		e, file := summarized(t, source)
		if _, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("accepted: %s", source)
		}
	}
}

func TestPureTextCaseVersionGate(t *testing.T) {
	base, file := environment(t, `function f($s){return strtolower($s);}probe(f('AbC'));`, nil)
	base.types.PHP = phpversion.PHP81
	snapshot := NewSnapshot()
	for range 3 {
		snapshot = NewSnapshot(Extract(file, base.types.Index, base.types.PHP, snapshot, Options{}))
	}
	e := New(file, base.types, snapshot, Options{})
	if _, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
		t.Fatal("locale-dependent case change accepted before PHP 8.2")
	}
}

func TestPureTextAndPaddingProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($s){return strtoupper($s);}probe(f($_GET['v']));`, true},
		{`function f($p){return str_pad('ab',5,$p);}probe(f($_GET['v']));`, true},
		{`function f($v){return array_pad(['a'],3,$v)[0];}probe(f($_GET['v']));`, false},
		{`function f($v){return array_pad(['a'],3,$v)[2];}probe(f($_GET['v']));`, true},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(value.Sources) != 0; got != tc.tainted {
			t.Errorf("%s tainted=%v", tc.source, got)
		}
	}
}

func BenchmarkPureTextAndPadding(b *testing.B) {
	for _, tc := range textPaddingCases {
		b.Run(tc.name, func(b *testing.B) {
			e, file := summarized(b, tc.source)
			call := argument(callsNamed(file, "probe")[0])
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				env := NewLocal(file, e.types, e.snapshot, Options{}, nil)
				env.Value(call)
			}
		})
	}
}
