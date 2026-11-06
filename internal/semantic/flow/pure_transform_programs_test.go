package flow

import "testing"

func TestPureTransformPrograms(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($xs){$s=0;foreach($xs as [2=>$v]){$s+=$v;}return $s;}probe(f([[2=>7]]));`, constantScalar{k: 'i', i: 7}},

		{`function f($xs){return array_map(fn($x)=>$x+1,$xs)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 2}},
		{`declare(strict_types=1);function f($xs){return array_map(fn(int $x):int=>$x+1,$xs)[0];}probe(f(['1']));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return array_map(static fn($x,$extra=2)=>$x+$extra,$xs)['x'];}probe(f(['x'=>1]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){return count(array_filter($xs,fn($x)=>$x!==null));}probe(f([null,1]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){return count(array_filter($xs,fn($k)=>$k==='x',2));}probe(f(['x'=>1,'y'=>2]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){return count(array_filter($xs,fn($v,$k)=>$k==='x'&&$v>1,1));}probe(f(['x'=>2,'y'=>2]));`, constantScalar{k: 'i', i: 1}},
		{`function f($a,$b){return array_map(fn($x,$y)=>$x+$y,$a,$b)[0];}probe(f([1],[2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){return array_map(callback:fn($x)=>$x+1,array:$xs)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return array_map(...[fn($x)=>$x+1,$xs])[0];}probe(f([1]));`, constantScalar{k: 'i', i: 2}},
		{`function add($c,$x){return $c+$x;}function f($xs){return array_reduce($xs,'add',0);}probe(f([1,2]));`, constantScalar{k: 'i', i: 3}},
		{`function add($c,$x){return $c+$x;}function f($xs){return array_reduce(initial:7,callback:'add',array:$xs);}probe(f([]));`, constantScalar{k: 'i', i: 7}},
		{`function add(int $c,int $x):int{return $c+$x;}function f($xs){return array_reduce(...[$xs,'add',0]);}probe(f(['1','2']));`, constantScalar{k: 'i', i: 3}},
		{`function add($c,$x,$z=1){return $c+$x+$z;}function f($xs){return array_reduce($xs,'add',0);}probe(f([1,2]));`, constantScalar{k: 'i', i: 5}},
		{`function cat($c,$x){return $c.$x;}function f($xs){return array_reduce($xs,'cat','');}probe(f(['a','b']));`, constantScalar{k: 's', s: "ab"}},
		{`function f($xs){return array_slice($xs,1,1)[0];}probe(f([9=>1,5=>2]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return array_slice($xs,1,1,true)[5];}probe(f([9=>1,5=>2]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return array_slice($xs,-2,-1)['b'];}probe(f(['a'=>1,'b'=>2,'c'=>3]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return array_slice($xs,0,null)[1];}probe(f(['x'=>1,5=>2,6=>3]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){return array_slice($xs,-1)[0];}probe(f([1,2]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return count(array_slice($xs,99));}probe(f([1,2]));`, constantScalar{k: 'i'}},
		{`function f($xs){return count(array_slice($xs,0,-9223372036854775807));}probe(f([1,2]));`, constantScalar{k: 'i'}},
		{`function f($xs){return array_slice(offset:-9223372036854775807,array:$xs,length:9223372036854775807)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){return array_keys($xs,1,true)[0];}probe(f(['a'=>'1','b'=>1]));`, constantScalar{k: 's', s: "b"}},
		{`function f($xs){return array_keys($xs,[1],true)[0];}probe(f(['a'=>[2],2=>[1]]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return count(array_keys(array:$xs,filter_value:1,strict:true));}probe(f(['a'=>1,'b'=>1]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return count(array_keys(...[$xs,1,true]));}probe(f([1,2,1]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){$s=0;foreach($xs as [$a,$b]){$s+=$a+$b;}return $s;}probe(f([[1,2],[3,4]]));`, constantScalar{k: 'i', i: 10}},
		{`function f($xs){$s=0;foreach($xs as list($a,$b)){$s+=$a+$b;}return $s;}probe(f([[1,2],[3,4]]));`, constantScalar{k: 'i', i: 10}},
		{`function f($xs){$s=0;foreach($xs as ['v'=>$v]){$s+=$v;}return $s;}probe(f([['v'=>2],['v'=>3]]));`, constantScalar{k: 'i', i: 5}},
		{`function f($xs){$s=0;foreach($xs as [,$v]){$s+=$v;}return $s;}probe(f([[1,2]]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){$s=0;foreach($xs as $k=>[$v]){$s+=$v+$k;}return $s;}probe(f([2=>[3]]));`, constantScalar{k: 'i', i: 5}},
		{`function f($xs){$out=[];foreach($xs as ['id'=>$id,'v'=>$v]){$out[$id]=$v;}return $out['x'];}probe(f([['id'=>'x','v'=>'yes']]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$s=0;foreach($xs as [$v]){if($v===null){continue;}$s+=$v;}return $s;}probe(f([[1],[null],[2]]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$s=0;foreach($xs as [$v]){if($v===null){break;}$s+=$v;}return $s;}probe(f([[1],[null],[2]]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){$s=0;foreach($xs as [$v]){$s+=$v;}return $s+1;}probe(f([]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){$a=[];foreach($xs as [$v]){$a[]=$v;}$b=0;foreach($a as $v){$b+=$v;}return $b;}probe(f([[1],[2]]));`, constantScalar{k: 'i', i: 3}},
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

func TestPureTransformProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($xs){return array_map(fn($r)=>$r['v'],$xs);}probe(f([['v'=>'fixed','unused'=>$_GET['x']]])[0]);`, false},
		{`function f($xs){return array_map(fn($r)=>$r['v'],$xs);}probe(f([['v'=>$_GET['x']]])[0]);`, true},
		{`function f($xs){return array_slice($xs,1);}probe(f([$_GET['x'],'fixed'])[0]);`, false},
		{`function f($xs){return array_slice($xs,1);}probe(f(['fixed',$_GET['x']])[0]);`, true},
		{`function last($carry,$x){return $x;}function f($xs){return array_reduce($xs,'last','fixed');}probe(f([$_GET['x']]));`, true},
		{`function last($carry,$x){return 'fixed';}function f($xs){return array_reduce($xs,'last','fixed');}probe(f([$_GET['x'],'fixed']));`, false},
		{`function f($xs){$out=[];foreach($xs as ['v'=>$v,'unused'=>$unused]){$out[]=$v;}return $out;}probe(f([['v'=>'fixed','unused'=>$_GET['x']]])[0]);`, false},
		{`function f($xs){$out=[];foreach($xs as ['v'=>$v]){$out[]=$v;}return $out;}probe(f([['v'=>$_GET['x']]])[0]);`, true},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if (len(value.Sources) != 0) != tc.tainted {
			t.Errorf("provenance: %s", tc.source)
		}
	}
}

func BenchmarkPureTransformPrograms(b *testing.B) {
	for _, tc := range []struct{ name, source string }{
		{"arrow", `function f($xs){return array_map(fn($x)=>$x+1,$xs)[0];}probe(f([1]));`},
		{"fold", `function add($c,$x){return $c+$x;}function f($xs){return array_reduce($xs,'add',0);}probe(f([1,2]));`},
		{"slice", `function f($xs){return array_slice($xs,1,1)[0];}probe(f([1,2]));`},
		{"keys", `function f($xs){return array_keys($xs,1,true)[0];}probe(f(['a'=>1]));`},
		{"destructure", `function f($xs){$s=0;foreach($xs as [$a,$b]){$s+=$a+$b;}return $s;}probe(f([[1,2],[3,4]]));`},
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
