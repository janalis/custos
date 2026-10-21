package untrustedfilesystempath

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedFilesystemPath"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"function openPath($p){readfile($p);} openPath($_GET['file']);", 1},
		{"function openPath($p){readfile($p);} openPath(unknownValidator($_GET['file']));", 0},
		{"rmdir(directory: $_GET['dir']);", 1},
		{"rmdir(context: $c, directory: '/tmp/x');", 0},
		{"function openPath($p){readfile($p);} function outer($p){openPath($p);}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
