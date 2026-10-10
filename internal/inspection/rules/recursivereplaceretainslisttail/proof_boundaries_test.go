package recursivereplaceretainslisttail

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "RecursiveReplaceRetainsListTail" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"array_replace_recursive($a,$b);", 0}, {"array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]);", 0}, {"$x=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]);", 0}, {"$x=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]); echo $x;", 0}, {"$x=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]); echo $x[\"k\"][$unknown];", 0}, {"$x=array_replace_recursive([\"k\"=>[1]],[\"q\"=>[2]]); echo $x[\"k\"][0];", 0}, {"$x=array_replace_recursive([\"k\"=>[\"x\"=>1]],[\"k\"=>[]]); echo $x[\"k\"][0];", 0}, {"echo array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]);", 0}, {"foo($a=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]));", 0}, {"$a=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]);echo $a[0];", 0}, {"$a=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]);echo $a[\"k\"][0],1;", 0}, {"$obj->p=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]);echo $obj->p[\"k\"][0];", 0}, {"$a=array_replace_recursive([\"k\"=>1],[\"k\"=>[]]);echo $a[\"k\"][0];", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"RecursiveReplaceRetainsListTail"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
