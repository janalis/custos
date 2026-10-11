package imageencodercontenttypemismatch

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
		{"imagepng($im);", 0, ""},
		{"header(\"Content-Type: image/jpeg\");imagepng($im,\"a.png\");", 0, ""},
		{"header(\"Content-Type: image/jpeg\");imagepng($im,null);", 1, ""},
		{"header($type);imagepng($im);", 0, ""},
		{"header(\"Content-Type: image/jpeg\");other();imagepng($im);", 0, ""},
		{"header(\"Content-Type: image/jpeg\");$x=1;imagepng($im);", 0, ""},
		{"if($ok){}imagepng($im);", 0, ""},
		{"header(\"X-Test: a\");imagepng($im);", 0, ""},
		{"header(\"Content-Type: image/jpeg; charset=x\");header(\"X-Test: a\");imagepng($im);", 1, ""},
		{"header(\"Content-Type: text/plain\");imagepng($im);", 0, ""},
		{"header(\"Content-Type: image/jpeg\");header(\"Content-Type: image/png\");imagepng($im);", 0, ""},
		{"header(\"Content-Type: image/png\");imagejpeg($im);", 1, ""},
		{"header('Content-Type: image/x-ms-bmp');imagebmp($im);", 0, ""},
		{"header('Content-Type: image/x-png');imagepng($im);", 0, ""},
		{"header('Content-Type: image/jpg');imagejpeg($im);", 0, ""},
		{"header('Content-Type: image/pjpeg');imagejpeg($im);", 0, ""},
		{"header('Content-Type: image/svg+xml');imagepng($im);", 0, ""},
		{"header('Content-Type: image/jpeg');imagepng(...$args);", 0, ""},
		{"header('Content-Type: image/jpeg');imagepng($im, ...$rest);", 0, ""},
		{"header('Content-Type: image/jpeg');imagepng(...);", 0, ""},
		{"header('Content-Type: IMAGE/BMP');imagegif($im);", 1, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ImageEncoderContentTypeMismatch"}, PHP: phpversion.MustParse(tc.php)}
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
