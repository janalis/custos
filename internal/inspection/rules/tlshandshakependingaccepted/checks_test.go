package tlshandshakependingaccepted

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"TlsHandshakePendingAccepted"}})
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
		{"\nif (stream_set_blocking($s, false)) { if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) { fwrite($s, $secret); } }\n", 1},
		{"\nif (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) === true) { fwrite($s, $secret); }\n", 0},
		{"if(stream_set_blocking($s,false)){if(stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false){echo 'ok';}}", 0},
		{"stream_set_blocking($s,false);if(stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false){fwrite($s,'x');}", 0},
		{"if(stream_set_blocking($s,true)){if(stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false){fwrite($s,'x');}}", 0},
		{"if(false){\nif (stream_set_blocking($s, false)) { if (stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT) !== false) { fwrite($s, $secret); } }\n}", 0},
		{"$ok=stream_socket_enable_crypto($s,true,STREAM_CRYPTO_METHOD_TLS_CLIENT)!==false;", 0},
		{"function f($s){if(!stream_set_blocking($s,false)){return;}if(stream_socket_enable_crypto($s,true)!==false){a: fwrite($s,'x'); goto a;}}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
