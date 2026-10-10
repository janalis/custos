package catchvariableoverwriteslocal

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "CatchVariableOverwritesLocal" || len(r.Kinds()) == 0 {
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
	}{{"try{work();}catch(Throwable $x){} echo $x;", 0}, {"echo 1; try{work();}catch(Throwable $x){} echo $x;", 0}, {"$x=unknown(); try{work();}catch(Throwable $x){} echo $x;", 0}, {"$x=1; try{}catch(Throwable $x){} echo $x;", 0}, {"$x=1; try{echo 1;}catch(Throwable $x){} echo $x;", 0}, {"$x=1; try{return;}catch(Throwable $x){} echo $x;", 0}, {"$x=1; try{work();}catch(Throwable $x){}", 0}, {"if($a) try{work();}catch(Throwable $x){}", 0}, {"$x=1; try{work();}catch(Throwable $x){} echo $other;", 0}, {"$x+=1;try{work();}catch(Throwable $x){}echo $x;", 0}, {"$obj->p=1;try{work();}catch(Throwable $x){}echo $x;", 0}, {"$x=1;try{$y=1;}catch(Throwable $x){}echo $x;", 0}, {"$x=1;try{work();}catch(Throwable $x){}echo $x,1;", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CatchVariableOverwritesLocal"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
