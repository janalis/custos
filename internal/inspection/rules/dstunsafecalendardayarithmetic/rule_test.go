package dstunsafecalendardayarithmetic_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/dstunsafecalendardayarithmetic"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$day=new DateTimeImmutable('2025-03-30 00:00:00',new DateTimeZone('Europe/Paris')); $next=$day->getTimestamp()+86400;\n", 1},
		{"negative0", "<?php\n$d=new DateTimeImmutable('2025-03-30 00:00:00',new DateTimeZone('UTC'));$d->getTimestamp()+86400;\n", 0},
		{"negative1", "<?php\n$d=new DateTimeImmutable('2025-03-30 12:00:00',new DateTimeZone('Europe/Paris'));$d->getTimestamp()+86400;\n", 0},
		{"negative2", "<?php\n$ts+86400;\n", 0},
		{"negative3", "<?php\n$d=new DateTimeImmutable();$d->getTimestamp()+3600;\n", 0},
		{"negative4", "<?php\n$d=new DateTimeImmutable('2025-03-30 00:00:00');$d->getTimestamp()+86400;\n", 0},
		{"negative5", "<?php\n$d->getTimestamp()-86400;\n", 0},
		{"negative6", "<?php\n86400+$x;\n", 0},
		{"negative7", "<?php\n$x=new DateTimeImmutable('2025-01-01 00:00:00',new DateTimeZone('Unknown'));$x->getTimestamp()+86400;\n", 0},
		{"negative8", "<?php\n$x=new DateTimeImmutable('2025-99-99 00:00:00',new DateTimeZone('Europe/Paris'));$x->getTimestamp()+86400;\n", 0},
		{"negative9", "<?php\n$x=new DateTimeImmutable('2025-01-01 00:00:00',new DateTimeZone($zone));$x->getTimestamp()+86400;\n", 0},
		{"negative10", "<?php\n$x=new DateTimeImmutable($date,new DateTimeZone('Europe/Paris'));$x->getTimestamp()+86400;\n", 0},
		{"negative11", "<?php\n$x=new DateTimeImmutable();$x->format('c')+86400;\n", 0},
		{"negative12", "<?php\n$x=DateTimeImmutable::createFromFormat('Y-m-d',$date);$x->getTimestamp()+86400;\n", 0},
		{"mutable date changed", "<?php $d=new DateTime('2025-03-30 00:00:00',new DateTimeZone('Europe/Paris'));$d->setTimezone(new DateTimeZone('UTC'));$next=$d->getTimestamp()+86400;", 0},
		{"fresh mutable midnight", "<?php $next=(new DateTime('2025-03-30 00:00:00',new DateTimeZone('Europe/Paris')))->getTimestamp()+86400;", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rule.New().(analysis.SemanticRule).Semantic()
			e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte(tc.source), syntax.Options{})
			findings := e.Analyze(f)
			if len(findings) != tc.count {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.count, findings)
			}
		})
	}
}
