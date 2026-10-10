package sodiumnoncelengthmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
		php    string
	}{
		{"other(); $x->other(); echo $x[0];", 0, ""},
		{"sodium_crypto_secretbox($m,\"short\",$k);", 1, ""},
		{"sodium_crypto_secretbox($m,$nonce,$k);", 0, ""},
		{"sodium_crypto_secretbox($m,str_repeat(\"a\",3),$k);", 1, ""},
		{"sodium_crypto_secretbox($m,str_repeat($s,3),$k);", 0, ""},
		{"sodium_crypto_secretbox($m,str_repeat(\"a\",-1),$k);", 0, ""},
		{"sodium_crypto_secretbox($m,random_bytes(0),$k);", 0, ""},
		{"sodium_crypto_secretbox($m,random_bytes(+12),$k);", 1, ""},
		{"$n=12;sodium_crypto_secretbox($m,random_bytes($n),$k);", 1, ""},
		{"sodium_crypto_secretbox($m,strrev(\"short\"),$k);", 0, ""},
		{"sodium_crypto_secretbox($m,random_bytes(12),$k);", 0, "7.1"},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"SodiumNonceLengthMismatch"}, PHP: phpversion.MustParse(tc.php)}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edge.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
