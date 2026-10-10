package encodingconversionargumentsreversed

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
		{"$s=mb_convert_encoding(\"été\",\"ISO-8859-1\",\"UTF-8\");echo json_encode($s);", 1, ""},
		{"$s=mb_convert_encoding(\"abc\",\"ISO-8859-1\",\"UTF-8\");echo json_encode($s);", 0, ""},
		{"$s=mb_convert_encoding(\"😀\",\"ISO-8859-1\",\"UTF-8\");echo json_encode($s);", 0, ""},
		{"$s=mb_convert_encoding($input,\"ISO-8859-1\",\"UTF-8\");echo json_encode($s);", 0, ""},
		{"$s=mb_convert_encoding(\"été\",\"UTF-8\",\"UTF-8\");echo json_encode($s);", 0, ""},
		{"$s=mb_convert_encoding(\"été\",\"ISO-8859-1\",\"ISO-8859-1\");echo json_encode($s);", 0, ""},
		{"$s=mb_convert_encoding(\"été\",$to,\"UTF-8\");echo json_encode($s);", 0, ""},
		{"$s=strtolower(\"été\",\"ISO-8859-1\",\"UTF-8\");echo json_encode($s);", 0, ""},
		{"echo json_encode($data);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"EncodingConversionArgumentsReversed"}, PHP: phpversion.MustParse(tc.php)}
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
