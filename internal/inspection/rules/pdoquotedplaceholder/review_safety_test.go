package pdoquotedplaceholder

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// The rule reports without offering an edit: unquoting is not provably valid SQL.
func TestReviewReportsWithoutFix(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
	}{
		{"value position", `function f(PDO $p){$s=$p->prepare("SELECT * FROM items WHERE label = ':label'");$s->execute(['label'=>'new']);}`},
		{"column alias", `function f(PDO $p){$s=$p->prepare("SELECT 1 AS ':label'");$s->execute(['label'=>'new']);}`},
		{"adjacent literal", `function f(PDO $p){$s=$p->prepare("SELECT x'00' ':label'");$s->execute(['label'=>'new']);}`},
		{"shared statement", `function f(PDO $p){$s=$p->prepare("SELECT * FROM items WHERE label = ':label'");$s->execute(['label'=>'a']);$s->execute(['label'=>'b']);}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoQuotedPlaceholder"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) == 0 {
				t.Fatal("expected a finding")
			}
			for _, g := range got {
				if len(g.Fixes) != 0 {
					t.Fatalf("unexpected fix on %+v", g)
				}
			}
		})
	}
}
