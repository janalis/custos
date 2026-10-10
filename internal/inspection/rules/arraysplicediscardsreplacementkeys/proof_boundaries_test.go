package arraysplicediscardsreplacementkeys

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArraySpliceDiscardsReplacementKeys" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"array_splice($a,0,0,[\"k\"=>1]);", 0}, {"$a=[]; array_splice($a,0,0,[\"k\"=>1]);", 0}, {"$a=[]; echo array_splice($a,0,0,[\"k\"=>1]);", 0}, {"$a=[\"k\"=>1]; array_splice($a,0,0,[\"k\"=>2]); echo $a[\"k\"];", 0}, {"$a=[]; array_splice($a,0,0,[\"q\"=>1]); echo $a[\"k\"];", 0}, {"$a=[]; array_splice($a,0,0,[\"q\"=>1]); echo $a[$unknown];", 0}, {"$a=[]; array_splice($a,0,0,[\"q\"=>1]); echo $b[\"q\"];", 0}, {"array_splice($obj->p,0,0,[\"k\"=>1]);", 0}, {"$a=[];array_splice($a,0,0,[\"k\"=>1]);echo $a[\"k\"],1;", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArraySpliceDiscardsReplacementKeys"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
