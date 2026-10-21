package pdoplaceholderbindingmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoPlaceholderBindingMismatch"}})
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
		{"function f(PDO $p,$flag){$s=$p->prepare('SELECT :id');$flag && $s->bindValue(':id',1);$s->execute();}", 0},
		{"function f(PDO $p,$flag){$s=$p->prepare('SELECT :id');$flag ? $s->bindValue(':id',1) : 0;$s->execute();}", 0},
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"function f(PDO $p){$s=$p->prepare('SELECT :a,:b');$s->execute(['a'=>1]);}", 1},
		{"function f(PDO $p){$s=$p->prepare('SELECT ?');$s->bindValue(1,7);$s->execute();}", 0},
		{"function f(PDO $p,$name){$s=$p->prepare('SELECT :a');$s->bindValue($name,7);$s->execute();}", 0},
		{"function f(PDO $p){$s=$p->prepare('SELECT :a');retain($s);$s->execute();}", 0},
		{"function f(PDOStatement $s){$s->execute(['a'=>1]);}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
