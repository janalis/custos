package sqliteclearwithoutrequiredreset

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f(SQLite3Stmt $s){$s->clear();$s->bindValue(1,'x');$s->execute();}", 0},
		{"function f(SQLite3Stmt $s){$s->execute();$s->clear();$s->reset();$s->bindParam(1,$v);$s->execute();}", 0},
		{"function f(SQLite3Stmt $s){$r=$s->execute();if($r->fetchArray()){$s->clear();$s->bindParam(1,$v);$s->execute();}}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP71, Only: []string{"SqliteClearWithoutRequiredReset"}})
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
