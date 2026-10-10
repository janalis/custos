package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionBcOperands(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"bcadd('1','2');", 2},
		{"bcsub('1','2');", 2},
		{"bcmul('1','2');", 2},
		{"bcdiv('1','2');", 2},
		{"bcmod('1','2');", 2},
		{"bccomp('1','2');", 2},
		{"bcpow('1',2);", 1},
		{"bcsqrt('1');", 1},
		{"bcpowmod('1','2','3');", 2},
		{"strlen('1');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			called := false
			p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
				called = true
				if got := len(ExpansionBcOperands(ctx, c)); got != tc.want {
					t.Fatalf("got %d want %d", got, tc.want)
				}
			}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.source), syntax.Options{}))
			if !called {
				t.Fatal("missing call")
			}
		})
	}
}

func BenchmarkExpansionBcOperands(b *testing.B) {
	file := syntax.Parse("test.php", []byte("<?php bcadd('1','2'); bcsub('3','4'); bcpowmod('5','6','7');"), syntax.Options{})
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) { ExpansionBcOperands(ctx, c) }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
