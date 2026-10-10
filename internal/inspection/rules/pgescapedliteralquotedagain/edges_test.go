package pgescapedliteralquotedagain

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"$v=pg_escape_literal($c,'x');pg_query(\"SELECT '$v'\");", 1},
		{"$v=pg_escape_literal($c,'x');pg_query($c,'SELECT 3');pg_query($c,\"SELECT '$v' -- note\");", 1},
		{"$v=other($c,'x');pg_query($c,\"SELECT '$v'\");", 0},
		{"$v=pg_escape_literal($c,'x');pg_query($c,\"SELECT 'prefix' || '$v'\");", 1},
		{"$v=pg_escape_literal($c,'x');pg_query($c,\"SELECT \\\"quoted\\\" || '$v'\");", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PgEscapedLiteralQuotedAgain"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
