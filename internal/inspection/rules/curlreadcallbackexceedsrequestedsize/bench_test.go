package curlreadcallbackexceedsrequestedsize

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkReadCallback(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlReadCallbackExceedsRequestedSize"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $h=curl_init();curl_setopt($h,CURLOPT_READFUNCTION,function($h,$stream,$size){--$size;return str_repeat('x',$size+1);});"), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
