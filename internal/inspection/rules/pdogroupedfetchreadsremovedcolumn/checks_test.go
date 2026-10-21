package pdogroupedfetchreadsremovedcolumn

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoGroupedFetchReadsRemovedColumn"}})
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
		{"\n$pdo = new PDO($dsn); $rows = $pdo->query(\"SELECT 'x' AS category, 7 AS value\")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['category'];\n", 1},
		{"\n$pdo = new PDO($dsn); $rows = $pdo->query(\"SELECT 'x' AS category, 7 AS value\")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['value'];\n", 0},
		{"function f(PDO $p){$r=$p->query('SELECT * FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0},
		{"function f(PDO $p){$r=$p->query('SELECT x AS a, y AS a FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['a'];}", 0},
		{"function f(PDO $p){$r=$p->query('SELECT x AS a,y AS b FROM x')->fetchAll(PDO::FETCH_ASSOC);echo $r[0]['a'];}", 0},
		{"function f(PDO $p){$r=$p->query('SELECT x AS a,y AS b FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['a'];}", 1},
		{"if(false){\n$pdo = new PDO($dsn); $rows = $pdo->query(\"SELECT 'x' AS category, 7 AS value\")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['category'];\n}", 0},
		{"echo $r['x'][0]['id'];", 0},
		{"function f(PDO $p){$r=$p->query('SELECT id,value FROM x')->fetchAll(PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0},
		{"function f(PDOStatement $s){$r=$s->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0},
		{"function f(PDO $p){$r=$p->query($sql)->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);echo $r['x'][0]['id'];}", 0},
		{`function f(PDO $p){$r=$p->query('SELECT x AS a,y AS b FROM x')->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);$cb=function()use(&$r){$r=[];};$cb();echo $r['x'][0]['a'];}`, 0}, // state escaped through a reference
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
