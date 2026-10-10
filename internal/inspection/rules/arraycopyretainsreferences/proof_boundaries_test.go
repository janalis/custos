package arraycopyretainsreferences

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "ArrayCopyRetainsReferences" || len(r.Kinds()) == 0 {
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
	}{{"$a[0]=1;", 0}, {"echo $a[0]=1;", 0}, {"echo 1; $a[0]=1;", 0}, {"$a=[1]; $r=&$a[0]; $b=$a; $b[1]=9; echo $a[0];", 0}, {"$a=[1]; $r=&$a[0]; $b=$a; $b[0]=9;", 0}, {"$a=[1]; $r=&$a[0]; $b=$a; $b[0]=9; echo $b[0];", 0}, {"$a=[1]; $r=&$a[0]; $b=$a; $b[0]=9; echo $a[$unknown];", 0}, {"$a=[1]; echo 1; $b=$a; $b[0]=9; echo $a[0];", 0}, {"$a=$unknown; $r=&$a[0]; $b=$a; $b[0]=9; echo $a[0];", 0}, {"if($x) $a[0]=1;", 0}, {"$b=1;$b[0]=9;", 0}, {"$b=$a;$b[0]=9;", 0}, {"$r=&$a;$b=$a;$b[0]=9;", 0}, {"$r=&$a[0];$b=$a;$b[0]=9;", 0}, {"if($x){} $r=&$a[0];$b=$a;$b[0]=9;", 0}, {"$a=[1];$r=&$a[0];$b=$a;$b[0]=9;echo $a[0],1;", 0}, {"$obj->p[0]=9;", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayCopyRetainsReferences"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
