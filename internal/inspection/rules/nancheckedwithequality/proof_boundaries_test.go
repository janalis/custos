package nancheckedwithequality

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "NanCheckedWithEquality" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"if($x!==NAN){}", 1}, {"if(NAN==$x){}", 1}, {"if(NAN===1){}", 1}, {"if($x===2){}", 0}, {"function f(float $x){if($x /* keep */ === NAN){}}", 1}, {"function f(float $x){if($x!==NAN){}}", 1}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NanCheckedWithEquality"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
