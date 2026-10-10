package sqlitenullbindingdiscardsvalue

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f(SQLite3Stmt $s){$s->clear();$s->bindValue(1,'x',SQLITE3_TEXT);$s->bindValue(1,$unknown,SQLITE3_NULL);}", 0},
		{"function f(SQLite3Stmt $s){$s->bindValue(1,true,SQLITE3_NULL);}", 1},
		{"function f(SQLite3Stmt $s){$s->bindValue(1,false,SQLITE3_NULL);}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SqliteNullBindingDiscardsValue"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
