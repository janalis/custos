package hashcontextusedafterfinalization

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HashContextUsedAfterFinalization"}})
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
		{"\n$h = hash_init('sha256'); hash_final($h); hash_update($h, 'next');\n", 1},
		{"\n$h = hash_init('sha256'); hash_update($h, 'next'); hash_final($h);\n", 0},
		{"$h=hash_init('sha256');hash_final($h);hash_copy($h);", 1},
		{"$h=hash_init('sha256');hash_final($h);$h=hash_init('sha256');hash_update($h,'x');", 0},
		{"$h=hash_init('sha256');if($flag){hash_final($h);}hash_update($h,'x');", 0},
		{"$h=hash_init('sha256');hash_final($h);unknown($h);hash_update($h,'x');", 0},
		{"namespace Local; function hash_final($h){} hash_final($h); hash_update($h,'x');", 0},
		{"if(false){\n$h = hash_init('sha256'); hash_final($h); hash_update($h, 'next');\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
