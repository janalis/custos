package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewPdoIdentifierProbe dispatches on method calls only, as the rule package does.
type reviewPdoIdentifierProbe struct{ nativeCheckProbe }

func (reviewPdoIdentifierProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func TestReviewPdoIdentifierPlaceholder(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		want      int
	}{
		{"named after from", `SELECT * FROM :t`, 1},
		{"named marker opens the sql", `:t`, 0},
		{"positional after from", `SELECT * FROM ?`, 1},
		{"positional after join", `SELECT * FROM t JOIN ? ON 1 = 1`, 1},
		{"positional after update", `UPDATE ? SET a = 1`, 1},
		{"positional after into", `INSERT INTO ? VALUES (1)`, 1},
		{"named after table", `TRUNCATE TABLE :t`, 1},
		{"lower-case keyword", `delete from ?`, 1},
		{"comment between keyword and marker", "SELECT * FROM /* which */ -- one\n # two\n :t", 1},
		{"subquery from", `SELECT * FROM (SELECT * FROM :t) x`, 1},
		{"subquery inside value-from call", `SELECT TRIM((SELECT a FROM ?))`, 1},
		{"from after closed value-from call", `SELECT EXTRACT(YEAR FROM d) FROM :t`, 1},
		{"second marker in table position", `SELECT :a FROM ?`, 1},
		{"unbalanced closing parenthesis", `SELECT 1) FROM ?`, 1},
		{"is distinct from", `SELECT 1 WHERE a IS DISTINCT FROM :v`, 0},
		{"is not distinct from positional", `SELECT 1 WHERE a IS NOT DISTINCT FROM ?`, 0},
		{"extract", `SELECT EXTRACT(YEAR FROM :d)`, 0},
		{"substring", `SELECT SUBSTRING(name FROM :start FOR :len) FROM people`, 0},
		{"substr positional", `SELECT substr(name FROM ? FOR ?) FROM people`, 0},
		{"mid", `SELECT MID(name FROM ?) FROM people`, 0},
		{"trim", `SELECT TRIM(BOTH 'x' FROM :s)`, 0},
		{"overlay", `SELECT OVERLAY(a PLACING 'x' FROM ? FOR 2)`, 0},
		{"value markers", `SELECT * FROM items WHERE id = :id AND kind = ?`, 0},
		{"marker in parentheses", `INSERT INTO t VALUES (:a, ?)`, 0},
		{"keyword inside a string", `SELECT 'it''s \' FROM' :t`, 0},
		{"quoted identifier before marker", "SELECT * FROM `t` WHERE a = \"FROM\" OR b = :v", 0},
		{"cast is not a marker", `SELECT a::text FROM t WHERE b = :v`, 0},
		{"trailing block comment", `SELECT 1 FROM t /* FROM ? */`, 0},
		{"trailing line comment", `SELECT 1 FROM t -- FROM ?`, 0},
		{"subtraction", `SELECT a - :b FROM t`, 0},
		{"unparsed sql", `SELECT * FROM :t WHERE a = $1`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewPdoIdentifierProbe{nativeCheckProbe{"PdoIdentifierPlaceholder", CheckPdoIdentifierPlaceholder}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"PdoIdentifierPlaceholder"}})
			if err != nil {
				t.Fatal(err)
			}
			quoted := strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(tc.sql)
			src := "<?php function f(PDO $pdo){$pdo->prepare('" + quoted + "');}"
			got := e.Analyze(syntax.Parse("review.php", []byte(src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
