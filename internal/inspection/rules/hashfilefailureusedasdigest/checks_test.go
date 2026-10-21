package hashfilefailureusedasdigest

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HashFileFailureUsedAsDigest"}})
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
		{"\nfile_put_contents($manifest, hash_file('sha256', $path));\n", 1},
		{"\n$h = hash_file('sha256', $path); if ($h !== false) { file_put_contents($manifest, $h); }\n", 0},
		{"$h=hash_file('sha256',$p);if($h===false){throw new RuntimeException();}file_put_contents($p,$h);", 0},
		{"$h=hash_file('sha256',$p);if($h==='abc'){return true;}", 1},
		{"return hash_file('sha256',$p);", 0},
		{"file_put_contents($p,'x');", 0},
		{"$h=hash_file('sha256',$p); if('abc'!==$h){return false;}", 1},
		{"if(false){\nfile_put_contents($manifest, hash_file('sha256', $path));\n}", 0},
		{"hash_file('sha256',$p)=='abc';", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
