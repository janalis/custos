package filereadfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditProofGuards(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f(){$x=file_get_contents('/tmp/f');if($x===false){return;}trim($x);}", 0},
		{"function f(){$x=stream_get_contents($h);if($x!==false){trim($x);}}", 0},
		{"namespace Audit;function file_get_contents($s){return 'ok';}trim(file_get_contents('/tmp/f'));", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FileReadFailureUnchecked"}})
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
