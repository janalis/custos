package curlheadercallbackmissingbytecount

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlHeaderCallbackMissingByteCount"}})
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
		{"\ncurl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { echo $line; });\n", 1},
		{"\ncurl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { return strlen($line); });\n", 0},
		{"curl_setopt($c,CURLOPT_HEADERFUNCTION,function($c,$line){return;});", 1},
		{"curl_setopt($c,CURLOPT_HEADERFUNCTION,fn($c,$line)=>strlen($line));", 0},
		{"curl_setopt($c,CURLOPT_HEADERFUNCTION,function(){yield 1;});", 0},
		{"if(false){\ncurl_setopt($ch, CURLOPT_HEADERFUNCTION, function($ch, $line) { echo $line; });\n}", 0},
		{"curl_setopt($c,CURLOPT_HEADERFUNCTION,function($c,$l){$f=function(){};return strlen($l);});", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
