package contentlengthusescharactercount

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkContentLengthBodyProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentLengthUsesCharacterCount"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("body.php", []byte("<?php function response($body){header('Content-Length: '.mb_strlen($body,'UTF-8'));echo $body;}"), syntax.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
