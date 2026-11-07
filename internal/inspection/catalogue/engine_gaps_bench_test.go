package catalogue

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func BenchmarkEngineGapQueries(b *testing.B) {
	for _, tc := range []struct{ rule, source string }{
		{"SuspiciousAssignments", `function f($x,$other){$saved=compact('other');$x=1;$cb=fn()=> $x;return [$x,$cb,$saved];}`},
		{"CurlSslServerSpoofing", `function f($h){curl_setopt(...['value'=>false,'handle'=>$h,'option'=>CURLOPT_SSL_VERIFYPEER]);}`},
		{"UnqualifiedReference", `namespace A { strlen('a'); } namespace { strlen('b'); array_map('strlen',['a']); }`},
		{"CascadeStringReplacement", `function f($s){$search=['a','b'];$s=str_replace($search,'x',$s);$s=str_replace([],[],$s);return str_replace('c','y',$s);}`},
	} {
		b.Run(tc.rule, func(b *testing.B) {
			engine, err := analysis.NewEngine(All(), analysis.Config{PHP: phpversion.Max, Only: []string{tc.rule}})
			if err != nil {
				b.Fatal(err)
			}
			file := syntax.Parse("bench.php", []byte("<?php "+tc.source), syntax.Options{Version: phpversion.Max})
			b.ReportAllocs()
			for b.Loop() {
				engine.Analyze(file)
			}
		})
	}
}
