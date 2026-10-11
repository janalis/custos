package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewPdoSelectRowCountAssumption(t *testing.T) {
	const decls = "class G{function query($q){return new R();}} class R{function rowCount(){return 0;}} class W extends PDO{function query($q,...$r){return new R();}} "
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"query result", "function f(PDO $p){$s=$p->query('SELECT a FROM t');return $s->rowCount();}", 1},
		{"chained query", "function f(PDO $p){return $p->query('select a from t')->rowCount();}", 1},
		{"case insensitive name", "function f(PDO $p){$s=$p->query('SELECT a FROM t');return $s->ROWCOUNT();}", 1},
		{"prepared", "function f(PDO $p){$s=$p->prepare('SELECT a FROM t');$s->execute();return $s->rowCount();}", 1},
		{"write statement", "function f(PDO $p){$s=$p->query('DELETE FROM t');return $s->rowCount();}", 0},
		{"unknown sql", "function f(PDO $p,$q){$s=$p->query($q);return $s->rowCount();}", 0},
		{"other method", "function f(PDO $p){$s=$p->query('SELECT a FROM t');return $s->columnCount();}", 0},
		{"dynamic method", "function f(PDO $p,$m){$s=$p->query('SELECT a FROM t');return $s->$m();}", 0},
		{"unknown receiver", "function f($s){return $s->rowCount();}", 0},
		{"receiver not a call", "function f(){$s=new R();return $s->rowCount();}", 0},
		{"origin is another method", "function f(PDO $p){$s=$p->errorInfo();return $s->rowCount();}", 0},
		{"origin has a dynamic name", "function f(PDO $p,$m){$s=$p->$m('SELECT a FROM t');return $s->rowCount();}", 0},
		{"untyped connection", "function f($p){$s=$p->query('SELECT a FROM t');return $s->rowCount();}", 0},
		{"user query method", "function f(G $g){$s=$g->query('SELECT a FROM t');return $s->rowCount();}", 0},
		{"overridden query", "function f(W $w){$s=$w->query('SELECT a FROM t');return $s->rowCount();}", 0},
	} {
		for _, target := range []string{"5.3", "7.4", "8.0", "8.5"} {
			t.Run(tc.name+" "+target, func(t *testing.T) {
				version, err := phpversion.Parse(target)
				if err != nil {
					t.Fatal(err)
				}
				probe := expansionCryptoDBProbe{"PdoSelectRowCountAssumption", []syntax.NodeKind{syntax.KMethodCall}, func(ctx *analysis.Context, n syntax.Node) {
					CheckPdoSelectRowCountAssumption(ctx, n, "m")
				}}
				engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.id}, PHP: version})
				if err != nil {
					t.Fatal(err)
				}
				got := engine.Analyze(syntax.Parse("review.php", []byte("<?php "+decls+tc.src), syntax.Options{}))
				if len(got) != tc.want {
					t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
				}
			})
		}
	}
}
