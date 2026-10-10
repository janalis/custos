package jsonnumericcheckchangesidentifiers

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "JsonNumericCheckChangesIdentifiers" || len(r.Kinds()) == 0 {
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
	}{{"json_decode($x);", 0}, {"json_encode([\"n\"=>\"+x\"],JSON_NUMERIC_CHECK);", 0}, {"json_encode([\"n\"=>\"0x\"],JSON_NUMERIC_CHECK);", 0}, {"json_encode([\"n\"=>\"1\"],0);", 0}, {"json_encode([[[[[[[[[[[[[[[[[[\"+123\"]]]]]]]]]]]]]]]]]],JSON_NUMERIC_CHECK);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonNumericCheckChangesIdentifiers"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
