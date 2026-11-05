package flow

import "testing"

func BenchmarkImmutableShapeOperations(b *testing.B) {
	e, file := environment(b, `function row(){return ['a'=>'safe','b'=>['nested'=>'text']];}`, nil)
	out := Extract(file, e.types.Index, e.types.PHP, nil, Options{})
	template := out.Summaries["row"].Return
	b.Run("compact", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			fuel := MaxSources
			compactElements(template.elements, MaxDepth, &fuel)
		}
	})
	b.Run("substitute", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			substitute(template, nil, true)
		}
	})
}

func BenchmarkEmptyShapeAnalysis(b *testing.B) {
	e, file := environment(b, `$a=[];probe($a);`, nil)
	at := argument(callsNamed(file, "probe")[0])
	b.ReportAllocs()
	for b.Loop() {
		New(file, e.types, nil, Options{}).Value(at)
	}
}

func BenchmarkFiniteCallbackAnalysis(b *testing.B) {
	for name, source := range map[string]string{
		"filter":         `$a=array_filter(['safe',$_GET['x']],fn($x)=>$x);probe($a);`,
		"zip":            `$a=array_map(null,['safe'],[$_GET['x']]);probe($a[0][0]);`,
		"returned-shape": `function factory(){return fn($x)=>['safe'=>'text','dirty'=>$x];}$cb=factory();$a=$cb($_GET['x']);probe($a['safe']);`,
		"result-close":   `function shut($h){return fclose($h);}$h=tmpfile();if(shut($h)){probe($h);}`,
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
