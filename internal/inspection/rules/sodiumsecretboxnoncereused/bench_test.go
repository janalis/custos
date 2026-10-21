package sodiumsecretboxnoncereused

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkNonceReuse(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SodiumSecretboxNonceReused"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $k=random_bytes(32);$n=random_bytes(24);sodium_crypto_secretbox('first',$n,$k);sodium_increment($n);sodium_crypto_secretbox('second',$n,$k);"), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}

func BenchmarkDense(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SodiumSecretboxNonceReused"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $k=random_bytes(32);$n=random_bytes(24);"+strings.Repeat("sodium_crypto_secretbox('same',$n,$k);", 100)), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
