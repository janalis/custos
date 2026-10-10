package aeaddecryptiontaglengthunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "AeadDecryptionTagLengthUnchecked" || len(r.Kinds()) == 0 {
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
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: 503, Only: []string{"AeadDecryptionTagLengthUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php $p=openssl_decrypt($c,\"aes-256-gcm\",$key,OPENSSL_RAW_DATA,$iv,$_POST[\"tag\"]);"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("unavailable API reported: %+v", got)
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"strlen(\"x\");", 0}, {"openssl_decrypt($x,\"other\",$key,0,$iv,\"tag\");", 0}, {"session_id($unknown);", 0}, {"session_id($_GET[$unknown]);", 0}, {"session_id($custom[\"password\"]);", 0}, {"session_id($name);", 0}, {"openssl_decrypt($x,$cipher,$key,0,$iv,$_GET[\"tag\"]);", 0}, {"openssl_decrypt($x,\"aes-256-ccm\",$key,0,$iv,$_GET[\"tag\"]);", 0}, {"openssl_decrypt($x,\"aes-256-gcm\",$key,0,$iv,\"literal\");", 0}, {"openssl_decrypt($x,\"aes-256-gcm\",$key,0,$iv,$o->p[\"tag\"]);", 0}, {"openssl_decrypt($x,\"aes-256-gcm\",$key,0,$iv,$custom[\"tag\"]);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"AeadDecryptionTagLengthUnchecked"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
