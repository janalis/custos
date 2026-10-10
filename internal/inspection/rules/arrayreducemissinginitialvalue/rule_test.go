package arrayreducemissinginitialvalue

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
		{"array_reduce($parts, fn($a,$v)=>array_merge($a,[$v]));", 1},
		{"array_reduce($parts, function($a,$v){return array_merge($a,[$v]);});", 1},
		{"array_reduce($parts, fn($a,$v)=>array_merge($a,[$v]), []);", 0},
		{"array_reduce($parts, fn($a,$v)=>$a+$v);", 0},
		{"array_reduce($parts, fn($a,$v)=>array_merge($a??[],[$v]));", 0},
		{"array_reduce($parts, function($a,$v){$a=[];return array_merge($a,[$v]);});", 0},
		{"array_reduce($parts, function($a,$v){array_merge($a,[$v]);});", 0},
		{"array_reduce($parts,$callback);", 0},
		{"array_reduce(callback:fn($a,$v)=>array_merge($a,[$v]));", 0},
		{"array_reduce(array:$parts,callback:fn($a,$v)=>array_merge($a,[$v]));", 1},
		{"array_reduce($parts,function(){return array_merge([],[]);});", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayReduceMissingInitialValue"}})
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
