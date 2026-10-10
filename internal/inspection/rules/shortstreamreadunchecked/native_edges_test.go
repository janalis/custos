package shortstreamreadunchecked

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
		{"unpack('nsize',fread($h,2));", 1},
		{"unpack('Nsize',fread($h,$length));", 0},
		{"unpack('Nsize',fread($h,0));", 0},
		{"unpack($format,fread($h,4));", 0},
		{"unpack('Nsize',$unknown);", 0},
		{"unpack('Nsize',file_get_contents($path));", 0},
		{"unpack('Nsize',fread($h,8));", 0},
		{"unpack('Cvalue',fread($h,4));", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ShortStreamReadUnchecked"}})
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
