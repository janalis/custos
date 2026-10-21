package sqlitenondeterministicfunctiondeclareddeterministic

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SqliteNondeterministicFunctionDeclaredDeterministic"}})
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
		{"\n$db = new SQLite3(':memory:'); $db->createFunction('dice', fn() => random_int(1, 6), 0, SQLITE3_DETERMINISTIC);\n", 1},
		{"\n$db = new SQLite3(':memory:'); $db->createFunction('twice', fn($n) => $n * 2, 1, SQLITE3_DETERMINISTIC);\n", 0},
		{"function f(SQLite3 $d){$d->createFunction('x',fn()=>time()+1,0,SQLITE3_DETERMINISTIC);}", 0},
		{"function f(SQLite3 $d){$d->createFunction('x',fn()=>time(),0,0);}", 0},
		{"function f(SQLite3 $d){$d->createFunction('x',$callback,0,SQLITE3_DETERMINISTIC);}", 0},
		{"if(false){\n$db = new SQLite3(':memory:'); $db->createFunction('dice', fn() => random_int(1, 6), 0, SQLITE3_DETERMINISTIC);\n}", 0},
		{"$x->createFunction('x',fn()=>time(),0,SQLITE3_DETERMINISTIC);", 0},
		{"function f(SQLite3 $d){$d->createFunction('x',fn()=>fn()=>time(),0,SQLITE3_DETERMINISTIC);}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
