package curlreadcallbackexceedsrequestedsize

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlReadCallbackExceedsRequestedSize"}})
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
		{"\ncurl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit + 2));\n", 1},
		{"\ncurl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit));\n", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,fn($c)=>'x');", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,function($c,$h,$n){});", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>'x');", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>str_repeat('xx',$n+1));", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>str_repeat('x',$n-1));", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,fn($c,$h,$n)=>str_repeat('x',$unknown+1));", 0},
		{"curl_setopt($c,CURLOPT_READFUNCTION,function($c,$h,$n){$n=1;return str_repeat('x',$n+1);});", 0},
		{"if(false){\ncurl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit + 2));\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
