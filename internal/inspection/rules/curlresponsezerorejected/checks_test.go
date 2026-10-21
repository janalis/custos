package curlresponsezerorejected

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlResponseZeroRejected"}})
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
		{"\ncurl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (!curl_exec($ch)) { throw new RuntimeException(); }\n", 1},
		{"\ncurl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (curl_exec($ch) === false) { throw new RuntimeException(); }\n", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(curl_exec($c)==false){return false;}", 1},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,false);if(!curl_exec($c)){return false;}", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(!curl_exec($c)){echo 'empty';}", 0},
		{"if(!curl_exec($c)){return false;}", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);unknown($c);if(!curl_exec($c)){return false;}", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(curl_exec($c)===''){return false;}", 0},
		{"if(false){\ncurl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (!curl_exec($ch)) { throw new RuntimeException(); }\n}", 0},
		{"if($f){curl_setopt($c,CURLOPT_RETURNTRANSFER,true);}if(!curl_exec($c)){return false;}", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(!(curl_exec($c))){return false;}", 1},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(+curl_exec($c)){return false;}", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(curl_exec($c)){return false;}", 0},
		{"curl_setopt($c,CURLOPT_RETURNTRANSFER,true);if(!curl_exec($c)){$f=function(){return false;};}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
