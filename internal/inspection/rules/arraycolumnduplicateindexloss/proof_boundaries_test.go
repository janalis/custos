package arraycolumnduplicateindexloss

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArrayColumnDuplicateIndexLoss" || len(r.Kinds()) == 0 {
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
	}{{"strtoupper(\"a\");", 0}, {"array_column($unknown,\"v\",\"id\");", 0}, {"array_column([1],\"v\",\"id\");", 0}, {"array_column([[\"id\"=>1]],\"v\",\"id\");", 0}, {"array_column([[\"v\"=>1]],\"v\",\"id\");", 0}, {"array_column([[\"v\"=>1,\"id\"=>$x]],\"v\",\"id\");", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayColumnDuplicateIndexLoss"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
