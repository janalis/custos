package unserializefalseambiguity

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "UnserializeFalseAmbiguity" || len(r.Kinds()) == 0 {
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
	}{{"if($x){}", 0}, {"if($x==false){return false;}", 0}, {"if(false===1){return false;}", 0}, {"if(false===unserialize($bytes)){echo \"no\";}", 0}, {"if(unserialize($bytes)===false) return false;", 0}, {"if(unserialize($bytes)===false){}", 0}, {"if(unserialize($bytes)===false){return true;}", 0}, {"if(unserialize($bytes)===false){$x=1;}", 0}, {"if($x===$y){return false;}", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnserializeFalseAmbiguity"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
