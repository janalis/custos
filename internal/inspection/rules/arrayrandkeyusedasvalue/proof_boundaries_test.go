package arrayrandkeyusedasvalue

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArrayRandKeyUsedAsValue" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"array_rand([\"a\"],2);", 0}, {"array_rand($unknown);", 0}, {"array_rand([1]);", 0}, {"array_rand([\"k\"=>\"v\"]);", 0}, {"array_rand([\"a\"]);", 0}, {"unknown(array_rand([\"a\"]));", 0}, {"function c($x){echo $x;} c(array_rand([\"a\"]));", 0}, {"function c($x){switch($x){default:break;}} c(array_rand([\"a\"]));", 0}, {"function c($x){switch($x){case \"b\":break;}} c(array_rand([\"a\"]));", 0}, {"function c($x){switch($x){case 1:break;}} c(array_rand([\"a\"]));", 0}, {"new stdClass(array_rand([\"a\"]));", 0}, {"function c($x){switch(\"a\"){case \"a\":break;}} c(array_rand([\"a\"]));", 0}, {"function c($x){switch($y){case \"a\":break;}} c(array_rand([\"a\"]));", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayRandKeyUsedAsValue"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
