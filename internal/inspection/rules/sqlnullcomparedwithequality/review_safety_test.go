package sqlnullcomparedwithequality

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		version      phpversion.Version
		findings     int
	}{
		{"query before 8.0", `function f(PDO $p){$p->query("SELECT * FROM t WHERE x = NULL");}`, phpversion.PHP74, 1},
		{"query on 5.3", `function f(PDO $p){$p->query("SELECT * FROM t WHERE x = NULL");}`, phpversion.PHP53, 1},
		{"query on 8.5", `function f(PDO $p){$p->query("SELECT * FROM t WHERE x = NULL");}`, phpversion.PHP85, 1},
		{"overridden query", `class A extends PDO{function query($s){}} function f(A $p){$p->query("SELECT * FROM t WHERE x = NULL");}`, phpversion.PHP74, 0},
		{"is not null result", `function f(PDO $p){$p->query("SELECT * FROM t WHERE x IS NOT NULL = 1");}`, phpversion.PHP85, 0},
		{"is null result", `function f(PDO $p){$p->query("SELECT * FROM t WHERE x IS NULL = y IS NULL");}`, phpversion.PHP85, 0},
		{"is not null compared with null", `function f(PDO $p){$p->query("SELECT * FROM t WHERE x IS NOT NULL = NULL");}`, phpversion.PHP85, 1},
		{"bare not null", `function f(PDO $p){$p->query("SELECT * FROM t WHERE NOT NULL = x");}`, phpversion.PHP85, 1},
		{"after subquery", `function f(PDO $p){$p->query("SELECT id FROM t WHERE id IN (SELECT id FROM t) AND x = NULL");}`, phpversion.PHP85, 1},
		{"after extract", `function f(PDO $p){$p->query("SELECT id FROM t WHERE EXTRACT(YEAR FROM d) = 1 AND x = NULL");}`, phpversion.PHP85, 1},
		{"after distinct from", `function f(PDO $p){$p->query("SELECT id FROM t WHERE a IS DISTINCT FROM b AND x = NULL");}`, phpversion.PHP85, 1},
		{"subquery select list", `function f(PDO $p){$p->query("SELECT id FROM t WHERE id IN (SELECT a = NULL FROM u)");}`, phpversion.PHP85, 0},
		{"scalar subquery predicate ends", `function f(PDO $p){$p->query("SELECT (SELECT 1 FROM u WHERE a = 1), x = NULL FROM t");}`, phpversion.PHP85, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.version, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.ParseBest("review.php", []byte("<?php "+tc.source), syntax.Options{Version: tc.version})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
