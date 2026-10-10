package arrayuniondropsrightvalues

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArrayUnionDropsRightValues" || len(r.Kinds()) == 0 {
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
	}{{"$a=1+2;", 0}, {"$a=[1=>1]+[1];", 0}, {"[1]+[2];", 0}, {"echo [1]+[2];", 0}, {"$a=[1]+[2];", 0}, {"$a=[1]+[2]; echo $a;", 0}, {"if($x) $a=[1]+[2];", 0}, {"$obj->p=[1]+[2]; foreach($obj->p as $v){}", 0}, {"$a=[]+[1];", 0}, {"$a=[1]-[2];", 0}, {"foo($a=[1]+[2]);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayUnionDropsRightValues"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
