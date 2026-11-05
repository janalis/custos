package flow

import "testing"

func BenchmarkPureContracts(b *testing.B) {
	for name, source := range map[string]string{
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
