package flow

import "testing"

func TestPureCheckedPrograms(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($a,$b){return $a/$b;}probe(f(6,3));`, constantScalar{k: 'i', i: 2}},
		{`function f($a,$b){return $a/$b;}probe(f(-6,3));`, constantScalar{k: 'i', i: -2}},
		{`function f($a,$b){return $a/$b;}probe(f(6,-3));`, constantScalar{k: 'i', i: -2}},
		{`function f($a,$b){return $a%$b;}probe(f(7,3));`, constantScalar{k: 'i', i: 1}},
		{`function f($a,$b){return $a%$b;}probe(f(-7,3));`, constantScalar{k: 'i', i: -1}},
		{`function f($a,$b){return $a%$b;}probe(f(7,-3));`, constantScalar{k: 'i', i: 1}},
		{`function f($a,$b){return $a%$b;}probe(f(-9223372036854775807-1,-1));`, constantScalar{k: 'i'}},
		{`function f($a,$b){return $a&$b;}probe(f(7,3));`, constantScalar{k: 'i', i: 3}},
		{`function f($a,$b){return $a|$b;}probe(f(4,3));`, constantScalar{k: 'i', i: 7}},
		{`function f($a,$b){return $a^$b;}probe(f(7,3));`, constantScalar{k: 'i', i: 4}},
		{`function f($a){return ~$a;}probe(f(1));`, constantScalar{k: 'i', i: -2}},
		{`function f($a,$b){return $a<<$b;}probe(f(2,3));`, constantScalar{k: 'i', i: 16}},
		{`function f($a,$b){return $a<<$b;}probe(f(-1,63));`, constantScalar{k: 'i', i: -9223372036854775808}},
		{`function f($a,$b){return $a<<$b;}probe(f(7,0));`, constantScalar{k: 'i', i: 7}},
		{`function f($a,$b){return $a>>$b;}probe(f(-8,2));`, constantScalar{k: 'i', i: -2}},
		{`function f($a,$b){return $a>>$b;}probe(f(-8,63));`, constantScalar{k: 'i', i: -1}},
		{`function f($a,$b){return $a>>$b;}probe(f(8,63));`, constantScalar{k: 'i'}},
		{`function f($x){return str_replace('a','b',$x);}probe(f('a,a'));`, constantScalar{k: 's', s: "b,b"}},
		{`function f($x){return str_replace(search:'a',replace:'',subject:$x);}probe(f('aaa'));`, constantScalar{k: 's'}},
		{`function f($x){return str_replace(...['aa','a',$x]);}probe(f('aaa'));`, constantScalar{k: 's', s: "aa"}},
		{`function f($x){return str_replace('é','x',$x);}probe(f('aéb'));`, constantScalar{k: 's', s: "axb"}},
		{`function f($x){return str_replace('missing','new',$x);}probe(f('same'));`, constantScalar{k: 's', s: "same"}},
		{`function f($x,$n){return str_repeat($x,$n);}probe(f('ab',2));`, constantScalar{k: 's', s: "abab"}},
		{`function f($x,$n){return str_repeat(string:$x,times:$n);}probe(f('ab',0));`, constantScalar{k: 's'}},
		{`function f($x,$n){return str_repeat(...[$x,$n]);}probe(f('',9223372036854775807));`, constantScalar{k: 's'}},
		{`function f($xs){return array_reverse($xs)[0];}probe(f([9=>'a',5=>'b']));`, constantScalar{k: 's', s: "b"}},
		{`function f($xs){return array_reverse($xs,true)[5];}probe(f([9=>'a',5=>'b']));`, constantScalar{k: 's', s: "b"}},
		{`function f($xs){return array_keys(array_reverse(array:$xs,preserve_keys:false))[0];}probe(f([9=>'a','x'=>'b']));`, constantScalar{k: 's', s: "x"}},
		{`function f($xs){return array_reverse(...[$xs])['01'];}probe(f(['01'=>'yes',2=>'no']));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return count(array_reverse($xs));}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($x){switch($x){case 1:return 'one';default:return 'other';}}probe(f(1));`, constantScalar{k: 's', s: "one"}},
		{`function f($x){switch($x){default:return 'other';case 1:return 'one';}}probe(f(1));`, constantScalar{k: 's', s: "one"}},
		{`function f($x){switch($x){case 1:return 'one';default:return 'other';}}probe(f(2));`, constantScalar{k: 's', s: "other"}},
		{`function f($x){switch($x){case 1:{return 'one';}default:{return 'other';}}}probe(f(1));`, constantScalar{k: 's', s: "one"}},
		{`function f($x){switch($x){case -1:return 'negative';case 1:return 'one';default:return 'other';}}probe(f(-1));`, constantScalar{k: 's', s: "negative"}},
		{`function f($x){switch($x){case 1:return 'first';case 1:return 'second';default:return 'other';}}probe(f(1));`, constantScalar{k: 's', s: "first"}},
		{`function f($x,$bad){switch($x){case 1:return 'one';case $bad['missing']:return 'later';default:return 'other';}}probe(f(1,[]));`, constantScalar{k: 's', s: "one"}},
		{`function f($x,$bad){switch($x){case 1:return 'one';case $bad:return 'later';default:return 'other';}}probe(f(1,$_GET['bad']));`, constantScalar{k: 's', s: "one"}},
		{`function f($x,$bad){switch($x){case 1:return 'one';default:return $bad['missing'];}}probe(f(1,[]));`, constantScalar{k: 's', s: "one"}},
		{`function f($x){switch($x){case 1:return;default:return 'other';}}probe(f(1));`, constantScalar{k: 'n'}},
	} {
		e, file := summarized(t, tc.source)
		for _, env := range []*Env{e, NewLocal(file, e.types, e.snapshot, Options{}, nil)} {
			got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
			if !known || got != tc.want {
				t.Errorf("%s got=%+v known=%v want=%+v", tc.source, got, known, tc.want)
			}
		}
	}
}

func TestPureCheckedProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($xs){return array_reverse($xs);}probe(f(['fixed',$_GET['x']])[0]);`, true},
		{`function f($xs){return array_reverse($xs);}probe(f(['fixed',$_GET['x']])[1]);`, false},
		{`function f($x,$v){switch($x){case 1:return $v;default:return 'fixed';}}probe(f(1,$_GET['x']));`, true},
		{`function f($x,$v){switch($x){case 1:return 'fixed';default:return $v;}}probe(f(1,$_GET['x']));`, false},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if (len(value.Sources) != 0) != tc.tainted {
			t.Errorf("provenance %s", tc.source)
		}
	}
}

func BenchmarkPureCheckedPrograms(b *testing.B) {
	for _, tc := range []struct{ name, source string }{
		{"division", `function f($a,$b){return $a/$b;}probe(f(6,3));`},
		{"bitwise", `function f($a,$b){return $a<<$b;}probe(f(2,3));`},
		{"replace", `function f($x){return str_replace('a','b',$x);}probe(f('a,a'));`},
		{"repeat", `function f($x,$n){return str_repeat($x,$n);}probe(f('ab',2));`},
		{"reverse", `function f($xs){return array_reverse($xs)[0];}probe(f([9=>'a',5=>'b']));`},
		{"switch", `function f($x){switch($x){case 1:return 'one';default:return 'other';}}probe(f(1));`},
	} {
		b.Run(tc.name, func(b *testing.B) {
			e, file := summarized(b, tc.source)
			call := argument(callsNamed(file, "probe")[0])
			if _, known := knownScalar(e.Value(call)); !known {
				b.Fatal("unknown result")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				NewLocal(file, e.types, e.snapshot, Options{}, nil).Value(call)
			}
		})
	}
}
