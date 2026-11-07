package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

var bindingCases = []struct {
	name, source string
	want         constantScalar
}{
	{"column", `function f($xs){return array_column($xs,'v','id')[2];}probe(f([['v'=>'a','id'=>'2'],['v'=>'b','id'=>2]]));`, constantScalar{k: 's', s: "b"}},
	{"column_order", `function f($xs){return array_keys(array_column($xs,'v','id'))[0];}probe(f([['v'=>1,'id'=>'x'],['v'=>2,'id'=>'y'],['v'=>3,'id'=>'x']]));`, constantScalar{k: 's', s: "x"}},
	{"combine", `function f($k,$v){return array_combine($k,$v)[2];}probe(f([9=>'2',4=>2],['a','b']));`, constantScalar{k: 's', s: "b"}},
	{"combine_order", `function f($k,$v){return array_keys(array_combine(values:$v,keys:$k))[0];}probe(f(['x','y','x'],[1,2,3]));`, constantScalar{k: 's', s: "x"}},
	{"replace", `function f($a,$b){return array_replace($a,$b)[9];}probe(f([9=>'a','x'=>'c'],[9=>'b',3=>'d']));`, constantScalar{k: 's', s: "b"}},
	{"replace_order", `function f($a,$b){return array_keys(array_replace($a,$b))[2];}probe(f([9=>'a','x'=>'c'],[9=>'b',3=>'d']));`, constantScalar{k: 'i', i: 3}},
	{"local", `function f($xs){[$a,$b]=$xs;return $a+$b;}probe(f([2,3]));`, constantScalar{k: 'i', i: 5}},
	{"local_snapshot", `function f($xs){[$xs,$b]=$xs;return $xs+$b;}probe(f([2,3]));`, constantScalar{k: 'i', i: 5}},
	{"local_keyed", `function f($xs){['2'=>$a,'x'=>$b]=$xs;return $a.$b;}probe(f([2=>'a','x'=>'b']));`, constantScalar{k: 's', s: "ab"}},
	{"local_hole", `function f($xs){list(,$a)=$xs;return $a;}probe(f([0,'yes']));`, constantScalar{k: 's', s: "yes"}},
	{"local_twice", `function f($xs){[$a,$b]=$xs;[$b,$a]=[$a,$b];return $a-$b;}probe(f([2,3]));`, constantScalar{k: 'i', i: 1}},
	{"prepare", `function f($xs){$s=0;foreach($xs as $x){$n=$x['n'];$s+=$n;}return $s;}probe(f([['n'=>2],['n'=>3]]));`, constantScalar{k: 'i', i: 5}},
	{"prepare_build", `function f($xs){$s=[];foreach($xs as $x){$n=trim($x);$s[]=$n;}return $s;}probe(f([' a ',' b '])[1]);`, constantScalar{k: 's', s: "b"}},
	{"prepare_keyed", `function f($xs){$s=[];foreach($xs as $k=>$x){$n=trim($x);$s[$k]=$n;}return $s;}probe(f(['x'=>' a '])['x']);`, constantScalar{k: 's', s: "a"}},
	{"prepare_destruct", `function f($xs){$s=0;foreach($xs as [$a,$b]){$n=$a+$b;$s+=$n;}return $s;}probe(f([[2,3],[4,5]]));`, constantScalar{k: 'i', i: 14}},
	{"prepare_pipeline", `function f($xs){$a=[];foreach($xs as $x){$n=trim($x);$a[]=$n;}$s='';foreach($a as $x){$m=$x.'!';$s.=$m;}return $s;}probe(f([' a ',' b ']));`, constantScalar{k: 's', s: "a!b!"}},
	{"local_loop", `function f($xs){[$a,$b]=$xs;$s=0;foreach($a as $x){$n=$x+$b;$s+=$n;}return $s;}probe(f([[1,2],3]));`, constantScalar{k: 'i', i: 9}},
}

func TestPureBindings(t *testing.T) {
	for _, tc := range bindingCases {
		t.Run(tc.name, func(t *testing.T) {
			e, file := summarized(t, tc.source)
			if Extract(file, e.types.Index, e.types.PHP, nil, Options{}).Summaries["f"].pure == nil {
				t.Fatal("missing contract")
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

func TestPureBindingRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($xs){return array_column($xs,'v','id')[0];}probe(f([['v'=>1]]));`,
		`function f($xs){return array_column($xs,'v',null)[0];}probe(f([['v'=>1,'id'=>0]]));`,
		`function f($xs){return array_column($xs,'v','id')[0];}probe(f([['v'=>1,'id'=>false]]));`,
		`function f($k,$v){return array_combine($k,$v)[0];}probe(f([],[]));`,
		`function f($k,$v){return array_combine($k,$v)[0];}probe(f([0],[1,2]));`,
		`function f($k,$v){return array_combine($k,$v)[0];}probe(f([null],[1]));`,
		`function f($a){return array_replace(...[$a,$a])[0];}probe(f([1]));`,
		`function f($a){return array_replace(array:$a,replacements:$a)[0];}probe(f([1]));`,
		`function f($a){return array_replace($a,$a,$a)[0];}probe(f([1]));`,
		`function f($xs){[$a,$b]=$xs;return 1;}probe(f([2]));`,
		`function f($xs){[$a,$a]=$xs;return 1;}probe(f([2,3]));`,
		`function f($xs){[[$a]]=$xs;return 1;}probe(f([[2]]));`,
		`function f($xs){[&$a]=$xs;return 1;}probe(f([2]));`,
		`function f($xs){[$GLOBALS]=$xs;return 1;}probe(f([2]));`,
		`function f($xs){[$_GET]=$xs;return 1;}probe(f([2]));`,
		`function f($xs){[$this]=$xs;return 1;}probe(f([2]));`,
		`function f($xs){[$obj->x]=$xs;return 1;}probe(f([2]));`,
		`function f($xs){[$a]=opaque();return 1;}probe(f([2]));`,
		`function f($xs){[$a]=$xs;return 1;}probe(f([new stdClass]));`,
		`function f($xs){[$a]=$xs['missing'];return 1;}probe(f([]));`,
		`function f($xs){$s=0;foreach($xs as $x){$n=$x['missing'];$s+=1;}return $s;}probe(f([[]]));`,
		`function f($xs,$n){$s=0;foreach($xs as $x){$n=$x;$s+=$n;}return $s;}probe(f([1],2));`,
		`function f($xs){$n=0;$s=0;foreach($xs as $x){$n=$x;$s+=$n;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$x;$s+=1;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$x=1;$s+=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $n=>$x){$n=1;$s+=$n;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$n=opaque();$s+=1;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$n=$x;if($n){$s+=1;}}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$n=$x;$s+=1;}return $n;}probe(f([1]));`,
	} {
		e, file := summarized(t, source)
		contract := Extract(file, e.types.Index, e.types.PHP, nil, Options{}).Summaries["f"].pure
		if contract == nil {
			continue
		}
		var args []Value
		for _, item := range callsNamed(file, "f")[0].Args.Args {
			args = append(args, e.Value(item.(*syntax.Arg).Value))
		}
		if got, known := e.runPure(contract, args, &scopeResult{}, MaxDepth); known {
			t.Errorf("%s invented %+v", source, got)
		}
	}
}

func BenchmarkPureBindings(b *testing.B) {
	for _, tc := range bindingCases {
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
