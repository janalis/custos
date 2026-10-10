package mysqlddlimplicitlycommitstransaction

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqlDdlImplicitlyCommitsTransaction"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"function unknown(PDO $p){$p->rollBack();}", 0},
		{"$p=new PDO(\"mysql:host=x\");$p->beginTransaction();$p->commit();$p->rollBack();", 0},
		{"$p=new PDO(\"mysql:host=x\");$p->beginTransaction();$p->exec($query);$p->rollBack();", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
