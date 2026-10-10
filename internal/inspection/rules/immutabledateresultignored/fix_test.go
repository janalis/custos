package immutabledateresultignored_test

import (
	"strings"
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/immutabledateresultignored"
	"custos/internal/php/syntax"
)

func TestFixSafety(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		fix          bool
	}{
		{"fresh top level", "$day=new DateTimeImmutable('2025-03-10'); $day->modify('+1 day'); echo $day->format('c');", true},
		{"fresh function", "function f($other){$day=new DateTimeImmutable(); $day->modify('-2 days'); echo $day->format('Y');}", true},
		{"fresh method", "class C{function f(){ $day=new DateTimeImmutable; $day->modify('0 hours'); echo $day->format('c');}}", true},
		{"namespace", "namespace App; $day=new \\DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');", true},
		{"comments preserved", "$day=new DateTimeImmutable; $day /* receiver */ ->modify(/* adjustment */ '+1 day'); echo $day->format('c');", true},
		{"typed prior non-expression", "/** @var DateTimeImmutable $day */ echo 'x'; $day->modify('+1 day'); echo $day->format('c');", false},
		{"typed prior non-assignment", "/** @var DateTimeImmutable $day */ print 'x'; $day->modify('+1 day'); echo $day->format('c');", false},
		{"typed other initializer", "/** @var DateTimeImmutable $day */ $other=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');", false},
		{"no adjacent read", "$day=new DateTimeImmutable; $day->modify('+1 day');", false},
		{"unknown flow", "$day=new DateTimeImmutable; $alias=$day; $day->modify('+1 day'); echo $day->format('c');", false},
		{"parameter alias", "function f(&$day){$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');}", false},
		{"closure capture", "$f=function()use(&$day){$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');};", false},
		{"loop inherited alias", "while($ok){$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');}", false},
		{"case inherited alias", "switch($x){case 1:$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');}", false},
		{"unbraced", "$day=new DateTimeImmutable; if($ok)$day->modify('+1 day'); echo $day->format('c');", false},
		{"temporary receiver", "(new DateTimeImmutable)->modify('+1 day');", false},
		{"non-local", "$box->day=new DateTimeImmutable; $box->day->modify('+1 day'); echo $box->day->format('c');", false},
		{"reference binding", "$other=new DateTimeImmutable; $day=&$other; $day->modify('+1 day'); echo $day->format('c');", false},
		{"this", "class C extends DateTimeImmutable{function f(){$this->modify('+1 day');}}", false},
		{"superglobal", "$_SESSION=new DateTimeImmutable; $_SESSION->modify('+1 day'); echo $_SESSION->format('c');", false},
		{"globals", "$GLOBALS=new DateTimeImmutable; $GLOBALS->modify('+1 day'); echo $GLOBALS->format('c');", false},
		{"other mutator", "$day=new DateTimeImmutable; $day->setTimestamp(10); echo $day->format('c');", false},
		{"invalid modifier", "$day=new DateTimeImmutable; $day->modify('invalid'); echo $day->format('c');", false},
		{"invalid count", "$day=new DateTimeImmutable; $day->modify('many days'); echo $day->format('c');", false},
		{"huge count", "$day=new DateTimeImmutable; $day->modify('10000 days'); echo $day->format('c');", false},
		{"unsupported unit", "$day=new DateTimeImmutable; $day->modify('+1 parsecs'); echo $day->format('c');", false},
		{"dynamic modifier", "$day=new DateTimeImmutable; $day->modify($change); echo $day->format('c');", false},
		{"named modifier", "$day=new DateTimeImmutable; $day->modify(modifier:'+1 day'); echo $day->format('c');", false},
		{"unpacked modifier", "$day=new DateTimeImmutable; $day->modify(...['+1 day']); echo $day->format('c');", false},
		{"multiple modifier args", "$day=new DateTimeImmutable; $day->modify('+1 day',call()); echo $day->format('c');", false},
		{"no modifier arg", "$day=new DateTimeImmutable; $day->modify(); echo $day->format('c');", false},
		{"first class callable", "$day=new DateTimeImmutable; $day->modify(...); echo $day->format('c');", false},
		{"constructor evaluation", "$day=new DateTimeImmutable(factory()); $day->modify('+1 day'); echo $day->format('c');", false},
		{"discarded creation", "new DateTimeImmutable; $day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');", false},
		{"non assignment initializer", "function f(DateTimeImmutable $day){print 'x'; $day->modify('+1 day'); echo $day->format('c');}", false},
		{"different initializer", "function f(DateTimeImmutable $day){$other=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');}", false},
		{"assigned parameter", "function f(DateTimeImmutable $other){$day=$other; $day->modify('+1 day'); echo $day->format('c');}", false},
		{"not echo", "$day=new DateTimeImmutable; $day->modify('+1 day'); return $day->format('c');", false},
		{"multiple echo values", "$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c'),call();", false},
		{"direct echo", "$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day;", false},
		{"different receiver", "$day=new DateTimeImmutable; $day->modify('+1 day'); echo $other->format('c');", false},
		{"different read", "$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->getTimestamp();", false},
		{"dynamic read name", "$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->{$name}('c');", false},
		{"unknown read args", "$day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format(call());", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte("<?php " + tc.source)
			result := fixing.FixSource(e, "test.php", src, fixing.Options{SinglePass: true})
			if (result.Applied > 0) != tc.fix {
				t.Fatalf("applied %d: %s", result.Applied, result.Source)
			}
			if tc.fix {
				if !strings.Contains(string(result.Source), "$day = $day") {
					t.Fatalf("missing assignment: %s", result.Source)
				}
				if len(e.Analyze(syntax.Parse("test.php", result.Source, syntax.Options{}))) != 0 {
					t.Fatal("fix left a finding")
				}
			}
		})
	}
}

func BenchmarkFreshLocalFix(b *testing.B) {
	e, _ := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
	f := syntax.Parse("bench.php", []byte("<?php $day=new DateTimeImmutable; $day->modify('+1 day'); echo $day->format('c');"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(f)
	}
}
