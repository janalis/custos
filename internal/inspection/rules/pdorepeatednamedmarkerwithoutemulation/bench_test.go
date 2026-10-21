package pdorepeatednamedmarkerwithoutemulation

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkRepeatedPrepares(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoRepeatedNamedMarkerWithoutEmulation"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);"+strings.Repeat("$p->prepare('SELECT :a+:a');", 100)), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
