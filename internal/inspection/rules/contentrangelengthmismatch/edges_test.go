package contentrangelengthmismatch

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
		{"header('Content-Range: bytes 0-2/9',true,206);header('Location: /next');echo 'x';", 0},
		{"strlen('x');", 0},
		{"header('Content-Range: nonsense');", 0},
		{"header('Content-Range: bytes 9-1/20',true,206);echo 'a';", 0},
		{"header('Content-Range: bytes 1-9/9',true,206);echo 'a';", 0},
		{"header('Content-Range: bytes 0-1/3');echo 'a';", 0},
		{"header('Content-Range: bytes 0-1/3',true,206);echo 'ab';", 0},
		{"header('Content-Range: bytes 0-1/3',true,206);echo 'a';", 1},
		{"header('Content-Range: bytes 999999999999999999999-999999999999999999999/999999999999999999999',true,206);echo 'a';", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentRangeLengthMismatch"}})
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
