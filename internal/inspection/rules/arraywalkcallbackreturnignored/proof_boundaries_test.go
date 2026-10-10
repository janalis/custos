package arraywalkcallbackreturnignored

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArrayWalkCallbackReturnIgnored" || len(r.Kinds()) == 0 {
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
	}{{"array_walk($x,fn($a)=>strtoupper($a));", 0}, {"count([]);", 0}, {"array_walk($x,$unknown);", 0}, {"array_walk($x,function(string $a){return strtoupper($a);});", 1}, {"array_walk($x,function($a){});", 0}, {"array_walk($x,fn($a)=>\"x\");", 1}, {"array_walk($x,fn($a)=>unknown($a));", 0}, {"array_walk($x,fn($a)=>strtoupper(...$args));", 0}, {"array_walk($x,fn($a)=>$a+1);", 0}, {"array_walk($x,fn(string $a)=>trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim(trim($a)))))))))))))))))));", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayWalkCallbackReturnIgnored"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
