package curlmultisuccessassumedpertransfer

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlMultiSuccessAssumedPerTransfer"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"if($result===CURLM_OK && !$running){return true;}", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(curl_multi_exec($h,$running)===OTHER && !$running){return true;}", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(curl_multi_exec($h,$running)===CURLM_OK && $running===0){return true;}", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(curl_multi_exec($h,$running)===CURLM_OK && !$other){return true;}", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(curl_multi_exec($h,$running)===CURLM_OK && !$running)return true;", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(curl_multi_exec($h,$running)===CURLM_OK && !$running){echo 1;}", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());if(curl_multi_exec($h,$running)===CURLM_OK && !$running){return false;}", 0},
		{"if(curl_multi_exec($h,$running)===CURLM_OK && !$running){return true;}", 0},
		{"$h=curl_multi_init();curl_multi_add_handle($h,curl_init());$other=new stdClass();$other->read();if(curl_multi_exec($h,$running)===CURLM_OK && !$running){return true;}", 1},
		{"$h=curl_multi_init();if(curl_multi_exec($h,$running)===CURLM_OK && !$running){return true;}", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
