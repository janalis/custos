package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func BenchmarkSummaryDeclarationInventory(b *testing.B) {
	var source strings.Builder
	source.WriteString(`function pass($x){return $x;} `)
	for i := range 64 {
		fmt.Fprintf(&source, `function caller%d($x){return pass($x);}`, i)
	}
	e, file := environment(b, source.String(), nil)
	b.ReportAllocs()
	for b.Loop() {
		Extract(file, e.types.Index, e.types.PHP, nil, Options{})
	}
}

func BenchmarkMappedPrecision(b *testing.B) {
	for name, source := range map[string]string{
		"multiple": `$xs=array_map(fn($x,$y)=>$x,[$_GET['x'],'safe'],['a','b']);probe($xs[1]);`,
		"nested":   `$xs=array_map(fn($x)=>['clean'=>'safe','dirty'=>$x],[$_GET['x']]);probe($xs[0]['clean']);`,
	} {
		b.Run(name, func(b *testing.B) {
			e, file := environment(b, source, nil)
			at := argument(callsNamed(file, "probe")[0])
			b.ReportAllocs()
			for b.Loop() {
				New(file, e.types, nil, Options{}).Value(at)
			}
		})
	}
}

func BenchmarkMultipleReferenceOutputs(b *testing.B) {
	e, file := summarized(b, `function fill(&$a,&$b){$a='safe';$b='safe';}function caller(){$a='x';$b='y';fill($a,$b);}`)
	at := callsNamed(file, "fill")[0]
	b.ReportAllocs()
	for b.Loop() {
		e.ReferenceResult(at, at.Args.Args[0].(*syntax.Arg))
	}
}

func BenchmarkBorrowedResourceEffects(b *testing.B) {
	e, file := summarized(b, `function readIt($h){fread($h,1);}$h=tmpfile();readIt($h);probe($h);`)
	at := callsNamed(file, "probe")[0]
	b.ReportAllocs()
	for b.Loop() {
		New(file, e.types, e.snapshot, Options{}).StateBefore(at, argument(at), "fread")
	}
}
