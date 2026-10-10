package foreachreferencesurvivesloop

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ForeachReferenceSurvivesLoop" || len(r.Kinds()) == 0 {
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
	}{{"foreach($xs as $x){} $x=1;", 0}, {"if($a) foreach($xs as &$x){}", 0}, {"foreach($xs as &$x){} unset($other); $x=1;", 0}, {"foreach($xs as &$x){} echo $x;", 0}, {"foreach($xs as &$x){} $obj->p=1; $x=1;", 0}, {"foreach($xs as &$x){} $y=$x; $x=1;", 0}, {"foreach($xs as &$x){} $y=1; $x=&$other;", 0}, {"foreach($xs as &$x){} $y=1;", 0}, {"foreach($xs as &$x){} work();", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ForeachReferenceSurvivesLoop"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
