package flow

import "testing"

func BenchmarkPureContracts(b *testing.B) {
	for name, source := range map[string]string{
		"presence":  `function f($r){return isset($r['x']);}probe(f(['x'=>[]]));`,
		"count":     `function f($r){return count($r);}probe(f([1,2]));`,
		"member":    `function f($x){return in_array($x,['a','b'],true);}probe(f('b'));`,
		"keyed":     `function f($xs){$s=0;foreach($xs as $k=>$v){$s+=$k;}return $s;}probe(f([2=>1,3=>1]));`,
		"instance":  `final class K{function leaf($x){return $x+1;}function wrap($x){return $this->leaf($x);}}$k=new K;$cb=$k->wrap(...);probe($cb(2));`,
		"match":     `function f($v){return match($v){1=>2,default=>3};}probe(f(1));`,
		"quiet":     `function f($r){return $r['x']??'fallback';}probe(f([]));`,
		"static":    `final class K{static function add($v){return $v+1;}}function f($v){return K::add($v);}probe(f(2));`,
		"foreach":   `function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}return $s;}probe(f([1,2,3]));`,
		"reduction": `function add($c,$v){return $c+$v;}probe(array_reduce([1,-1,2,-2],'add',0));`,
		"wrapper":   `function pass($x){return $x;}function wrap($x){$y=$x;return pass($y);}$r=['h'=>tmpfile()];$a=wrap($r);fclose($a['h']);probe($r['h']);`,
		"capture":   `function factory(int $v){return fn($x)=>$v+$x;}$cb=factory(2);probe($cb(3));`,
		"truth":     `function drop($v){if($v)return false;return null;}$xs=array_filter([$_GET['x']],'drop');probe($xs);`,
		"default":   `function choose($x,int $v=0){return $v;}$xs=array_filter([$_GET['x']],'choose');probe($xs);`,
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				e, f := summarized(b, source)
				e.Value(argument(callsNamed(f, "probe")[0]))
			}
		})
	}
}
