package curlwritecallbackmissingbytecount

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlWriteCallbackMissingByteCount"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){return true;});", 0},
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){if($x){return false;}});", 0},
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){if($x){return;}echo \"x\";});", 0},
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){return;});", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestOperationalAdditionalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlWriteCallbackMissingByteCount"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){throw new RuntimeException();});", 0},
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){return 3;return;});", 0},
		{"$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function(){return $unknown;});", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
