package semanticquery

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/stubs"
)

func TestIsBuiltinClass(t *testing.T) {
	f := syntax.Parse("mine.php", []byte("<?php class Mine extends \\RuntimeException {}"), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !IsBuiltinClass(ix.Class("RuntimeException", phpversion.PHP84), phpversion.PHP84) {
		t.Fatal("RuntimeException should be builtin")
	}
	if IsBuiltinClass(ix.Class("Mine", phpversion.PHP84), phpversion.PHP84) {
		t.Fatal("Mine should not be builtin")
	}
	if IsBuiltinClass(nil, phpversion.PHP84) {
		t.Fatal("nil is not builtin")
	}
}

func TestBuiltinClassConstant(t *testing.T) {
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
			if got := BuiltinClassConstant(ctx, CallArgument(c.Args, 0, ""), "IntlCalendar", "FIELD_HOUR"); got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestBuiltinStaticCall(t *testing.T) {
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
			if got := BuiltinStaticCall(ctx, CallArgument(c.Args, 0, ""), "Normalizer", "normalize"); got != tc.want {
				t.Fatalf("%s got %v", tc.src, got)
			}
		})
	}
}

func TestReadsGlobalConstant(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`probe(SQLITE3_NULL);`, true}, {`probe(5);`, false}, {`namespace Custom;const SQLITE3_NULL=4;probe(SQLITE3_NULL);`, false},
	} {
		probeNative(t, tc.src, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := ReadsGlobalConstant(ctx, CallArgument(c.Args, 0, ""), "SQLITE3_NULL"); got != tc.want {
				t.Errorf("%s: %v", tc.src, got)
			}
		})
	}
}
