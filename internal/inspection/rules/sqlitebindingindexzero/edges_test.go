package sqlitebindingindexzero

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
		{"function f(SQLite3Stmt $s){$s->clear();$s->bindValue($x,'x');$s->bindValue(':x','x');}", 0},
		{"function f(SQLite3Stmt $s){$s->bindParam(0,$v);}", 1},
		// A closure holding a reference to the position may change it.
		{"function f(SQLite3Stmt $s){$i=0;$g=function() use (&$i){$i=1;};$s->bindValue($i,'x');}", 0},
		{"function f(SQLite3Stmt $s,$c){$i=0;while($c--){$s->bindValue($i,'x');$g=function() use (&$i){$i=1;};}}", 0},
		{"function f(SQLite3Stmt $s){$i=0;$s->bindValue($i,'x');$g=function() use (&$i){$i=1;};}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SqliteBindingIndexZero"}})
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
