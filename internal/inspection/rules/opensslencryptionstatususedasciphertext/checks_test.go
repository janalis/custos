package opensslencryptionstatususedasciphertext

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"OpenSslEncryptionStatusUsedAsCiphertext"}})
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
		{"\n$cipher = openssl_public_encrypt($data, $out, $key); echo base64_encode($cipher);\n", 1},
		{"\nif (openssl_public_encrypt($data, $out, $key)) { echo base64_encode($out); }\n", 0},
		{"$ok=openssl_private_encrypt($d,$out,$k);echo $ok;", 1},
		{"$out=openssl_public_encrypt($d,$out,$k);echo $out;", 0},
		{"$ok=openssl_public_encrypt($d,$out,$k);file_put_contents($p,$ok);", 1},
		{"$ok=openssl_public_encrypt($d,$out,$k);$ok='x';echo base64_encode($ok);", 0},
		{"if(false){\n$cipher = openssl_public_encrypt($data, $out, $key); echo base64_encode($cipher);\n}", 0},
		{"echo openssl_public_encrypt($d,$out[0],$k);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
