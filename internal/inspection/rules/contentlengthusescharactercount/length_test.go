package contentlengthusescharactercount

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestLengthFixEvaluation(t *testing.T) {
	for _, tc := range []struct {
		src   string
		fixes int
		want  int
	}{
		{"header('Content-Length: '.mb_strlen($body,'UTF-8'));", 1, 1},
		{"header('Content-Length: '.mb_strlen($body,chooseEncoding()));", 0, 1},
		{"header('Content-Length: '.mb_strlen(...$args));", 0, 0},
		{"header('Content-Length: '.mb_strlen($body,&$encoding));", 0, 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentLengthUsesCharacterCount"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("length.php", []byte("<?php function f($body,$encoding,$args){"+tc.src+"echo $body;}"), syntax.Options{}))
			if len(got) != tc.want || (len(got) > 0 && len(got[0].Fixes) != tc.fixes) {
				t.Fatalf("findings %+v, want one with %d fixes", got, tc.fixes)
			}
		})
	}
}
