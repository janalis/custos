package flow

import "testing"

func BenchmarkCallbackRecordContracts(b *testing.B) {
	for name, source := range map[string]string{
		"surplus":         `$a=array_map(fn()=>false,['a','b']);probe($a);`,
		"reduce":          `$a=array_reduce(['safe',$_GET['x']],fn($carry,$entry)=>$entry,'');probe($a);`,
		"named-predicate": `function reject($v){return false;}$a=array_filter([$_GET['x']],'reject');probe($a);`,
		"coercion":        `$a=array_filter(['00','01'],fn(int $v)=>$v);probe($a);`,
		"record":          `function factory(){$h=tmpfile();return ['h'=>$h,'alias'=>$h];}$r=factory();fclose($r['h']);probe($r['alias']);`,
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
