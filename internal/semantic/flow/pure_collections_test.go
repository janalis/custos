package flow

import "testing"

func TestPureCollectionsResults(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($v){return in_array($v,[],true);}probe(f($_GET['x']));`, constantScalar{k: 'b'}},
		{`function f($v){return in_array('yes',['yes',$v],true);}probe(f($_GET['x']));`, constantScalar{k: 'b', b: true}},
		{`function f($xs){return count($xs,COUNT_NORMAL);}probe(f([1,2]));`, constantScalar{k: 'i', i: 2}},
		{`namespace N;use const COUNT_NORMAL as MODE;function f($xs){return count($xs,MODE);}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($s){return str_starts_with($s,'ok:');}probe(f('ok:value'));`, constantScalar{k: 'b', b: true}},
		{`function f($s){return str_ends_with(needle:'.php',haystack:$s);}probe(f('a.php'));`, constantScalar{k: 'b', b: true}},
		{`function f($s){return str_contains(...[$s,'needle']);}probe(f('a needle b'));`, constantScalar{k: 'b', b: true}},
		{`function f($s){return str_contains($s,'');}probe(f(''));`, constantScalar{k: 'b', b: true}},
		{`function f($s){return str_starts_with($s,'x');}probe(f('y'));`, constantScalar{k: 'b'}},
		{`function f($xs){return [...$xs,'last'][0];}probe(f(['first']));`, constantScalar{k: 's', s: "first"}},
		{`function f($xs){return [5=>'a',...$xs,'last'][7];}probe(f([9=>'first']));`, constantScalar{k: 's', s: "last"}},
		{`function f($xs){return ['x'=>'a',...$xs]['x'];}probe(f(['x'=>'b']));`, constantScalar{k: 's', s: "b"}},
		{`function f($xs){return [...$xs,'x'=>'last']['x'];}probe(f(['x'=>'first']));`, constantScalar{k: 's', s: "last"}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=count($out);}return $out;}probe(f([1,2])[1]);`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x;}return $out;}function g($xs){return count(f($xs),COUNT_NORMAL);}probe(g([]));`, constantScalar{k: 'i'}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x;}return $out;}probe(f([1,2])[1]);`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){$out=[];foreach($xs as $k=>$x){$out[]=$k;}return $out;}probe(f(['01'=>1,'x'=>2])[1]);`, constantScalar{k: 's', s: "x"}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x['a'];}return $out;}probe(f([['a'=>'yes']])[0]);`, constantScalar{k: 's', s: "yes"}},
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

func TestPureCollectionsRefusals(t *testing.T) {
	for _, source := range []string{
		`namespace N;const COUNT_NORMAL=1;function f($x){return count($x,COUNT_NORMAL);}probe(f([]));`,
		`function f($x){return in_array('yes',[$x],true);}probe(f($_GET['x']));`,
		`function f($x){return in_array($x,[],true);}probe(f(new stdClass));`,
		`function f($x){return in_array('yes',['yes',$x],true);}probe(f(new stdClass));`,
		`function f($s){return str_contains($s,'1');}probe(f(123));`,
		`function f($s){return str_contains($s);}probe(f('a'));`,
		`function f($s){return str_contains(unknown:$s,needle:'a');}probe(f('a'));`,
		`function f($xs){return [...$xs][0];}probe(f(opaque()));`,
		`function f($xs){return [...$xs][0];}probe(f([new stdClass]));`,
		`function f($xs){return [...$xs,-1=>'a'][0];}probe(f([]));`,
		`function f($xs){$out=[];foreach($xs as &$x){$out[]=$x;}return $out;}probe(f([1])[0]);`,
		`function f($xs){$out=[];foreach($xs as $x){$out[1]=$x;}return $out;}probe(f([1])[1]);`,
		`function f($xs){$out=[];foreach($xs as $x){$out[]=$x;}return $out;}probe(f([tmpfile()])[0]);`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}

func BenchmarkPureCollections(b *testing.B) {
	for _, tc := range []struct {
		name, source string
		want         constantScalar
	}{
		{"append", `function f($xs){$out=[];foreach($xs as $x){$out[]=str_contains($x,'a');}return $out;}probe(f(['a','b','ca'])[1]);`, constantScalar{k: 'b'}},
		{"spread", `function f($xs){return [...$xs,'last'][1];}probe(f(['first']));`, constantScalar{k: 's', s: "last"}},
		{"membership", `function f($xs){return in_array('yes',$xs,true);}probe(f(['yes',$_GET['x']]));`, constantScalar{k: 'b', b: true}},
		{"constant", `function f($xs){return count($xs,COUNT_NORMAL);}probe(f([1,2]));`, constantScalar{k: 'i', i: 2}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			e, file := summarized(b, tc.source)
			call := argument(callsNamed(file, "probe")[0])
			if got, known := knownScalar(e.Value(call)); !known || got != tc.want {
				b.Fatalf("result=%+v known=%v want=%+v", got, known, tc.want)
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

func TestPureCollectionProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x['a'];}return $out;}probe(f([['a'=>'fixed','other'=>$_GET['x']]])[0]);`, false},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x['a'];}return $out;}probe(f([['a'=>$_GET['x']]])[0]);`, true},
		{`function f($xs){return [...$xs];}probe(f(['fixed',$_GET['x']])[0]);`, false},
		{`function f($xs){return [...$xs];}probe(f(['fixed',$_GET['x']])[1]);`, true},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(value.Sources) != 0; got != tc.tainted {
			t.Errorf("%s: tainted=%v want=%v", tc.source, got, tc.tainted)
		}
	}
}
