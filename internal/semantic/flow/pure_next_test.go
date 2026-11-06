package flow

import "testing"

func TestPureNextResults(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   constantScalar
	}{
		{`function f($r){return isset($r['a']);}probe(f(['a'=>[]]));`, constantScalar{k: 'b', b: true}},
		{`function f($r){return ($r['a']??['x'=>'fallback'])['x'];}probe(f(['a'=>['x'=>'kept']]));`, constantScalar{k: 's', s: "kept"}},
		{`function f($r){return ($r['a']??['x'=>'fallback'])['x'];}probe(f([]));`, constantScalar{k: 's', s: "fallback"}},
		{`function f($r){return count($r);}probe(f([1,2]));`, constantScalar{k: 'i', i: 2}},
		{`function f($r){return count(value:$r,mode:0);}probe(f([]));`, constantScalar{k: 'i'}},
		{`function f($r){return count(...[$r]);}probe(f(['x'=>1]));`, constantScalar{k: 'i', i: 1}},
		{`function f($x){return in_array($x,['a','b'],true);}probe(f('z'));`, constantScalar{k: 'b'}},
		{`function f($x){return in_array(needle:$x,haystack:['a','b'],strict:true);}probe(f('b'));`, constantScalar{k: 'b', b: true}},
		{`function f($x){return in_array($x,[1],true);}probe(f('1'));`, constantScalar{k: 'b'}},
		{`function f($xs){$s=0;foreach($xs as $x){$s+=$x;}return $s;}probe(f([1,2]));`, constantScalar{k: 'i', i: 3}},
		{`function f($xs){$s=0;foreach($xs as $x){$s+=1;}return $s;}probe(f([1,2]));`, constantScalar{k: 'i', i: 2}},
		{`function f($xs){$s='';foreach($xs as $x){$s.=$x;}return $s;}probe(f(['a','b']));`, constantScalar{k: 's', s: "ab"}},
		{`function f($xs){$s=0;foreach($xs as $k=>$x){$s=$s+$k;}return $s;}probe(f([2=>10,3=>20]));`, constantScalar{k: 'i', i: 5}},
		{`function f($xs){$s='';foreach($xs as $k=>$x){$s.=$k;}return $s;}probe(f(['01'=>1,'x'=>2]));`, constantScalar{k: 's', s: "01x"}},
		{`final class K{function leaf($x){return $x+1;}function wrap($x){return $this->leaf($x);}}$k=new K;$cb=$k->wrap(...);probe($cb(2));`, constantScalar{k: 'i', i: 3}},
		{`class K{private function leaf($x){return $x+1;}final function wrap($x){return $this->leaf(x:$x);}}$k=new K;$cb=$k->wrap(...);probe($cb(2));`, constantScalar{k: 'i', i: 3}},
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

func TestPureNextRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($r){return count($r,1);}probe(f([]));`,
		`function f($r){return count($r);}probe(f(new stdClass));`,
		`function f($r){return count(unknown:$r);}probe(f([]));`,
		`function f($r){return count($r);}probe(f(opaque()));`,
		`function f($r){return in_array(1,$r);}probe(f([1]));`,
		`function f($r){return in_array(1,$r,false);}probe(f([1]));`,
		`function f($r){return in_array(1,$r,1);}probe(f([1]));`,
		`function f($r){return in_array(1,$r,true);}probe(f([opaque()]));`,
		`function f($r){return isset($r['a']);}probe(f(['a'=>new stdClass]));`,
		`function f($r){return ($r['a']??[])['x'];}probe(f(['a'=>tmpfile()]));`,
		`function f($xs){$s=0;foreach($xs as $s=>$x){$s+=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x=>$x){$s+=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as $x){$s-=$x;}return $s;}probe(f([1]));`,
		`function f($xs){$s=0;foreach($xs as &$x){$s+=$x;}return $s;}probe(f([1]));`,
		`class K{function leaf($x){return $x+1;}final function wrap($x){return $this->leaf($x);}}$k=new K;$cb=$k->wrap(...);probe($cb(2));`,
		`final class K{function leaf($x){return $this->value;}function wrap($x){return $this->leaf($x);}}$k=new K;$cb=$k->wrap(...);probe($cb(2));`,
		`final class K{function leaf($x){return $x;}static function wrap($x){return $this->leaf($x);}}$cb=K::wrap(...);probe($cb(2));`,
		`final class K{function leaf($x){return $x;}function wrap($x){return $this?->leaf($x);}}$k=new K;$cb=$k->wrap(...);probe($cb(2));`,
	} {
		e, file := summarized(t, source)
		if got, known := knownScalar(e.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Errorf("%s: invented %+v", source, got)
		}
	}
}
