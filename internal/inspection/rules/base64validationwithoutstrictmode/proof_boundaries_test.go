package base64validationwithoutstrictmode

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "Base64ValidationWithoutStrictMode" || len(r.Kinds()) == 0 {
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
	}{{"if($x){}", 0}, {"if($x==false){return false;}", 0}, {"if(false===1){return false;}", 0}, {"if(false===base64_decode($bytes)){echo \"no\";}", 0}, {"if(base64_decode($bytes,$strict)===false){return false;}", 0}, {"if(base64_decode($bytes)===false) return false;", 0}, {"if(base64_decode($bytes)===false){}", 0}, {"if(base64_decode($bytes)===false){throw new RuntimeException();}", 1}, {"if(base64_decode($bytes)===false){return true;}", 0}, {"if(base64_decode($bytes)===false){$x=1;}", 0}, {"if($x===$y){return false;}", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"Base64ValidationWithoutStrictMode"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
