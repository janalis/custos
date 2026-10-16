package filelockfailureunchecked

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

func TestOptionalAssertionLockAttempt(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, tc := range []struct {
			body string
			want int
		}{
			{`assert(flock($h,LOCK_EX));fwrite($h,'x');`, 1},
			{`assert(flock(operation:LOCK_EX|LOCK_NB,stream:$h));ftruncate($h,0);`, 1},
			{`assert(flock($h,LOCK_EX));$h=fopen('other','c');fwrite($h,'x');`, 0},
			{`assert(flock($h,LOCK_EX));unknown($h);fwrite($h,'x');`, 0},
			{`assert(flock($h,LOCK_EX));fclose($h);fwrite($h,'x');`, 0},
			{`assert(false&&flock($h,LOCK_EX));fwrite($h,'x');`, 0},
			{`assert(flock($other,LOCK_EX));fwrite($h,'x');`, 0},
			{`flock($other,LOCK_EX);fwrite($h,'x');`, 0},
			{`assert(flock(operation:LOCK_EX));fwrite($h,'x');`, 0},
			{`assert(flock($h,$mode));fwrite($h,'x');`, 0},
			{`assert(flock($h,LOCK_SH));fwrite($h,'x');`, 0},
			{`if($c){assert(flock($h,LOCK_EX));}fwrite($h,'x');`, 0},
			{`assert(flock($h,LOCK_EX));assert(true);fwrite($h,'x');`, 1},
			{strings.Repeat(`strlen('x');`, 257) + `assert(flock($h,LOCK_EX));fwrite($h,'x');`, 1},
			{`assert(flock($h,LOCK_EX));` + strings.Repeat(`strlen('x');`, 257) + `fwrite($h,'x');`, 0},
		} {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Flow: flow.Options{Assertions: enabled}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte(`<?php function f($h,$other){`+tc.body+`}`), syntax.Options{}))
			want := tc.want
			if enabled && want == 1 {
				want = 0
			}
			if len(got) != want {
				t.Fatalf("enabled=%v %s got=%v", enabled, tc.body, got)
			}
		}
	}
}

func BenchmarkOptionalAssertionLock(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(`<?php function f($h){`+strings.Repeat(`strlen('x');`, 257)+`assert(flock($h,LOCK_EX));fwrite($h,'x');}`), syntax.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(f)
	}
}

func TestAssertionLockUserDeclaration(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte(`<?php namespace Project; function flock($h,$operation){} function f($h){assert(flock($h,LOCK_EX));fwrite($h,'x');}`), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("application flock considered builtin: %v", got)
	}
}
