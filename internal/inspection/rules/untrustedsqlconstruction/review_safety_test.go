package untrustedsqlconstruction

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSinksAndProvenScalars(t *testing.T) {
	deep := "$_GET['id']" + strings.Repeat(")", maxDepth+2)
	deep = strings.Repeat("('a' . ", maxDepth+2) + deep
	for _, tc := range []struct {
		name string
		php  string
		src  string
		want int
	}{
		{"query on 5.3", "5.3", "function f(PDO $p){$p->query('SELECT * FROM t WHERE id='.$_GET['id']);}", 1},
		{"query on 7.4", "7.4", "function f(PDO $p){$p->query('SELECT * FROM t WHERE id='.$_GET['id']);}", 1},
		{"query on 8.0", "8.0", "function f(PDO $p){$p->query('SELECT * FROM t WHERE id='.$_GET['id']);}", 1},
		{"exec on 7.4", "7.4", "function f(PDO $p){$p->exec('DELETE FROM t WHERE id='.$_GET['id']);}", 1},
		{"prepare on 7.4", "7.4", "function f(PDO $p){$p->prepare('SELECT * FROM t WHERE id='.$_GET['id']);}", 1},
		{"overridden query on 7.4", "7.4", "class R extends PDO{function query($s){}} function f(R $p){$p->query('SELECT '.$_GET['id']);}", 0},
		{"inherited query on 7.4", "7.4", "class R extends PDO{} function f(R $p){$p->query('SELECT '.$_GET['id']);}", 1},
		{"placeholder list", "7.4", "function f(PDO $p){$p->prepare('SELECT * FROM t WHERE id IN ('.implode(',', array_fill(0, count($_GET['ids']), '?')).')');}", 0},
		{"count", "7.4", "function f(PDO $p){$p->query('SELECT * FROM t LIMIT '.count($_GET['ids']));}", 0},
		{"strlen", "7.4", "function f(PDO $p){$p->query('SELECT * FROM t WHERE n='.strlen($_GET['q']));}", 0},
		{"product", "7.4", "function f(PDO $p){$p->query('SELECT * FROM t LIMIT 10 OFFSET '.($_GET['page'] * 10));}", 0},
		{"array union", "7.4", "function f(PDO $p){$p->query('SELECT * FROM t WHERE id IN ('.implode(',', $_GET['a'] + $_POST['b']).')');}", 1},
		{"shadowed count", "7.4", "namespace App; function count($v){return $v;} function f(\\PDO $p){$p->query('SELECT * FROM t LIMIT '.count($_GET['ids']));}", 1},
		{"match keywords", "8.0", "function f(PDO $p){$d=match($_GET['s']){'desc'=>'DESC',default=>'ASC'};$p->query('SELECT * FROM t ORDER BY id '.$d);}", 0},
		{"match passthrough", "8.0", "function f(PDO $p){$d=match(true){default=>$_GET['s']};$p->query('SELECT * FROM t ORDER BY id '.$d);}", 1},
		{"walk budget", "8.0", "function f(PDO $p){$p->query(" + deep + ");}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			version, err := phpversion.Parse(tc.php)
			if err != nil {
				t.Fatal(err)
			}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedSqlConstruction"}, PHP: version})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
