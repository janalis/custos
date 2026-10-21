package sodiumsecretboxnoncereused

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SodiumSecretboxNonceReused"}})
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
		{"\n$n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); sodium_crypto_secretbox('alpha', $n, $k); sodium_crypto_secretbox('beta', $n, $k);\n", 1},
		{"\nsodium_crypto_secretbox('alpha', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k); sodium_crypto_secretbox('beta', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k);\n", 0},
		{"function f($k,$n){sodium_crypto_secretbox('a',$n,$k);sodium_crypto_secretbox('a',$n,$k);}", 0},
		{"function f($k,$n){sodium_crypto_secretbox('a',$n,$k);sodium_crypto_secretbox($m,$n,$k);}", 0},
		{"function f($k,$n){sodium_crypto_secretbox('a',$n,$k);$n=random_bytes(24);sodium_crypto_secretbox('b',$n,$k);}", 0},
		{"function f($k,$n){sodium_crypto_secretbox('a',$n,$k);$k=random_bytes(32);sodium_crypto_secretbox('b',$n,$k);}", 0},
		{"if(false){\n$n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); sodium_crypto_secretbox('alpha', $n, $k); sodium_crypto_secretbox('beta', $n, $k);\n}", 0},
		{`function f($k,$n){sodium_crypto_secretbox('a',$n,$k);$cb=function()use(&$k){$k='changed';};$cb();sodium_crypto_secretbox('b',$n,$k);}`, 0}, // state escaped through a reference
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
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP71, Only: []string{"SodiumSecretboxNonceReused"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Analyze(syntax.Parse("old.php", []byte("<?php $n=random_bytes(24);sodium_crypto_secretbox('a',$n,$k);sodium_crypto_secretbox('b',$n,$k);"), syntax.Options{})); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
