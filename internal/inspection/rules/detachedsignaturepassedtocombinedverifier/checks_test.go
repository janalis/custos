package detachedsignaturepassedtocombinedverifier

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DetachedSignaturePassedToCombinedVerifier"}})
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
		{"\n$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_open($sig, $public);\n", 1},
		{"\n$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_verify_detached($sig, $message, $public);\n", 0},
		{"$s=sodium_crypto_sign_detached($m,$k);sodium_crypto_sign_open($s.$m,$p);", 0},
		{"sodium_crypto_sign_open($s,$p);", 0},
		{"if(false){\n$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_open($sig, $public);\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
