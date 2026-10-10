package imagedecodefailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
		php    string
	}{
		{"other(); $x->other(); echo $x[0];", 0, ""},
		{"imagepng($other);", 0, ""},
		{"imagepng(imagecreatefromstring($bytes));", 1, ""},
		{"$im=imagecreatefromstring(\"literal bytes\");imagepng($im);", 0, ""},
		{"$im=imagecreatefromstring($bytes);if($im!==false){imagepng($im);}", 0, ""},
		{"$im=imagecreatefromstring($bytes);$im=new stdClass();imagepng($im);", 0, ""},
		{"$im=imagecreatefromstring($bytes);echo imagesx($im);", 1, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ImageDecodeFailureUnchecked"}, PHP: phpversion.MustParse(tc.php)}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, cfg)
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
