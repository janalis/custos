package curlheaderlistoverwritten

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlHeaderListOverwritten"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{{"$h=curl_init();curl_setopt($h,CURLOPT_HTTPHEADER,[$unknown]);curl_exec($h);", 0}} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
