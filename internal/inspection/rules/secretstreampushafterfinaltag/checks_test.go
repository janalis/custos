package secretstreampushafterfinaltag

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SecretstreamPushAfterFinalTag"}})
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
		{"\nsodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'more');\n", 1},
		{"\nsodium_crypto_secretstream_xchacha20poly1305_push($s, 'more'); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL);\n", 0},
		{"function f($s){sodium_crypto_secretstream_xchacha20poly1305_push($s,'a','',0);sodium_crypto_secretstream_xchacha20poly1305_push($s,'b');}", 0},
		{"function f($s){if($flag){sodium_crypto_secretstream_xchacha20poly1305_push($s,'a','',3);}sodium_crypto_secretstream_xchacha20poly1305_push($s,'b');}", 0},
		{"if(false){\nsodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'more');\n}", 0},
		{"function f($s){sodium_crypto_secretstream_xchacha20poly1305_push($s,'a','',$tag);sodium_crypto_secretstream_xchacha20poly1305_push($s,'b');}", 0},
		{"unknown();", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
