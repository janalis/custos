package sqlnullcomparedwithequality

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SqlNullComparedWithEquality"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{{"function quoted(PDO $p){$p->query(\"SELECT 'a\\\\'b'\");$p->query(\"SELECT 'a''b'\");$p->query(\"SELECT 'unterminated\");$p->query(\"SELECT x # = NULL\\n FROM t\");$p->query(\"SELECT x -- = NULL\\n FROM t\");$p->query(\"SELECT /* unterminated\");}", 0}} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
