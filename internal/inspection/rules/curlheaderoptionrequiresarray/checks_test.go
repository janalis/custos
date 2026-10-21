package curlheaderoptionrequiresarray

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlHeaderOptionRequiresArray"}})
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
		{"\ncurl_setopt($ch, CURLOPT_HTTPHEADER, 'Accept: text/plain');\n", 1},
		{"\ncurl_setopt($ch, CURLOPT_HTTPHEADER, ['Accept: text/plain']);\n", 0},
		{"curl_setopt_array($ch,[CURLOPT_HTTPHEADER=>'Accept: x']);", 1},
		{"curl_setopt_array($ch,[CURLOPT_HTTPHEADER=>['Accept: x']]);", 0},
		{"curl_setopt_array($ch,$options);", 0},
		{"curl_setopt_array($ch,[CURLOPT_TIMEOUT=>1]);", 0},
		{"curl_setopt($ch,$option,'x');", 0},
		{"curl_setopt($ch,CURLOPT_HTTPHEADER,$unknown);", 0},
		{"if(false){\ncurl_setopt($ch, CURLOPT_HTTPHEADER, 'Accept: text/plain');\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
