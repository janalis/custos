package flow

import "testing"

func TestPureCallbackVariants(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function keep($x){return $x>1;}function f($xs){return count(array_filter($xs,'keep',0));}probe(f([1,2]));`, constantScalar{k: 'i', i: 1}},
		{`function keep($k){return $k==='yes';}function f($xs){return array_filter($xs,'keep',ARRAY_FILTER_USE_KEY)['yes'];}probe(f(['yes'=>7,'no'=>8]));`, constantScalar{k: 'i', i: 7}},
		{`function keep($v,$k){return $v>1&&$k==='yes';}function f($xs){return count(array_filter(mode:ARRAY_FILTER_USE_BOTH,callback:'keep',array:$xs));}probe(f(['yes'=>2,'no'=>3]));`, constantScalar{k: 'i', i: 1}},
		{`function keep($k){return $k===2;}function f($xs){return count(array_filter($xs,'keep',2));}probe(f([2=>7,'02'=>8]));`, constantScalar{k: 'i', i: 1}},
		{`function keep($v,$k,$extra=1){return $v===$extra&&$k===0;}function f($xs){return count(array_filter($xs,'keep',1));}probe(f([1,2]));`, constantScalar{k: 'i', i: 1}},
		{`function keep($x){return $x>1;}function f($xs){return count(array_filter(...[$xs,'keep',0]));}probe(f([1,2]));`, constantScalar{k: 'i', i: 1}},
		{`function add($x,$y){return $x+$y;}function f($a,$b){return array_map('add',$a,$b)[0];}probe(f(['x'=>1],['y'=>2]));`, constantScalar{k: 'i', i: 3}},
		{`function add(int $x,int $y):int{return $x+$y;}function f($a,$b){return array_map('add',$a,$b)[1];}probe(f(['b'=>'1','a'=>'2'],[9=>3,0=>4]));`, constantScalar{k: 'i', i: 6}},
		{`function add($x,$y,$z=3){return $x+$y+$z;}function f($a,$b){return array_map('add',$a,$b)[0];}probe(f([1],[2]));`, constantScalar{k: 'i', i: 6}},
		{`function add($x,$y){return $x+$y;}function f($a,$b){return count(array_map('add',$a,$b));}probe(f([],[]));`, constantScalar{k: 'i'}},
		{`function f($xs){return array_map('strlen',$xs)['x'];}probe(f(['x'=>'hello']));`, constantScalar{k: 'i', i: 5}},
		{`function f($xs){return array_map('\trim',$xs)[0];}probe(f([" \tword\n"]));`, constantScalar{k: 's', s: "word"}},
		{`function f($xs){return array_map(callback:'strlen',array:$xs)[0];}probe(f(['hello']));`, constantScalar{k: 'i', i: 5}},
		{`function f($xs){return array_map('trim',$xs)[0];}probe(f(['yes']));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$s=0;foreach($xs as $x){if($x===null){break;}$s+=$x;}return $s;}probe(f([1,null,2]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){$s=0;foreach($xs as $x){if($x){break 1;}$s+=$x;}return $s;}probe(f([1]));`, constantScalar{k: 'i'}},
		{`function f($xs){$s=0;foreach($xs as $x){if($x===null){break;}$s+=$x;}return $s;}probe(f([1,2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$s=0;foreach($xs as $x){if($x===null){break;}$s+=$x;}return $s;}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($xs){$out=[];foreach($xs as $x){if(!$x['keep']){break;}$out[$x['id']]=$x['v'];}return $out['x'];}probe(f([['keep'=>true,'id'=>'x','v'=>'yes'],['keep'=>false],[]]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $x){if($x===null){break;}$out[]=$x;}return count($out);}probe(f(['yes',null,'no']));`, constantScalar{k: 'i', i: 1}},
		{`function f($x){return match($x){[1]=>'yes',default=>'no'};}probe(f([1]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($x){return match($x){[1]=>'yes',default=>'no'};}probe(f([2]));`, constantScalar{k: 's', s: "no"}},
		{`function f($x){return match($x){1=>'scalar',[1]=>'yes',default=>'no'};}probe(f([1]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($x){return match($x){[1]=>'array',1=>'yes',default=>'no'};}probe(f(1));`, constantScalar{k: 's', s: "yes"}},
		{`function f($x){return match($x){[1]=>'yes',default=>$x['missing']};}probe(f([1]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($x){return in_array($x,[[1],[2]],true);}probe(f([1]));`, constantScalar{k: 'b', b: true}},
		{`function f($x){return in_array($x,[1,[2]],true);}probe(f([1]));`, constantScalar{k: 'b'}},
		{`function f($x){return in_array($x,[],true);}probe(f([1]));`, constantScalar{k: 'b'}},
		{`function f($x){return in_array($x,[[1],[$x['missing']]],true);}probe(f([1]));`, constantScalar{}},
	} {
		e, file := summarized(t, tc.source)
		for _, env := range []*Env{e, NewLocal(file, e.types, e.snapshot, Options{}, nil)} {
			got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
			if tc.want.k == 0 {
				if known {
					t.Errorf("unexecuted argument: %s", tc.source)
				}
				continue
			}
			if !known || got != tc.want {
				t.Errorf("%s: got=%+v known=%v want=%+v", tc.source, got, known, tc.want)
			}
		}
	}
}

func TestPureCallbackVariantProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($x,$v){return match($x){1=>'fixed',default=>$v};}probe(f(1,$_GET['v']));`, false},
		{`function f($x,$v){return match($x){[1]=>'fixed',default=>$v};}probe(f([1],$_GET['v']));`, false},
		{`function keep($k){return $k==='x';}function f($xs){return array_filter($xs,'keep',2);}probe(f(['x'=>$_GET['v']])['x']);`, true},
		{`function keep($v,$k){return $k==='x';}function f($xs){return array_filter($xs,'keep',1);}probe(f(['x'=>$_GET['v']])['x']);`, true},
		{`function pick($a,$b){return $b['v'];}function f($a,$b){return array_map('pick',$a,$b);}probe(f([$_GET['unused']],[['v'=>'fixed']])[0]);`, false},
		{`function pick($a,$b){return $b['v'];}function f($a,$b){return array_map('pick',$a,$b);}probe(f([1],[['v'=>$_GET['v']]])[0]);`, true},
		{`function f($xs){return array_map('trim',$xs);}probe(f([$_GET['v']])[0]);`, true},
		{`function f($xs){$out=[];foreach($xs as $x){if($x===null){break;}$out[]=$x;}return $out;}probe(f([$_GET['v'],null])[0]);`, true},
		{`function f($x,$v){return match($x){[1]=>$v,default=>'fixed'};}probe(f([1],$_GET['v']));`, true},
	} {
		e, file := summarized(t, tc.source)
		v := e.Value(argument(callsNamed(file, "probe")[0]))
		if (len(v.Sources) != 0) != tc.tainted {
			t.Errorf("provenance %s", tc.source)
		}
	}
}

func BenchmarkPureCallbackVariants(b *testing.B) {
	for _, tc := range []struct{ name, source string }{
		{"filter", `function keep($v,$k){return $v>1&&$k==='x';}function f($xs){return count(array_filter($xs,'keep',1));}probe(f(['x'=>2]));`},
		{"map-two", `function add($a,$b){return $a+$b;}function f($a,$b){return array_map('add',$a,$b)[0];}probe(f([1],[2]));`},
		{"builtin", `function f($xs){return array_map('strlen',$xs)[0];}probe(f(['yes']));`},
		{"query", `function f($xs){return in_array($xs,[[1],[2]],true);}probe(f([1]));`},
		{"break", `function f($xs){$s=0;foreach($xs as $x){if($x===null){break;}$s+=$x;}return $s;}probe(f([1,null,2]));`},
	} {
		b.Run(tc.name, func(b *testing.B) {
			e, file := summarized(b, tc.source)
			call := argument(callsNamed(file, "probe")[0])
			if _, known := knownScalar(e.Value(call)); !known {
				b.Fatal("unknown")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				NewLocal(file, e.types, e.snapshot, Options{}, nil).Value(call)
			}
		})
	}
}
