package pdoboundvalueassumedlive

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoBoundValueAssumedLive"}})
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
		{"\n$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindValue(':n', $n); $n = 11; $s->execute();\n", 1},
		{"\n$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindParam(':n', $n); $n = 11; $s->execute();\n", 0},
		{"function f(PDOStatement $s){$n=1;$s->bindValue(':n',$n);$n=2;$s->bindValue(':n',$n);$s->execute();}", 0},
		{"function f(PDOStatement $s){$n=1;$s->bindValue(':n',$n);$n=2;$s->execute(['n'=>2]);}", 0},
		{"function f(PDOStatement $s){$s->bindValue(':n',1);$s->execute();}", 0},
		{"function f(PDOStatement $s){$n=new stdClass();$s->bindValue(':n',$n);$n=new stdClass();$s->execute();}", 0},
		{"if(false){\n$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindValue(':n', $n); $n = 11; $s->execute();\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
