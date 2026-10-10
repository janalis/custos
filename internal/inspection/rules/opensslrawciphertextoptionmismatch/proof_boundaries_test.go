package opensslrawciphertextoptionmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "OpenSslRawCiphertextOptionMismatch" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"openssl_decrypt($unknown,\"aes-256-gcm\",$key);", 0}, {"$c=unknown();openssl_decrypt($c,\"aes-256-gcm\",$key);", 0}, {"$c=openssl_encrypt($x,\"aes-128-cbc\",\"key\",$flags,\"iv\");openssl_decrypt($c,\"aes-128-cbc\",\"key\",0,\"iv\");", 0}, {"$c=openssl_encrypt($x,\"aes-128-cbc\",\"key\",1,\"iv\");openssl_decrypt($c,\"aes-192-cbc\",\"key\",0,\"iv\");", 0}, {"$c=openssl_encrypt($x,\"aes-128-cbc\",$key,1,\"iv\");openssl_decrypt($c,\"aes-128-cbc\",$key,0,\"iv\");", 0}, {"$key=\"a\";$c=openssl_encrypt($x,\"aes-128-cbc\",$key,1,\"iv\");$key=\"b\";openssl_decrypt($c,\"aes-128-cbc\",$key,0,\"iv\");", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"OpenSslRawCiphertextOptionMismatch"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
