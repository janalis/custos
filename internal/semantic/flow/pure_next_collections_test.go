package flow

import "testing"

func TestPureNextCollectionsResults(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($xs){$sum=0;foreach($xs as $x){$sum+=$x;}return $sum+1;}probe(f([1,2]));`, constantScalar{k: 'i', i: 4}},
		{`function f($xs){$sum=0;foreach($xs as $x){$sum+=$x;}return $sum+1;}probe(f([]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs){$sum=0;foreach($xs as $x){$sum+=$x;}return 1;}probe(f([1]));`, constantScalar{k: 'i', i: 1}},
		{`function f($xs,$extra){$sum=0;foreach($xs as $x){$sum+=$x;}return $sum+$extra;}probe(f([1,2],4));`, constantScalar{k: 'i', i: 7}},
		{`function f($xs){$out=[];foreach($xs as $x){$out[]=$x;}return count($out);}probe(f([1,2]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){$out=[];foreach($xs as $x){if($x!==null){$out[]=$x;}}return $out;}probe(f([null,'yes'])[0]);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $k=>$x){if($k===2){$out[]=$x;}}return $out;}probe(f([1=>'no',2=>'yes'])[0]);`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){$out=[];foreach($xs as $x){if($x!==null){$out[]=$x;}}return count($out);}probe(f([null,null]));`, constantScalar{k: 'i'}},
		{`function f($xs){return array_values($xs)[0];}probe(f(['x'=>'yes']));`, constantScalar{k: 's', s: "yes"}},
		{`function f($xs){return array_keys($xs)[0];}probe(f(['x'=>'yes']));`, constantScalar{k: 's', s: "x"}},
		{`function f($xs){return array_keys(...[$xs])[0];}probe(f([2=>'yes']));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){return array_key_first($xs);}probe(f(['01'=>'yes']));`, constantScalar{k: 's', s: "01"}},
		{`function f($xs){return array_key_last(array:$xs);}probe(f(['x'=>'yes',4=>'last']));`, constantScalar{k: 'i', i: 4}},
		{`function f($xs){return array_key_first($xs);}probe(f([]));`, constantScalar{k: 'n'}},
		{`function f($xs){return array_key_last($xs);}probe(f([]));`, constantScalar{k: 'n'}},
		{`function f($v){return is_string($v);}probe(f('yes'));`, constantScalar{k: 'b', b: true}},
		{`function f($v){return is_string($v);}probe(f(1));`, constantScalar{k: 'b'}},
		{`function f($v){return is_int(...[$v]);}probe(f(1));`, constantScalar{k: 'b', b: true}},
		{`function f($v){return is_int($v);}probe(f(false));`, constantScalar{k: 'b'}},
		{`function f($v){return is_array($v);}probe(f([]));`, constantScalar{k: 'b', b: true}},
		{`function f($v){return is_array($v);}probe(f(null));`, constantScalar{k: 'b'}},
		{`function f($v){return is_string($v);}probe(f([]));`, constantScalar{k: 'b'}},
		{`function f($v){return substr($v,1,2);}probe(f('abcd'));`, constantScalar{k: 's', s: "bc"}},
		{`function f($v){return substr($v,-2);}probe(f('abcd'));`, constantScalar{k: 's', s: "cd"}},
		{`function f($v){return substr($v,1,-1);}probe(f('abcd'));`, constantScalar{k: 's', s: "bc"}},
		{`function f($v){return substr(offset:1,string:$v,length:2);}probe(f('abcd'));`, constantScalar{k: 's', s: "bc"}},
		{`function f($v){return substr(...[$v,1,2]);}probe(f('abcd'));`, constantScalar{k: 's', s: "bc"}},
		{`function f($v){return substr($v,99);}probe(f('abc'));`, constantScalar{k: 's'}},
		{`function f($v){return substr($v,2,-2);}probe(f('abc'));`, constantScalar{k: 's'}},
		{`function f($v){return substr($v,0,0);}probe(f('abc'));`, constantScalar{k: 's'}},
		{`function f($v){return substr($v,0);}probe(f(''));`, constantScalar{k: 's'}},
		{`function f($v){return substr($v,-9223372036854775807,9223372036854775807);}probe(f('abc'));`, constantScalar{k: 's', s: "abc"}},
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

func TestPureNextCollectionProvenance(t *testing.T) {
	for _, tc := range []struct {
		source  string
		tainted bool
	}{
		{`function f($rows){$out=[];foreach($rows as $row){if($row['keep']){$out[]=$row['label'];}}return $out;}probe(f([['keep'=>true,'label'=>'fixed','unused'=>$_GET['x']]])[0]);`, false},
		{`function f($rows){$out=[];foreach($rows as $row){if($row['keep']){$out[]=$row['label'];}}return $out;}probe(f([['keep'=>true,'label'=>$_GET['x']]])[0]);`, true},
		{`function f($rows){$out=[];foreach($rows as $row){if($row['keep']){$out[]=$row['label'];}}return count($out);}probe(f([['keep'=>false,'label'=>$_GET['x']]]));`, false},
		{`function f($xs){return array_values($xs);}probe(f(['a'=>'fixed','b'=>$_GET['x']])[0]);`, false},
		{`function f($xs){return array_values($xs);}probe(f(['a'=>'fixed','b'=>$_GET['x']])[1]);`, true},
		{`function f($xs){return array_keys($xs);}probe(f(['a'=>$_GET['x']])[0]);`, false},
	} {
		e, file := summarized(t, tc.source)
		value := e.Value(argument(callsNamed(file, "probe")[0]))
		if got := len(value.Sources) != 0; got != tc.tainted {
			t.Errorf("%s: tainted=%v want=%v", tc.source, got, tc.tainted)
		}
	}
}

func BenchmarkPureNextCollections(b *testing.B) {
	for _, tc := range []struct {
		name, source string
		want         constantScalar
	}{
		{"tail", `function f($xs){$s=0;foreach($xs as $x){$s+=$x;}return $s+1;}probe(f([1,2]));`, constantScalar{k: 'i', i: 4}},
		{"filter", `function f($xs){$out=[];foreach($xs as $x){if($x!==null){$out[]=$x;}}return $out;}probe(f([null,'yes'])[0]);`, constantScalar{k: 's', s: "yes"}},
		{"keys", `function f($xs){return array_keys($xs)[0];}probe(f(['key'=>1]));`, constantScalar{k: 's', s: "key"}},
		{"type", `function f($x){return is_string($x);}probe(f('value'));`, constantScalar{k: 'b', b: true}},
		{"slice", `function f($x){return substr($x,1,2);}probe(f('abcd'));`, constantScalar{k: 's', s: "bc"}},
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
