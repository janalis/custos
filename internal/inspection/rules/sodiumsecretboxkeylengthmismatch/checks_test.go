package sodiumsecretboxkeylengthmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SodiumSecretboxKeyLengthMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\nsodium_crypto_secretbox($message, $nonce, random_bytes(16));\n", 1},
		{"\nsodium_crypto_secretbox($message, $nonce, sodium_crypto_secretbox_keygen());\n", 0},
		{"sodium_crypto_secretbox_open($m,$n,str_repeat('x',16));", 1},
		{"sodium_crypto_secretbox($m,$n,str_repeat('xx',16));", 0},
		{"sodium_crypto_secretbox($m,$n,$k);", 0},
		{"sodium_crypto_secretbox($m,$n,str_repeat($s,16));", 0},
		{"sodium_crypto_secretbox($m,$n,random_bytes(0));", 0},
		{"if(false){\nsodium_crypto_secretbox($message, $nonce, random_bytes(16));\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

// Before PHP 7.2 the sodium functions do not exist and the rule stays silent.
func TestBeforeSodium(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP71, Only: []string{"SodiumSecretboxKeyLengthMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Analyze(syntax.Parse("old.php", []byte("<?php sodium_crypto_secretbox('m', random_bytes(24), 'short');"), syntax.Options{})); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
