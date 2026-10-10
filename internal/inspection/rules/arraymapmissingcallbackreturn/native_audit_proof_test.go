package arraymapmissingcallbackreturn

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditProofGuards(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"array_map(function($s){if(false){trim($s);}},$names);", 0}} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayMapMissingCallbackReturn"}})
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
