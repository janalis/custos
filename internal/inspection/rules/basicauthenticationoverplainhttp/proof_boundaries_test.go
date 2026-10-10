package basicauthenticationoverplainhttp

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "BasicAuthenticationOverPlainHttp" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"curl_setopt($ch,$option,$x);", 0}, {"curl_setopt($ch,CURLOPT_TIMEOUT,1);", 0}, {"curl_setopt($ch,CURLOPT_USERPWD,$x);", 0}, {"$ch=curl_init($url);curl_setopt($ch,CURLOPT_USERPWD,$x);curl_exec($ch);", 0}, {"$ch=curl_init(\"http://example.test\");curl_setopt($ch,CURLOPT_URL,\"https://example.test\");curl_setopt($ch,CURLOPT_USERPWD,$x);curl_exec($ch);", 0}, {"$ch=curl_init(\"http://example.test\");curl_setopt($ch,CURLOPT_URL,$unknown);curl_setopt($ch,CURLOPT_USERPWD,$x);curl_exec($ch);", 0}, {"$ch=curl_init(\"http://example.test\");curl_setopt($ch,CURLOPT_USERPWD,$x);", 0}, {"$ch=curl_init(\"http://example.test\");curl_setopt($ch,CURLOPT_USERPWD,$x);echo 1;", 0}, {"$ch=curl_init(\"http://example.test\");curl_setopt($ch,CURLOPT_USERPWD,$x);unknown();", 0}, {"$ch=curl_init(\"http://example.test\");echo curl_setopt($ch,CURLOPT_USERPWD,$x);curl_exec($ch);", 0}, {"$ch=curl_init(\"http://example.test\");curl_setopt($ch,99,1);curl_setopt($ch,CURLOPT_USERPWD,$x);curl_exec($ch);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"BasicAuthenticationOverPlainHttp"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
