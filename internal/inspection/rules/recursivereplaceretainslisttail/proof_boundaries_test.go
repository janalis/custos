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
	}{
		// Later replacement arrays take part in the proof.
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],[\"h\"=>[1=>\"z\"]]); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],[\"h\"=>[\"x\",\"z\"]]); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],[\"h\"=>\"gone\"]); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],$more); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],...$more); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],extra: [\"h\"=>[]]); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"p\"=>1],[\"q\"=>2]); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],[\"p\"=>1]); echo $d[\"h\"][1];", 1},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"p\"=>1],[\"h\"=>[\"c\"]]); echo $d[\"h\"][1];", 1},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]],[\"h\"=>[]]); echo $d[\"h\"][1];", 1},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]]); echo $d[\"h\"][1];", 0},
		{"$d=array_replace_recursive([\"h\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]]); echo $d[\"h\"][2];", 0},
		{"$d=array_replace_recursive([\"q\"=>[\"a\",\"b\"]],[\"h\"=>[\"c\"]]); echo $d[\"h\"][1];", 0},
		{"strlen(\"x\");", 0},
		{"array_replace_recursive($a,$b);", 0},
		{"array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]);", 0},
		{"$x=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]);", 0},
		{"$x=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]); echo $x;", 0},
		{"$x=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]); echo $x[\"k\"][$unknown];", 0},
		{"$x=array_replace_recursive([\"k\"=>[1]],[\"q\"=>[2]]); echo $x[\"k\"][0];", 0},
		{"$x=array_replace_recursive([\"k\"=>[\"x\"=>1]],[\"k\"=>[]]); echo $x[\"k\"][0];", 0},
		{"echo array_replace_recursive([\"k\"=>[1]],[\"k\"=>[2]]);", 0},
		{"foo($a=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]));", 0},
		{"$a=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]);echo $a[0];", 0},
		{"$a=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]);echo $a[\"k\"][0],1;", 0},
		{"$obj->p=array_replace_recursive([\"k\"=>[1]],[\"k\"=>[]]);echo $obj->p[\"k\"][0];", 0},
		{"$a=array_replace_recursive([\"k\"=>1],[\"k\"=>[]]);echo $a[\"k\"][0];", 0},
	} {
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
