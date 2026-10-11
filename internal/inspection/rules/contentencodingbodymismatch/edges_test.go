package contentencodingbodymismatch

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"ob_start('transform');header('Content-Encoding: gzip');echo gzcompress('x');", 0},
		{"header('X: y');header('Content-Encoding: gzip');echo gzcompress('x');", 1},
		{"echo 'x'; echo gzcompress('x');", 0},
		{"strlen('x');", 0},
		{"header('Content-Encoding: gzip');gzcompress('x');", 0},
		{"header('Content-Encoding: gzip');echo gzencode('x');", 0},
		{"header('Content-Encoding: gzip');echo gzcompress('x');", 1},
		{"header('Content-Encoding: identity');print gzdeflate('x');", 1},
		{"header('Content-Encoding: gzip');echo gzencode('x',-1,15);", 1},
		{"header('Content-Encoding: gzip');echo gzencode('x',-1,$unknown);", 0},
		{"header('Content-Encoding: gzip');echo gzencode('x',-1,99);", 0},
		{"header('Content-Encoding: gzip');echo gzcompress('x',-1,31);", 0},
		{"header('Content-Encoding: gzip');echo gzcompress('x',-1,-15);", 1},
		{"header('Content-Encoding: gzip');echo gzcompress('x',-1,15);", 1},
		{"header('Content-Encoding: gzip');echo gzcompress('x',-1,$unknown);", 0},
		{"header('Content-Encoding: gzip');echo gzcompress('x',-1,99);", 0},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,ZLIB_ENCODING_GZIP);", 0},
		{"header('Content-Encoding: deflate');echo gzdeflate('x',9,ZLIB_ENCODING_DEFLATE);", 0},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',encoding: 31);", 0},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,ZLIB_ENCODING_DEFLATE);", 1},
		{"header('Content-Encoding: deflate');echo gzdeflate('x',9,ZLIB_ENCODING_GZIP);", 1},
		{"header('Content-Encoding: gzip');echo gzdeflate('x');", 1},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,ZLIB_ENCODING_RAW);", 1},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,$unknown);", 0},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,99);", 0},
		{"header('Content-Encoding: gzip');echo gzencode('x',-1,ZLIB_ENCODING_RAW);", 1},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,UNDECLARED_FRAMING);", 0},
		{"header('Content-Encoding: gzip');echo gzdeflate('x',9,PHP_EOL);", 0},
		{"header('Content-Encoding: deflate');echo gzcompress('x',9,(ZLIB_ENCODING_RAW));", 1},
		{"echo gzcompress('x');", 0},
		{"header('Content-Encoding: gzip');$x=gzcompress('x');", 0},
		{"header('Content-Encoding: gzip');echo 'a'.gzcompress('x');", 0},
		{"header('Content-Encoding: gzip');unknown();echo gzcompress('x');", 0},
		{"$x=1;echo gzcompress('x');", 0},
		{"header('X: y');echo gzcompress('x');", 0},
		{"header($unknown);echo gzcompress('x');", 0},
		{"header('Content-Encoding: gzip');echo (gzcompress('x'));", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentEncodingBodyMismatch"}})
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

func TestStatementBudget(t *testing.T) {
	source := "<?php " + strings.Repeat("header('X: y');", 4097) + "header('Content-Encoding: gzip');echo gzcompress('x');"
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentEncodingBodyMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Analyze(syntax.Parse("bounded.php", []byte(source), syntax.Options{})); len(got) != 0 {
		t.Fatalf("budget exceeded: %+v", got)
	}
}
