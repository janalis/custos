package flow

import "testing"

func BenchmarkExtendedContracts(b *testing.B) {
	for name, source := range map[string]string{
		"named-scalar":      `function zero(int $v){return $v;}$xs=array_filter(['00','01'],'zero');probe($xs);`,
		"returned-scalar":   `function factory(){return fn(int $v)=>$v;}$xs=array_filter(['00','01'],factory());probe($xs);`,
		"null-predicate":    `function drop($v){}$xs=array_filter([$_GET['x']],'drop');probe($xs);`,
		"record-pass":       `function pass($r){return $r;}$r=pass(['h'=>tmpfile()]);fclose($r['h']);probe($r['h']);`,
		"record-capture":    `$r=['h'=>tmpfile()];$cb=fn()=>$r;$a=$cb();fclose($a['h']);probe($r['h']);`,
		"returned-resource": `function make(){return fn()=>['h'=>tmpfile()];}$cb=make();$r=$cb();fclose($r['h']);probe($r['h']);`,
		"constant-reduce":   `$sum=array_reduce([1,-1],fn($c,$v)=>$c+$v,0);probe($sum);`,
	} {
		b.Run(name, func(b *testing.B) {
			e, file := summarized(b, source)
			at := argument(callsNamed(file, "probe")[0])
			b.ReportAllocs()
			for b.Loop() {
				New(file, e.types, e.snapshot, Options{}).Value(at)
			}
		})
	}
}
