package dynamiccorsoriginwithoutvary

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "DynamicCorsOriginWithoutVary" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"header(\"X-Test: yes\");", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);unknown();", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header($unknown);", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"HTTP/1.1 200 OK\");", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"Access-Control-Allow-Origin: https://safe.test\");", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"Vary: Accept\");", 0}, {"echo header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);", 0}, {"if($x) header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);", 0}, {"header(\"Access-Control-Allow-Origin: \".$custom[\"HTTP_ORIGIN\"]);", 0}, {"header(\"Access-Control-Allow-Origin: \".$o->p[\"HTTP_ORIGIN\"]);", 0}, {"header(\"Access-Control-Allow-Origin: \".$unknown);", 0}, {"header(\"Other: \".$_SERVER[\"HTTP_ORIGIN\"]);", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);echo 1;", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"Cache-Control: private,max-age=30\");", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"Cache-Control: max-age=invalid\");", 0}, {"header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"Cache-Control: must-revalidate\");", 0}, {"header(\"Access-Control-Allow-Origin: \".$_GET[\"HTTP_ORIGIN\"]);header(\"Cache-Control:public\");", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DynamicCorsOriginWithoutVary"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
