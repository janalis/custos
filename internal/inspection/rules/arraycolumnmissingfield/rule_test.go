package arraycolumnmissingfield

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
		{"array_column([[\"id\"=>1],[\"name\"=>\"Ada\"]],\"id\");", 1},
		{"array_column([[\"id\"=>1],[\"id\"=>2]],\"id\");", 0},
		{"array_column([[\"name\"=>1],[\"name\"=>2]],\"id\");", 0},
		{"array_column([[\"id\"=>1],$row],\"id\");", 0},
		{"array_column($rows,\"id\");", 0},
		{"array_column([[\"id\"=>1]],$key);", 0},
		{"array_column([[0=>1],[1=>2]],0);", 1},
		{"array_column([0=>['id'=>1],0=>['name'=>'Ada']], 'id');", 0},
		{"array_column([0=>['name'=>'Ada'],0=>['id'=>1]], 'id');", 0},
		{"namespace Custom; function other($a) {} other([1]);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayColumnMissingField"}})
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
