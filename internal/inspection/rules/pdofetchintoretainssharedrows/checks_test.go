package pdofetchintoretainssharedrows

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoFetchIntoRetainsSharedRows"}})
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
		{"\n$object=new stdClass();$pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = $row; }\n", 1},
		{"\n$object=new stdClass();$pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = clone $row; }\n", 0},
		{"function f(PDOStatement $s){$o=new stdClass();$s->setFetchMode(PDO::FETCH_INTO,$o);while($r=$s->fetch()){$r=new stdClass();$rows[]=$r;}}", 0},
		{"function f(PDOStatement $s){$o=new stdClass();$s->setFetchMode(PDO::FETCH_OBJ);while($r=$s->fetch()){$rows[]=$r;}}", 0},
		{"function f(PDOStatement $s){$o=new stdClass();$s->setFetchMode(PDO::FETCH_INTO,$o);$rows[]=$s->fetch();}", 0},
		{"if(false){\n$pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $object = new stdClass(); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = $row; }\n}", 0},
		{"$rows[]=$x;", 0},
		{"function f(PDOStatement $s){while($r=$s->fetch()){$rows[]=$r;}}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
