package intltimezoneoffsetmillisecondsusedasseconds

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkAuditProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IntlTimezoneOffsetMillisecondsUsedAsSeconds"}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php $tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$d=new DateTime();$d->modify($raw . " seconds");`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(syntax.Parse("audit.php", src, syntax.Options{}))
	}
}
