package filestatcacheafterexternalmutation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FileStatCacheAfterExternalMutation"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$x=1;exec(\"touch /tmp/a\");filesize(\"/tmp/a\");", 0},
		{"echo 1;exec(\"touch /tmp/a\");filesize(\"/tmp/a\");", 0},
		{"$x=filesize(\"/tmp/a\");exec(\"touch /tmp/a; printf done\");filesize(\"/tmp/a\");", 0},
		{"$x=filemtime(\"/tmp/a\");system(\"truncate -s 0 /tmp/a\");filemtime(\"/tmp/a\");", 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
