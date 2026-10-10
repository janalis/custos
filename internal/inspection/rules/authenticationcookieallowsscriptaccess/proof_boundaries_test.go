package authenticationcookieallowsscriptaccess

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "AuthenticationCookieAllowsScriptAccess" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"setcookie($name,$x,[]);", 0}, {"setcookie(\"other\",$x,[\"httponly\"=>false]);", 0}, {"setcookie(\"auth_token\",$x,[\"httponly\"=>$unknown]);", 0}, {"setcookie(\"auth_token\",$x,$options);", 0}, {"setcookie(\"auth_token\",$x,[...$options]);", 0}, {"setcookie(\"auth_token\",$x,0,\"/\",\"\",true,true);", 0}, {"setcookie(\"auth_token\",$x);", 0}, {"setcookie(\"auth_token\",$x,[$unknown=>false]);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"AuthenticationCookieAllowsScriptAccess"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
