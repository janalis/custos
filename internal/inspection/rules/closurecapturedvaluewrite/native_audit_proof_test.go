package closurecapturedvaluewrite

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func TestAuditProofGuards(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"$n=0;$f=function()use($n){if(false){++$n;}};$f();echo $n;", 0}} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ClosureCapturedValueWrite"}})
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
