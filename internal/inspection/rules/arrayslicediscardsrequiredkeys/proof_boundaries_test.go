package arrayslicediscardsrequiredkeys

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArraySliceDiscardsRequiredKeys" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"array_slice($a,0);", 0}, {"array_slice([1],0);", 0}, {"array_slice([\"x\"=>1],0);", 0}, {"$a=array_slice([100=>\"a\"],-9,-9); echo $a[100];", 0}, {"$a=array_slice([100=>\"a\"],9,1); echo $a[100];", 0}, {"$a=array_slice([100=>\"a\"],-1,$unknown); echo $a[100];", 0}, {"$a=array_slice([100=>\"a\"],0,1);", 0}, {"$a=array_slice([100=>\"a\"],0,1); echo $a[0];", 0}, {"echo array_slice([100=>\"a\"],0,1);", 0}, {"if($x) $a=array_slice([100=>\"a\"],0,1);", 0}, {"foo($a=array_slice([100=>\"a\"],0,1));", 0}, {"$a=array_slice([100=>\"a\"],0,1);echo $a[100],1;", 0}, {"$obj->p=array_slice([100=>\"a\"],0,1);echo $obj->p[100];", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArraySliceDiscardsRequiredKeys"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
