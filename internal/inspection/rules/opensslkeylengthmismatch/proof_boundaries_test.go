package opensslkeylengthmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "OpenSslKeyLengthMismatch" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"openssl_encrypt($d,$cipher,$key);", 0}, {"openssl_encrypt($d,\"other\",$key);", 0}, {"openssl_encrypt($d,\"aes-512-gcm\",$key);", 0}, {"openssl_encrypt($d,\"aes-256-unknown\",$key);", 0}, {"openssl_encrypt($d,\"aes-128-cbc\",$key);", 0}, {"openssl_encrypt($d,\"aes-192-cbc\",str_repeat(\"x\",24));", 0}, {"openssl_encrypt($d,\"aes-256-cbc\",str_repeat(\"x\",32));", 0}, {"openssl_encrypt($d,\"aes-256-cbc\",unknown());", 0}, {"openssl_encrypt($d,\"aes-256-cbc\",str_repeat($unknown,32));", 0}, {"openssl_encrypt($d,\"aes-256-cbc\",random_bytes($unknown));", 0}, {"openssl_encrypt($d,\"aes-128-cbc\",\"0123456789012345\");", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"OpenSslKeyLengthMismatch"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
