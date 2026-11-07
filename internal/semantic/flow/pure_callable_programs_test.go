package flow

import "testing"

func TestPureCallablePrograms(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($xs){foreach($xs as $x) if($x) return $x;return 0;}probe(f([1]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs,$x){foreach($xs as $x){if($x){return $x;}}return 0;}probe(f([1],9));`, constantScalar{k: 'i', i: 1}},
		{`function inc($x){return $x+1;}function f($xs){return array_map('inc',$xs)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 2}},
		{`function inc($x,$n=2){return $x+$n;}function f($xs){return array_map(callback:'inc',array:$xs)['x'];}probe(f(['x'=>1]));`, constantScalar{k: 'i', i: 3}},
		{`declare(strict_types=1);function inc(int $x):int{return $x+1;}function f($xs){return array_map('inc',$xs)[0];}probe(f(['1']));`, constantScalar{k: 'i', i: 2}},
		{`function keep($x){return $x!==null;}function f($xs){return count(array_filter($xs,'keep'));}probe(f([null,1,2]));`, constantScalar{k: 'i', i: 2}},
		{`function keep($x){return $x!==null;}function f($xs){return array_filter(...[$xs,'keep'])['x'];}probe(f([2=>null,'x'=>'yes']));`, constantScalar{k: 's', s: "yes"}},
		{`function keep($x){return $x!==null;}function f($xs){return count(array_filter($xs,'keep'));}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($xs){foreach($xs as $x){if($x>1){return $x;}}return 0;}probe(f([1,2,3]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){foreach($xs as $k=>$x){if($x>1){return $k;}}return 'none';}probe(f(['a'=>1,'01'=>2]));`, constantScalar{k: 's', s: "01"}},
		{`function f($xs){foreach($xs as $x){if($x['keep']){return $x['label'];}}return 'none';}probe(f([['keep'=>true,'label'=>'yes'],[]]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){foreach($xs as $x){if($x>1){return $x;}}return 0;}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($xs){foreach($xs as $x){if($x>1){return $x;}}return 0;}probe(f([0,1]));`, constantScalar{k: 'i'}},
		{`function f($xs){$out=[];foreach($xs as $x){if($x===null){continue;}$out[]=$x;}return $out;}probe(f([null,'yes'])[0]);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$sum=0;foreach($xs as $x){if($x===null){continue 1;}$sum+=$x;}return $sum;}probe(f([null,1,2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$out=[];foreach($xs as $x){if(!$x['keep']){continue;}$out[$x['id']]=$x['v'];}return $out;}probe(f([['keep'=>false],['keep'=>true,'id'=>'x','v'=>'yes']])['x']);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return (bool)$xs;}probe(f([1]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return (bool)$xs;}probe(f([]));`, constantScalar{k: 'b'}},
		{`function f($x){return (string)$x;}probe(f(12));`, constantScalar{k: 's', s: "12"}},
		{`function f($x){return (string)$x;}probe(f(null));`, constantScalar{k: 's'}},
		{`function f($x){return (int)$x;}probe(f('12'));`, constantScalar{k: 'i', i: 12}},
		{`function f($x){return (int)$x;}probe(f(true));`, constantScalar{k: 'i', i: 1}},
		{`function f($x){return (int)$x;}probe(f(null));`, constantScalar{k: 'i'}},
		{`function f($x){return (int)$x;}probe(f(12));`, constantScalar{k: 'i', i: 12}},
		{`function f($x){return (string)$x;}probe(f('yes'));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return $xs===[];}probe(f([]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return $xs===['x'=>1];}probe(f(['x'=>1]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return $xs!==['x'=>1];}probe(f(['x'=>2]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return $xs===['x'=>1,'y'=>2];}probe(f(['y'=>2,'x'=>1]));`, constantScalar{k: 'b'}},
		{`function f($xs){return $xs===['x'=>[1,'yes']];}probe(f(['x'=>[1,'yes']]));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return $xs===['x'=>[1]];}probe(f(['x'=>1]));`, constantScalar{k: 'b'}},
		{`function f($xs){return $xs===['x'=>1];}probe(f(['x'=>$_GET['x'],'y'=>2]));`, constantScalar{k: 'b'}},
		{`function f($xs){return $xs===['x'=>1,'y'=>2];}probe(f(['x'=>$_GET['x'],'y'=>3]));`, constantScalar{k: 'b'}},
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

func TestPureCallableProgramProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function project($x){return $x['v'];}function f($xs){return array_map('project',$xs);}probe(f([['v'=>'fixed','unused'=>$_GET['x']]])[0]);`, false},
		{`function project($x){return $x['v'];}function f($xs){return array_map('project',$xs);}probe(f([['v'=>$_GET['x']]])[0]);`, true},
		{`function keep($x){return $x['keep'];}function f($xs){return array_filter($xs,'keep');}probe(f(['x'=>['keep'=>true,'v'=>$_GET['x']]])['x']['v']);`, true},
		{`function f($xs){foreach($xs as $x){if($x['keep']){return $x['v'];}}return 'fixed';}probe(f([['keep'=>true,'v'=>$_GET['x']]]));`, true},
		{`function f($xs){foreach($xs as $x){if($x['keep']){return $x['v'];}}return 'fixed';}probe(f([['keep'=>false,'v'=>$_GET['x']]]));`, false},
		{`function f($xs){$out=[];foreach($xs as $x){if(!$x['keep']){continue;}$out[]=$x['v'];}return $out;}probe(f([['keep'=>true,'v'=>$_GET['x']]])[0]);`, true},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(value.Sources) != 0; got != tc.tainted {
			t.Errorf("%s tainted=%v want=%v", tc.source, got, tc.tainted)
		}
	}
}

func BenchmarkPureCallablePrograms(b *testing.B) {
	for _, tc := range []struct {
		name, source string
		want         constantScalar
	}{
		{"map", `function inc($x){return $x+1;}function f($xs){return array_map('inc',$xs)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 2}},
		{"find", `function f($xs){foreach($xs as $x){if($x>1){return $x;}}return 0;}probe(f([1,2,3]));`, constantScalar{k: 'i', i: 2}},
		{"continue", `function f($xs){$s=0;foreach($xs as $x){if($x===null){continue;}$s+=$x;}return $s;}probe(f([null,1,2]));`, constantScalar{k: 'i', i: 3}},
		{"cast", `function f($x){return (int)$x;}probe(f('12'));`, constantScalar{k: 'i', i: 12}},
		{"compare", `function f($xs){return $xs===['x'=>[1,'yes']];}probe(f(['x'=>[1,'yes']]));`, constantScalar{k: 'b', b: true}},
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
