package flow

import "testing"

func TestPureRecordPrograms(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($xs){return $xs?'yes':'no';}probe(f([1]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return $xs?'yes':'no';}probe(f([]));`, constantScalar{k: 's', s: "no"}},
		{`function f($xs){if($xs){return 'yes';}return 'no';}probe(f([1]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return !$xs;}probe(f([]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return $xs&&[1];}probe(f([1]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return $xs||[];}probe(f([]));`, constantScalar{k: 'b'}},
		{`function f($xs){return $xs&&$xs['missing'];}probe(f([]));`, constantScalar{k: 'b'}},
		{`function f($xs){return empty($xs['x']['y']);}probe(f([]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return empty($xs['x']);}probe(f(['x'=>null]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return empty($xs['x']);}probe(f(['x'=>[1]]));`, constantScalar{k: 'b'}},
		{`function f($x){return empty($x);}probe(f('0'));`, constantScalar{k: 'b', b: true}},
		{`function f($key,$v){return [$key=>$v]['x'];}probe(f('x','yes'));`, constantScalar{k: 's', s: "yes"}},
		{`function f($key,$v){return [$key=>$v][2];}probe(f('2','yes'));`, constantScalar{k: 's', s: "yes"}},
		{`function f($key){return array_key_first([$key=>1]);}probe(f('01'));`, constantScalar{k: 's', s: "01"}},
		{`function f($key){return [$key=>'first','x'=>'last']['x'];}probe(f('x'));`, constantScalar{k: 's', s: "last"}},
		{`function f($key){return [$key=>'first','last'][3];}probe(f(2));`, constantScalar{k: 's', s: "last"}},
		{`function f($key,$xs){return [$key=>'first',...$xs]['y'];}probe(f('x',['y'=>'yes']));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $k=>$x){$out[$k]=$x;}return $out;}probe(f(['x'=>'yes'])['x']);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[$x['id']]=$x['v'];}return $out;}probe(f([['id'=>'x','v'=>'yes']])['x']);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $x){if($x['keep']){$out[$x['id']]=$x['v'];}}return $out;}probe(f([['keep'=>false],['keep'=>true,'id'=>'x','v'=>'yes']])['x']);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[$x['id']]=$x['v'];}return array_values($out)[0];}probe(f([['id'=>'x','v'=>'first'],['id'=>'y','v'=>'second'],['id'=>'x','v'=>'last']]));`, constantScalar{k: 's', s: "last"}},
		{`function f($xs){$out=[];foreach($xs as $k=>$x){$out[$k]=$x;}return array_key_first($out);}probe(f([-2=>1]));`, constantScalar{k: 'i', i: -2}},
		{`function f($xs){return (['x'=>'left']+$xs)['x'];}probe(f(['x'=>'right']));`, constantScalar{k: 's', s: "left"}},
		{`function f($xs){return (['x'=>'left']+$xs)['y'];}probe(f(['y'=>'yes']));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return array_key_last([2=>'left']+$xs);}probe(f([2=>'right',5=>'yes']));`, constantScalar{k: 'i', i: 5}},
		{`function f($xs){$sum=0;foreach($xs as $x){$sum+=$x;}$again=0;foreach($xs as $x){$again+=$x;}return $sum+$again;}probe(f([1,2]));`, constantScalar{k: 'i', i: 6}},
		{`function f($xs){$out=[];foreach($xs as $x){if($x!==null){$out[]=$x;}}$sum=0;foreach($out as $x){$sum+=$x;}return $sum;}probe(f([null,1,2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x;}$sum=0;foreach($out as $x){$sum+=$x;}return $sum;}probe(f([]));`, constantScalar{k: 'i'}},
	} {
		e, file := summarized(t, tc.source)
		for _, env := range []*Env{e, NewLocal(file, e.types, e.snapshot, Options{}, nil)} {
			got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
			if !known || got != tc.want {
				t.Errorf("%s: got=%+v known=%v want=%+v", tc.source, got, known, tc.want)
			}
		}
	}
}

func TestPureRecordProgramProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($k,$v){return [$k=>$v,'fixed'=>'yes'];}probe(f('x',$_GET['x'])['fixed']);`, false},
		{`function f($k,$v){return [$k=>$v];}probe(f('x',$_GET['x'])['x']);`, true},
		{`function f($xs){$out=[];foreach($xs as $x){$out[$x['id']]=$x['v'];}return $out;}probe(f([['id'=>'x','v'=>$_GET['x']]])['x']);`, true},
		{`function f($xs){$out=[];foreach($xs as $x){$out[$x['id']]=$x['v'];}return $out;}probe(f([['id'=>'x','v'=>$_GET['x']],['id'=>'x','v'=>'fixed']])['x']);`, false},
		{`function f($xs){return ['x'=>'left']+$xs;}probe(f(['x'=>$_GET['x']])['x']);`, false},
		{`function f($xs){return ['x'=>'left']+$xs;}probe(f(['y'=>$_GET['y']])['y']);`, true},
		{`function f($xs){$a=[];foreach($xs as $x){if($x['keep']){$a[]=$x['v'];}}$b=[];foreach($a as $v){$b[]=$v;}return $b;}probe(f([['keep'=>true,'v'=>$_GET['v']]])[0]);`, true},
		{`function f($xs){$a=[];foreach($xs as $x){if($x['keep']){$a[]=$x['v'];}}$b=[];foreach($a as $v){$b[]=$v;}return $b;}probe(f([['keep'=>true,'v'=>'fixed','unused'=>$_GET['v']]])[0]);`, false},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(value.Sources) != 0; got != tc.tainted {
			t.Errorf("%s: tainted=%v want=%v", tc.source, got, tc.tainted)
		}
	}
}

func BenchmarkPureRecordPrograms(b *testing.B) {
	for _, tc := range []struct {
		name, source string
		want         constantScalar
	}{
		{"truth", `function f($xs){return $xs?'yes':'no';}probe(f([1]));`, constantScalar{k: 's', s: "yes"}},
		{"computed", `function f($k,$v){return [$k=>$v]['x'];}probe(f('x','yes'));`, constantScalar{k: 's', s: "yes"}},
		{"keyed", `function f($xs){$out=[];foreach($xs as $k=>$x){$out[$k]=$x;}return $out;}probe(f(['x'=>'yes'])['x']);`, constantScalar{k: 's', s: "yes"}},
		{"union", `function f($xs){return (['x'=>'left']+$xs)['y'];}probe(f(['y'=>'yes']));`, constantScalar{k: 's', s: "yes"}},
		{"pipeline", `function f($xs){$out=[];foreach($xs as $x){if($x!==null){$out[]=$x;}}$sum=0;foreach($out as $x){$sum+=$x;}return $sum;}probe(f([null,1,2]));`, constantScalar{k: 'i', i: 3}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			e, file := summarized(b, tc.source)
			call := argument(callsNamed(file, "probe")[0])
			if got, known := knownScalar(e.Value(call)); !known || got != tc.want {
				b.Fatalf("result=%+v known=%v", got, known)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				env := NewLocal(file, e.types, e.snapshot, Options{}, nil)
				env.Value(call)
			}
		})
	}
}
