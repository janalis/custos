package sodiumkdfcontextlengthmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SodiumKdfContextLengthMismatch"}})
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
		{"\nsodium_crypto_kdf_derive_from_key(32, 2, 'short', $key);\n", 1},
		{"\nsodium_crypto_kdf_derive_from_key(32, 2, 'APPKEY01', $key);\n", 0},
		{"sodium_crypto_kdf_derive_from_key(32,1,str_repeat('x',7),$k);", 1},
		{"sodium_crypto_kdf_derive_from_key(32,1,$c,$k);", 0},
		{"if(false){\nsodium_crypto_kdf_derive_from_key(32, 2, 'short', $key);\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
