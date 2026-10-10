package nestedpdotransactionwithoutsavepoint

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NestedPdoTransactionWithoutSavepoint"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"function empty(PDO $p){$p->beginTransaction();}", 0},
		{"function echoBefore(PDO $p){echo 1;$p->beginTransaction();}", 0},
		{"$p=new PDO(\"sqlite::memory:\");$p->beginTransaction();$p->exec(\"COMMIT\");$p->beginTransaction();", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
