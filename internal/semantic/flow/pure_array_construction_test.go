package flow

import "testing"

var constructionCases = []struct {
	name, source string
	want         constantScalar
}{
	{"fill", `function f($n,$v){return array_fill(3,$n,$v)[4];}probe(f(2,'yes'));`, constantScalar{k: 's', s: "yes"}},
	{"fill_keys_order", `function f($n){return array_keys(array_fill(start_index:3,count:$n,value:'v'))[1];}probe(f(2));`, constantScalar{k: 'i', i: 4}},
	{"fill_max", `function f($v){return array_fill(...[9223372036854775807,1,$v])[9223372036854775807];}probe(f('yes'));`, constantScalar{k: 's', s: "yes"}},
	{"fill_record", `function f($v){return array_fill(0,2,$v)[1]['v'];}probe(f(['v'=>'yes']));`, constantScalar{k: 's', s: "yes"}},
	{"fill_keys", `function f($k,$v){return array_fill_keys($k,$v)['x'];}probe(f(['x','y'],'yes'));`, constantScalar{k: 's', s: "yes"}},
	{"fill_keys_duplicates", `function f($k){return count(array_fill_keys($k,1));}probe(f(['2',2,'x']));`, constantScalar{k: 'i', i: 2}},
	{"fill_keys_numeric", `function f($k){return array_keys(array_fill_keys(value:1,keys:$k))[0];}probe(f(['2',2]));`, constantScalar{k: 'i', i: 2}},
	{"fill_keys_string", `function f($k){return array_keys(array_fill_keys(...[$k,1]))[0];}probe(f(['02',2,'02']));`, constantScalar{k: 's', s: "02"}},
	{"fill_keys_empty", `function f($k,$v){return count(array_fill_keys($k,$v));}probe(f([],1));`, constantScalar{k: 'i'}},
	{"flip", `function f($xs){return array_flip($xs)['x'];}probe(f([9=>'x']));`, constantScalar{k: 'i', i: 9}},
	{"flip_last", `function f($xs){return array_flip($xs)['x'];}probe(f(['a'=>'x','b'=>'x']));`, constantScalar{k: 's', s: "b"}},
	{"flip_order", `function f($xs){return array_keys(array_flip(array:$xs))[0];}probe(f(['a'=>'x','b'=>'y','c'=>'x']));`, constantScalar{k: 's', s: "x"}},
	{"flip_numeric", `function f($xs){return array_flip(...[$xs])[2];}probe(f([9=>'2',5=>2]));`, constantScalar{k: 'i', i: 5}},
	{"flip_empty", `function f($xs){return count(array_flip($xs));}probe(f([]));`, constantScalar{k: 'i'}},
	{"chunk", `function f($xs){return array_chunk($xs,2)[1][0];}probe(f([9=>'a',3=>'b',8=>'c']));`, constantScalar{k: 's', s: "c"}},
	{"chunk_preserve", `function f($xs){return array_chunk($xs,2,true)[1][8];}probe(f([9=>'a',3=>'b',8=>'c']));`, constantScalar{k: 's', s: "c"}},
	{"chunk_named", `function f($xs){return array_chunk(preserve_keys:false,length:1,array:$xs)[1][0];}probe(f(['a'=>'x','b'=>'y']));`, constantScalar{k: 's', s: "y"}},
	{"chunk_huge", `function f($xs){return count(array_chunk(...[$xs,9223372036854775807]));}probe(f([1,2,3]));`, constantScalar{k: 'i', i: 1}},
	{"chunk_empty", `function f($xs){return count(array_chunk($xs,1));}probe(f([]));`, constantScalar{k: 'i'}},
	{"sum", `function f($xs){return array_sum($xs);}probe(f([2,3,-1]));`, constantScalar{k: 'i', i: 4}},
	{"sum_empty", `function f($xs){return array_sum(array:$xs);}probe(f([]));`, constantScalar{k: 'i'}},
	{"sum_negative", `function f($xs){return array_sum(...[$xs]);}probe(f([-2,-3,1]));`, constantScalar{k: 'i', i: -4}},
	{"sum_max", `function f($xs){return array_sum($xs);}probe(f([9223372036854775807,-1]));`, constantScalar{k: 'i', i: 9223372036854775806}},
	{"sum_min", `function f($xs){return array_sum($xs);}probe(f([-9223372036854775807,-1]));`, constantScalar{k: 'i', i: -9223372036854775807 - 1}},
}

func TestPureArrayConstruction(t *testing.T) {
	for _, tc := range constructionCases {
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

func TestPureArrayConstructionProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($v){return array_fill(3,2,$v);}probe(f($_GET['v'])[4]);`, true},
		{`function f($v){return array_fill(3,2,$v);}probe(f(['v'=>'fixed','other'=>$_GET['v']])[4]['v']);`, false},
		{`function f($k,$v){return array_fill_keys($k,$v);}probe(f(['x'],$_GET['v'])['x']);`, true},
		{`function f($k,$v){return array_fill_keys($k,$v);}probe(f([], $_GET['v']));`, false},
		{`function f($xs){return array_chunk($xs,1);}probe(f(['fixed',$_GET['v']])[0][0]);`, false},
		{`function f($xs){return array_chunk($xs,1);}probe(f(['fixed',$_GET['v']])[1][0]);`, true},
		{`function f($row){return array_sum($row['nums']);}probe(f(['nums'=>[1,2],'unused'=>$_GET['v']]));`, false},
		{`function f($xs){return array_chunk($xs,1,true);}probe(f(['a'=>'fixed','b'=>$_GET['v']])[1]['b']);`, true},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(value.Sources) != 0; got != tc.tainted {
			t.Errorf("%s tainted=%v", tc.source, got)
		}
	}
}

func BenchmarkPureArrayConstruction(b *testing.B) {
	for _, tc := range constructionCases {
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
