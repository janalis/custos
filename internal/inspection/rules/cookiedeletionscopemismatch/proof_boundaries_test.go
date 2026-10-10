package cookiedeletionscopemismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "CookieDeletionScopeMismatch" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"setcookie(\"a\",\"\",[]);", 0}, {"setcookie(\"a\",\"\",[\"expires\"=>2]);", 0}, {"setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"echo 1;setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"strlen(\"x\");setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"setcookie(\"b\",\"x\",[]);setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"setcookie(\"a\",\"x\",$options);setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"setcookie(\"a\",\"x\",[...$options]);setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"setcookie(\"a\",\"x\",0);setcookie(\"a\",\"\",1);", 0}, {"setcookie(\"a\",\"x\",[\"path\"=>$unknown]);setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"setcookie(\"a\",\"x\",[\"domain\"=>$unknown]);setcookie(\"a\",\"\",[\"expires\"=>1]);", 0}, {"echo setcookie(\"a\",\"\",1);", 0}, {"$x=1;setcookie(\"a\",\"\",1);", 0}, {"setcookie(\"a\",\"x\");setcookie(\"a\",\"\",1);", 0}, {"setcookie(\"a\",\"x\",[$unknown=>1]);setcookie(\"a\",\"\",1);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CookieDeletionScopeMismatch"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
