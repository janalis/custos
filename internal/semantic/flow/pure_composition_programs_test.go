package flow

import "testing"

func TestPureCompositionPrograms(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($xs){return implode(',',$xs);}probe(f(['a','b']));`, constantScalar{k: 's', s: "a,b"}},
		{`function f($xs){return implode('unused',$xs);}probe(f([]));`, constantScalar{k: 's'}},
		{`function f($xs){return implode(array:$xs,separator:'');}probe(f(['a','b']));`, constantScalar{k: 's', s: "ab"}},
		{`function f($xs){return implode(...[',',$xs]);}probe(f(['a']));`, constantScalar{k: 's', s: "a"}},
		{`function f($x){return explode(',',$x)[1];}probe(f('a,b'));`, constantScalar{k: 's', s: "b"}},
		{`function f($x){return explode(separator:',',string:$x)[1];}probe(f('a,'));`, constantScalar{k: 's'}},
		{`function f($x){return count(explode(...[',',$x]));}probe(f(',a,'));`, constantScalar{k: 'i', i: 3}},
		{`function f($x){return explode(',',$x)[0];}probe(f(''));`, constantScalar{k: 's'}},
		{`function f($x){return explode('é',$x)[1];}probe(f('aéb'));`, constantScalar{k: 's', s: "b"}},
		{`function f($xs){return array_column($xs,'label')[0];}probe(f([['label'=>'yes']]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return array_column(column_key:0,array:$xs)[1];}probe(f(['a'=>['x'],9=>['y']]));`, constantScalar{k: 's', s: "y"}},
		{`function f($xs){return count(array_column(...[$xs,'label']));}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($xs){return array_column($xs,'label')[0]===null;}probe(f([['label'=>null]]));`, constantScalar{k: 'b', b: true}},
		{`function f($a,$b){return array_merge($a,$b)[1];}probe(f([9=>'a'],[7=>'b']));`, constantScalar{k: 's', s: "b"}},
		{`function f($a,$b){return array_merge($a,$b)['x'];}probe(f(['x'=>'old'],['x'=>'new']));`, constantScalar{k: 's', s: "new"}},
		{`function f($a,$b){return array_keys(array_merge($a,$b))[0];}probe(f(['x'=>'old','y'=>1],['x'=>'new']));`, constantScalar{k: 's', s: "x"}},
		{`function f($a,$b){return array_merge($a,$b)['01'];}probe(f(['01'=>'yes'],[1=>'no']));`, constantScalar{k: 's', s: "yes"}},
		{`function f($a,$b){return count(array_merge($a,$b));}probe(f([],[]));`, constantScalar{k: 'i'}},
		{`function f($xs){return array_map(function($x){return $x+1;},$xs)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return count(array_filter($xs,function($x){return $x!==null;}));}probe(f([null,1]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){return array_map(static function($x,$n=2){$v=$x+$n;if($v>2){return $v;}return 0;},$xs)[0];}probe(f([1]));`, constantScalar{k: 'i', i: 3}},
		{`declare(strict_types=1);function f($xs){return array_map(function(int $x):int{return $x+1;},$xs)[0];}probe(f(['1']));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return count(array_filter($xs,function($v,$k){return $k==='x'&&$v>1;},1));}probe(f(['x'=>2,'y'=>2]));`, constantScalar{k: 'i', i: 1}},
		{`function f($a,$b){return array_map(function($x,$y){return $x+$y;},$a,$b)[0];}probe(f([1],[2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){return count(array_filter($xs,function($v){}));}probe(f([1]));`, constantScalar{k: 'i'}},
		{`function f($xs){foreach($xs as ['keep'=>$keep,'v'=>$v]){if($keep){return $v;}}return 'none';}probe(f([['keep'=>true,'v'=>'yes'],[]]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[false,'no'],[true,'yes']]));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){foreach($xs as $key=>[$keep,$v]){if($keep){return $key;}}return 'none';}probe(f(['01'=>[true,'yes']]));`, constantScalar{k: 's', s: "01"}},
		{`function f($xs){foreach($xs as list($keep,$v)){if($keep){return $v;}}return 'none';}probe(f([[false,'no']]));`, constantScalar{k: 's', s: "none"}},
		{`function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([]));`, constantScalar{k: 's', s: "none"}},
		{`function f($xs,$v){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[true,'yes']],'old'));`, constantScalar{k: 's', s: "yes"}},
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

func TestPureCompositionProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($xs){return array_column($xs,'label');}probe(f([['label'=>'fixed','unused'=>$_GET['x']]])[0]);`, false},
		{`function f($xs){return array_column($xs,'label');}probe(f([['label'=>$_GET['x']]])[0]);`, true},
		{`function f($a,$b){return array_merge($a,$b);}probe(f(['x'=>$_GET['x']],['x'=>'fixed'])['x']);`, false},
		{`function f($a,$b){return array_merge($a,$b);}probe(f(['x'=>'fixed'],['x'=>$_GET['x']])['x']);`, true},
		{`function f($xs){return array_map(function($row){return $row['v'];},$xs);}probe(f([['v'=>'fixed','unused'=>$_GET['x']]])[0]);`, false},
		{`function f($xs){return array_map(function($row){return $row['v'];},$xs);}probe(f([['v'=>$_GET['x']]])[0]);`, true},
		{`function f($xs){foreach($xs as ['keep'=>$keep,'v'=>$v]){if($keep){return $v;}}return 'fixed';}probe(f([['keep'=>false,'v'=>$_GET['x']]]));`, false},
		{`function f($xs){foreach($xs as ['keep'=>$keep,'v'=>$v]){if($keep){return $v;}}return 'fixed';}probe(f([['keep'=>true,'v'=>$_GET['x']]]));`, true},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if (len(value.Sources) != 0) != tc.tainted {
			t.Errorf("provenance %s", tc.source)
		}
	}
}

func BenchmarkPureCompositionPrograms(b *testing.B) {
	for _, tc := range []struct{ name, source string }{
		{"join", `function f($xs){return implode(',',$xs);}probe(f(['a','b']));`},
		{"split", `function f($x){return explode(',',$x)[1];}probe(f('a,b'));`},
		{"column", `function f($xs){return array_column($xs,'label')[0];}probe(f([['label'=>'yes']]));`},
		{"merge", `function f($a,$b){return array_merge($a,$b)[1];}probe(f([9=>'a'],[7=>'b']));`},
		{"closure", `function f($xs){return array_map(function($x){return $x+1;},$xs)[0];}probe(f([1]));`},
		{"find", `function f($xs){foreach($xs as [$keep,$v]){if($keep){return $v;}}return 'none';}probe(f([[true,'yes']]));`},
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
