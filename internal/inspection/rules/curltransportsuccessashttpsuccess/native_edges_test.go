package curltransportsuccessashttpsuccess

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
		{"$h=new stdClass();if(curl_exec($h)!==false){return true;}", 0},
		{"$h=curl_init('https://example.invalid');if(curl_exec($h)!==false) echo 'done';", 0},
		{"if(curl_exec($unknown)!==false){return true;}", 0},
		{"$h=curl_init('file:///tmp/data'); if(curl_exec($h)!==false){return true;}", 0},
		{"$h=curl_init('https://example.invalid'); curl_exec($h);", 0},
		{"$h=curl_init('https://example.invalid'); if(curl_exec($h)!==true){return true;}", 0},
		{"$h=curl_init('https://example.invalid'); $ok=curl_exec($h)!==false;", 0},
		{"$h=curl_init('https://example.invalid'); if(curl_exec($h)!==false){echo 'done';}", 0},
		{"$h=curl_init('https://example.invalid'); if(curl_exec($h)!==false){return false;}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlTransportSuccessAsHttpSuccess"}})
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
