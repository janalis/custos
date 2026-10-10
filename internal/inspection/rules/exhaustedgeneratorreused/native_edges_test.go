package exhaustedgeneratorreused

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
		{"function values(){yield 1;} function consume(){foreach(values() as $v){}}", 0},
		{"function values(){return [1];} $g=values(); foreach($g as $v){}", 0},
		{"function values(){yield sendPacket();} $g=values(); foreach($g as $v){} foreach($g as $v){}", 0},
		{"function values(){if(true){yield 1;}} $g=values(); foreach($g as $v){} foreach($g as $v){}", 0},
		{"function values(){yield 1;} $g=values(); foreach($g as $v){} foreach($g as $v){}", 1},
		{"foreach($unknown as $v){}", 0},
		{"iterator_to_array($unknown);", 0},
		{"function values(){yield 1;} $g=values(); foreach($g as $v){break;} foreach($g as $v){}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ExhaustedGeneratorReused"}})
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
