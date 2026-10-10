package jsonintegerprecisionloss

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
		{"json_decode($unknown);", 0},
		{"json_decode('invalid');", 0},
		{"json_decode('{\"n\":1.25}');", 0},
		{"json_decode('[1,true,null,\"text\"]');", 0},
		{"json_decode('[9223372036854775808]');", 1},
		{"json_decode('[9223372036854775808]',false,512,0);", 1},
		{"json_decode(json:'[9223372036854775808]');", 1},
		{"json_decode('[9223372036854775808]',false,512,$flags);", 0},
		{"json_decode('[9223372036854775808]',false);", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonIntegerPrecisionLoss"}})
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
