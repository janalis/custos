package opensslsigningfailureignored

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"OpenSslSigningFailureIgnored"}})
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
		{"\nopenssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256); echo base64_encode($signature);\n", 1},
		{"\nif (openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256)) { echo base64_encode($signature); }\n", 0},
		{"openssl_sign($d,$s,$k);echo $s;", 1},
		{"openssl_sign($d,$s,$k);file_put_contents($p,$s);", 1},
		{"openssl_sign($d,$s,$k);$s='changed';echo base64_encode($s);", 0},
		{"$ok=openssl_sign($d,$s,$k);echo base64_encode($s);", 0},
		{"openssl_sign($d,$s,$k);unrelated();", 0},
		{"if(false){\nopenssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256); echo base64_encode($signature);\n}", 0},
		{"function f($d,$k){$unused=function(){};openssl_sign($d,$s,$k);echo base64_encode($s);}", 1},
		{"function f($data,$key){" + strings.Repeat("other();", 513) + "openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256); echo base64_encode($signature);}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
