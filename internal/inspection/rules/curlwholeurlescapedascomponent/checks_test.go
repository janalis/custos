package curlwholeurlescapedascomponent

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlWholeUrlEscapedAsComponent"}})
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
		{"\ncurl_setopt($ch, CURLOPT_URL, curl_escape($ch, 'https://example.net/a'));\n", 1},
		{"\ncurl_setopt($ch, CURLOPT_URL, 'https://example.net/' . curl_escape($ch, 'a b'));\n", 0},
		{"$e=curl_escape($c,'https://example.com/a');curl_setopt($c,CURLOPT_URL,$e);", 1},
		{"curl_setopt($c,CURLOPT_URL,curl_escape($c,'relative/path'));", 0},
		{"curl_setopt($c,CURLOPT_URL,curl_escape($c,$url));", 0},
		{"if(false){\ncurl_setopt($ch, CURLOPT_URL, curl_escape($ch, 'https://example.net/a'));\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
