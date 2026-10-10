package credentialedcorsuseswildcardorigin

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "CredentialedCorsUsesWildcardOrigin" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"header(\"X-Test: yes\");", 0}, {"header(\"Access-Control-Allow-Origin: *\");header(\"Access-Control-Allow-Credentials: true\");header_remove(\"Access-Control-Allow-Origin\");", 0}, {"header(\"Access-Control-Allow-Origin: *\");header(\"Access-Control-Allow-Credentials: true\");header_remove();", 0}, {"header(\"Access-Control-Allow-Origin: *\");header(\"Access-Control-Allow-Credentials: true\");header_remove($unknown);", 0}, {"header(\"Access-Control-Allow-Origin: *\");header(\"Access-Control-Allow-Credentials: true\",false);", 0}, {"header(\"Access-Control-Allow-Origin: *\");header(\"Access-Control-Allow-Credentials: true\");unknown();", 0}, {"header($unknown);", 0}, {"header(\"HTTP/1.1 200 OK\");", 0}, {"echo header(\"X-Test: yes\");", 0}, {"if($x) header(\"X-Test: yes\");", 0}, {"$x=1;header(\"X-Test: yes\");", 0}, {"echo 1;header(\"X-Test: yes\");", 0}, {"header(\"X-Test: yes\");echo 1;", 0}, {"header(\"X-Test: yes\");$x=1;", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CredentialedCorsUsesWildcardOrigin"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
