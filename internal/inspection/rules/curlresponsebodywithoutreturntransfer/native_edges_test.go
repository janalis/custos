package curlresponsebodywithoutreturntransfer

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"json_decode(strtolower('fixed'));", 0},
		{"$h=new stdClass();json_decode(curl_exec($h));", 0},
		{"$body=curl_exec($unknown);json_decode($body);", 0},
		{"$h=curl_init('https://example.invalid');retain($h);$body=curl_exec($h);json_decode($body);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlResponseBodyWithoutReturnTransfer"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
