package exhaustedgeneratorreused

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditProofGuards(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"function source(){yield 1;yield 2;}$g=source();iterator_to_array($g);$g=source();foreach($g as $x){}", 0}} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ExhaustedGeneratorReused"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
