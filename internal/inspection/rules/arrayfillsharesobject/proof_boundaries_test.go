package arrayfillsharesobject

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArrayFillSharesObject" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"array_fill(0,3,\"x\");", 0}, {"array_fill(0,3,new $name);", 0}, {"array_fill(0,3,new Missing);", 0}, {"array_fill(0,1,new stdClass);", 0}, {"array_fill(0,3,new stdClass);", 0}, {"echo array_fill(0,3,new stdClass);", 0}, {"$a=array_fill(0,3,new stdClass);", 0}, {"$a=array_fill(0,3,new stdClass); echo 1;", 0}, {"$a=array_fill(0,3,new stdClass); $a[0]=1;", 0}, {"$a=array_fill(0,3,new stdClass); $a[0]->p=1;", 0}, {"$a=array_fill(0,3,new stdClass); $a[9]->p=1; echo $a[0]->p;", 0}, {"$a=array_fill(0,3,new stdClass); $a[0]->p=1; echo 1;", 0}, {"$a=array_fill(0,3,new stdClass); $a[0]->p=1; echo $other[1]->p;", 0}, {"foo($a=array_fill(0,3,new stdClass));", 0}, {"$a=array_fill(0,3,new stdClass);work();", 0}, {"$a=array_fill(0,3,new stdClass);$obj->p=1;echo $a[1]->p;", 0}, {"$a=array_fill(0,3,new stdClass);$a[0]->p=1;echo $a[1]->p,1;", 0}, {"$obj->p=array_fill(0,3,new stdClass);$obj->p[0]->p=1;echo $obj->p[1]->p;", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayFillSharesObject"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
