package usortdiscardsrequiredkeys

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"usort($unknown,$cmp);", 0},
		{"$a=['k'=>1]; $success=usort($a,$cmp);", 0},
		{"$a=['k'=>1]; usort($a,$cmp);", 0},
		{"$a=['k'=>1]; usort($a,$cmp); echo $a[$unknown];", 0},
		{"$a=['k'=>1]; usort($a,$cmp); echo function(){return $a['k'];};", 0},
		{"$a=['k'=>1]; usort($a,$cmp); echo $a['k'];", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UsortDiscardsRequiredKeys"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
