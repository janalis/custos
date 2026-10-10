package validatedintegerzerorejected

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
		{"!filter_var($v,FILTER_VALIDATE_INT);", 1},
		{"if(!filter_var($v,FILTER_VALIDATE_INT,FILTER_NULL_ON_FAILURE)){return;}", 0},
		{"if(!filter_var($v,FILTER_VALIDATE_INT,['options'=>['min_range'=>1]])){return;}", 0},
		{"if(!filter_var($v,FILTER_VALIDATE_BOOLEAN)){return;}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ValidatedIntegerZeroRejected"}})
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
