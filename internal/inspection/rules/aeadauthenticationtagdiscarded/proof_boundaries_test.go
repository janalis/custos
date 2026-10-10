package aeadauthenticationtagdiscarded

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "AeadAuthenticationTagDiscarded" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestUnavailableVersion(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: 503, Only: []string{"AeadAuthenticationTagDiscarded"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php $c=openssl_encrypt($s,\"aes-256-gcm\",$key,OPENSSL_RAW_DATA,$iv);"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unavailable API reported: %+v", got)
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"strlen(\"x\");", 0}, {"openssl_encrypt($x,$cipher,$key);", 0}, {"openssl_encrypt($x,\"aes-128-cbc\",$key);", 0}, {"openssl_encrypt(...$args);", 0}, {"openssl_encrypt(data:$x,cipher_algo:\"aes-256-gcm\",passphrase:$key,...$args);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"AeadAuthenticationTagDiscarded"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
