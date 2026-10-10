package arraymapmissingcallbackreturn

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRule(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"array_map(function($v){trim($v);}, $names);", 1},
		{"array_map(function($v){return trim($v);},$names);", 0},
		{"array_map(fn($v)=>trim($v),$names);", 0},
		{"array_map(function($v){echo $v;},$names);", 0},
		{"array_map(function($v){trim($v);return null;},$names);", 0},
		{"array_map(function($v){trim($v);yield $v;},$names);", 0},
		{"array_map(function($v){$f=function(){return 1;};trim($v);},$names);", 1},
		{"array_map(function($v){foo($v);},$names);", 0},
		{"array_map($callback,$names);", 0},
		{"array_map(function($v){trim($v);strtoupper($v);},$names);", 1},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayMapMissingCallbackReturn"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
			for _, g := range got {
				if g.Message != message || g.Span.End <= g.Span.Start {
					t.Fatalf("bad finding: %+v", g)
				}
			}
		})
	}
}
