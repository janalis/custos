package passworduseddirectlyasencryptionkey

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "PasswordUsedDirectlyAsEncryptionKey" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"openssl_decrypt($x,\"other\",$key,0,$iv,\"tag\");", 0}, {"session_id($unknown);", 0}, {"session_id($_GET[$unknown]);", 0}, {"session_id($custom[\"password\"]);", 0}, {"session_id($name);", 0}, {"openssl_encrypt($x,\"aes-256-gcm\",$o->p[\"password\"],0,$iv,$tag);", 0}, {"openssl_encrypt($x,\"aes-256-gcm\",$custom[\"password\"],0,$iv,$tag);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PasswordUsedDirectlyAsEncryptionKey"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
