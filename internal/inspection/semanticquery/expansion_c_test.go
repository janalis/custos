package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionCConstants(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`probe(IntlCalendar::FIELD_HOUR);`, true},
		{`use IntlCalendar as Cal;probe(Cal::FIELD_HOUR);`, true},
		{`probe(IntlCalendar::FIELD_YEAR);`, false},
		{`probe($unknown);`, false},
		{`probe($unknown::FIELD_HOUR);`, false},
		{`class IntlCalendar{const FIELD_HOUR=10;}probe(IntlCalendar::FIELD_HOUR);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ExpansionCConstant(ctx, CallArgument(c.Args, 0, ""), "IntlCalendar", "FIELD_HOUR"); got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestExpansionCStatic(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`probe(Normalizer::normalize("text"));`, true},
		{`probe(Normalizer::isNormalized("text"));`, false},
		{`probe(Other::normalize("text"));`, false},
		{`probe(Normalizer::$name("text"));`, false},
		{`probe("text");`, false},
		{`class Normalizer {static function normalize($s){return $s;}}probe(Normalizer::normalize("text"));`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ExpansionCStatic(ctx, CallArgument(c.Args, 0, ""), "Normalizer", "normalize"); got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestExpansionCOutput(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`pcntl_waitpid($p,$status);probe($status);`, true},
		{`$r=pcntl_waitpid($p,$status);probe($status);`, true},
		{`pcntl_waitpid($p,$other);probe($status);`, false},
		{`if($flag){pcntl_waitpid($p,$status);}probe($status);`, false},
		{`pcntl_waitpid($p,$status);$status=10;probe($status);`, false},
		{`pcntl_waitpid($p,$status);mutate($status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);pcntl_wifexited($status);probe($status);`, true},
		{`pcntl_waitpid($p,$status);$object->call($status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);Other::call($status);probe($status);`, false},
		{`function check(){pcntl_waitpid($p,$status);probe($status);}`, true},
		{`function check(){pcntl_waitpid($p,$status);$status=2;probe($status);}`, false},
		{"pcntl_waitpid($p,$status);" + strings.Repeat("strlen(\"x\");", 300) + "probe($status);", false},
		{"pcntl_waitpid($p,$status);" + strings.Repeat("$noise=1;", 300) + "probe($status);", false},
		{`pcntl_waitpid($p,$status);function unrelated(){$status=2;}probe($status);`, true},
		{`pcntl_waitpid($p,$status);probe(10);`, false},
		{`other($p,$status);probe($status);`, false},
		{`pcntl_waitpid($p,$status);pcntl_waitpid($p,$status);probe($status);`, true},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			got := ExpansionCOutputCall(ctx, c, CallArgument(c.Args, 0, ""), []string{"pcntl_waitpid"}, 1, "status") != nil
			if got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestExpansionCBooleanContexts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`if(probe()) {}`, true},
		{`while(probe()) {}`, true},
		{`do{}while(probe());`, true},
		{`if(!probe()) {}`, true},
		{`if($r=probe()) {}`, true},
		{`$r=probe();`, false},
		{`if(probe()===0) {}`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ExpansionCTruthy(c); got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestExpansionCStableSources(t *testing.T) {
	probeNative(t, `probe($a,$a,$b);`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		a, b, d := CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, ""), CallArgument(c.Args, 2, "")
		if !ExpansionCSame(ctx, a, b) || ExpansionCSame(ctx, a, d) || ExpansionCSame(ctx, nil, b) {
			t.Fatal("source identity mismatch")
		}
	})
}

func BenchmarkExpansionCOutput(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			ExpansionCOutputCall(ctx, c, CallArgument(c.Args, 0, ""), []string{"pcntl_waitpid"}, 1, "status")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php pcntl_waitpid($p,$status);probe($status);`)
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}
