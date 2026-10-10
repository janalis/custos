package unpackbuffertooshort

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"unpack('nvalue','x');", 1},
		{"unpack('Cvalue','x');", 0},
		{"unpack('a2value','x');", 1},
		{"unpack('h3value','x');", 1},
		{"unpack('H2value','x');", 0},
		{"unpack('N*value','x');", 0},
		{"unpack('x2/Nv','xxxx');", 1},
		{"unpack('x2/X1/Cv','xx');", 0},
		{"unpack('@5/Cv','xxx');", 1},
		{"unpack('X2/Cv','xx');", 0},
		{"unpack('@*/Cv','xxx');", 0},
		{"unpack('X*/Cv','xxx');", 0},
		{"unpack('svalue','x');", 0},
		{"unpack('Nvalue/!','x');", 0},
		{"unpack('/Nvalue','x');", 0},
		{"unpack('N99999999999999999999999','x');", 0},
		{"unpack('N1000001','x');", 0},
		{"unpack('Nvalue','xxxx',1);", 1},
		{"unpack('Nvalue','xxxx',-1);", 0},
		{"unpack('Nvalue','xxxx',$offset);", 0},
		{"unpack('Nvalue',$data);", 0},
		{"unpack($format,'x');", 0},
		{"unpack('evalue','x');", 1},
		{"unpack('gvalue','x');", 1},
		{"unpack('Jvalue','x');", 1},
		{"unpack('h*value','x');", 0},
		{"strlen('x');", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnpackBufferTooShort"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edge.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
