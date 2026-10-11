package nestedpdotransactionwithoutsavepoint

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NestedPdoTransactionWithoutSavepoint"}})
	if err != nil {
		t.Fatal(err)
	}
	const open = "<?php $p = new PDO('sqlite::memory:'); $p->beginTransaction(); "
	for _, tc := range []struct {
		name, statement string
		findings        int
	}{
		{"direct", "$p->beginTransaction();", 1},
		{"parenthesized", "($p->beginTransaction());", 1},
		{"assigned", "$ok = $p->beginTransaction();", 1},
		{"rolled back first", "$ok = $p->rollBack() && $p->beginTransaction();", 0},
		{"short-circuited", "$p->inTransaction() || $p->beginTransaction();", 0},
		{"ternary", "$p->inTransaction() ? null : $p->beginTransaction();", 0},
		{"argument", "var_dump($p->commit(), $p->beginTransaction());", 0},
		{"condition", "if ($p->commit() && $p->beginTransaction()) {}", 0},
		{"returned", "return $p->commit() && $p->beginTransaction();", 0},
		{"keyed target", "$log[$p->rollBack()] = $p->beginTransaction();", 0},
		{"arrow body", "$later = fn() => $p->beginTransaction();", 0},
		{"negated", "!$p->beginTransaction();", 1},
		{"if condition", "if (!$p->beginTransaction()) { return; }", 1},
		{"returned directly", "return $p->beginTransaction();", 1},
		{"left operand", "$ok = $p->beginTransaction() && $p->commit();", 1},
		{"ternary condition", "$ok = $p->beginTransaction() ? 1 : 2;", 1},
		{"ternary branch", "$ok = $flag ? $p->beginTransaction() : false;", 0},
		{"coalescing assignment", "$ok ??= $p->beginTransaction();", 0},
		{"if body", "if ($flag) $p->beginTransaction();", 0},
		{"elseif condition", "if ($flag) {} elseif ($p->beginTransaction()) {}", 0},
		{"while condition", "while ($p->beginTransaction()) {}", 0},
		{"dynamic target", "${$p->beginTransaction()} = 1;", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.Parse("review.php", []byte(open+tc.statement), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
