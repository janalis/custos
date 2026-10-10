package curltransportsuccessashttpsuccess

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function bad($url) { $ch = curl_init('https://example.invalid/api'); if (curl_exec($ch) !== false) { return true; } return false; }", 1},
		{"function good($url) { $ch = curl_init('https://example.invalid/api'); curl_setopt($ch, CURLOPT_FAILONERROR, true); return curl_exec($ch) !== false; }", 0},
		{"namespace Custom; function unrelated($x) {} unrelated(1);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlTransportSuccessAsHttpSuccess"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
			for _, g := range got {
				if g.Message != message || g.Span.End <= g.Span.Start {
					t.Fatalf("invalid finding: %+v", g)
				}
			}
		})
	}
}
