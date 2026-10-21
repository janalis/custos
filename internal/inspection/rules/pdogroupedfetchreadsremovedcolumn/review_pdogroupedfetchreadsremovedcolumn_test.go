package pdogroupedfetchreadsremovedcolumn

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewPdoGroupedFetchReadsRemovedColumn(t *testing.T) {
	const id = "PdoGroupedFetchReadsRemovedColumn"
	const fetch = "->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);"
	const rows = "$rows=$pdo->query('SELECT category, value FROM t')" + fetch
	const read = "return $rows['x'][0]['category'];}"
	for _, tc := range []struct {
		name, src string
		php       phpversion.Version
		want      int
	}{
		{"plain read", "function f(PDO $pdo){" + rows + read, 0, 1},
		{"top level", "$pdo=new PDO($dsn);" + rows + "echo $rows['x'][0]['category'];", 0, 1},
		{"by-value foreach", "function f(PDO $pdo){" + rows + "foreach($rows as $k=>$g){echo $k;}" + read, 0, 1},
		{"by-value builtin", "function f(PDO $pdo){" + rows + "echo count($rows), implode(',', array_keys($rows['x'][0]));" + read, 0, 1},
		{"nested by-value use", "function f(PDO $pdo){" + rows + "echo strval(count($rows));" + read, 0, 1},
		{"other array by reference", "function f(PDO $pdo, array $o){" + rows + "foreach($o as &$v){$v=1;} sort($o); $a=&$o;" + read, 0, 1},
		{"closure body is another scope", "function f(PDO $pdo){" + rows + "$g=function($rows){sort($rows);};" + read, 0, 1},
		{"later rewrite", "function f(PDO $pdo){" + rows + "$v=$rows['x'][0]['category'];sort($rows);return $v;}", 0, 1},
		{"by-reference foreach", "function f(PDO $pdo){" + rows + "foreach($rows as $k=>&$g){foreach($g as &$r){$r['category']=$k;}}" + read, 0, 0},
		{"by-reference foreach over one group", "function f(PDO $pdo){" + rows + "foreach($rows['x'] as &$r){$r['category']='x';}" + read, 0, 0},
		{"array_walk", "function f(PDO $pdo){" + rows + "array_walk($rows,function(&$g,$k){foreach($g as &$r){$r['category']=$k;}});" + read, 0, 0},
		{"by-reference builtin on a group", "function f(PDO $pdo){" + rows + "array_unshift($rows['x'],['category'=>'x']);" + read, 0, 0},
		{"named by-reference argument", "function f(PDO $pdo){" + rows + "array_walk(callback: 'g', array: $rows);" + read, 0, 0},
		{"reference alias of a group", "function f(PDO $pdo){" + rows + "$g=&$rows['x'];$g[0]['category']='x';" + read, 0, 0},
		{"unknown method receives rows", "function f(PDO $pdo,$o){" + rows + "$o->fill($rows);" + read, 0, 0},
		{"constructor receives rows", "function f(PDO $pdo){" + rows + "new Filler($rows);" + read, 0, 0},
		{"query at 7.4", "function f(PDO $pdo){" + rows + read, phpversion.PHP74, 1},
		{"query at 5.3", "function f(PDO $pdo){" + rows + read, phpversion.PHP53, 1},
		{"query at 8.0", "function f(PDO $pdo){" + rows + read, phpversion.PHP80, 1},
		{"statement variable at 7.4", "function f(PDO $pdo){$s=$pdo->query('SELECT category, value FROM t');$rows=$s" + fetch + read, phpversion.PHP74, 1},
		{"prepare at 7.4", "function f(PDO $pdo){$s=$pdo->prepare('SELECT category, value FROM t');$s->execute();$rows=$s" + fetch + read, phpversion.PHP74, 1},
		{"subclass at 7.4", "class Db extends PDO{} function f(Db $pdo){" + rows + read, phpversion.PHP74, 1},
		{"other column at 7.4", "function f(PDO $pdo){" + rows + "return $rows['x'][0]['value'];}", phpversion.PHP74, 0},
		{"overridden query at 7.4", "class Db extends PDO{function query($q){return new Other;}} function f(Db $pdo){" + rows + read, phpversion.PHP74, 0},
		{"lookalike at 7.4", "class Db{function query($q){return new Other;}} function f(Db $pdo){" + rows + read, phpversion.PHP74, 0},
		{"untyped receiver at 7.4", "function f($pdo){" + rows + read, phpversion.PHP74, 0},
		{"dynamic query name at 7.4", "function f(PDO $pdo,$m){$rows=$pdo->$m('SELECT category, value FROM t')" + fetch + read, phpversion.PHP74, 0},
		{"other method on the result at 7.4", "function f(PDO $pdo){$rows=$pdo->query('SELECT category, value FROM t')->fetch(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);" + read, phpversion.PHP74, 0},
		{"dynamic fetch name at 7.4", "function f(PDO $pdo,$m){$rows=$pdo->query('SELECT category, value FROM t')->$m(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);" + read, phpversion.PHP74, 0},
		{"other PDO method at 7.4", "function f(PDO $pdo){$rows=$pdo->quote('SELECT category, value FROM t')" + fetch + read, phpversion.PHP74, 0},
		{"statement is not a call", "function f(PDOStatement $s){$rows=$s" + fetch + read, 0, 0},
		{"rows are not a call", "function f(array $rows){" + read, 0, 0},
		{"rows are not a variable", "function f(PDO $pdo){return $pdo->query('SELECT category, value FROM t')->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC)['x'][0]['category'];}", 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := expansionCryptoDBProbe{id, []syntax.NodeKind{syntax.KArrayDimFetch}, func(ctx *analysis.Context, n syntax.Node) {
				check(ctx, n, "Read the group column from the group key.")
			}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{id}, PHP: tc.php})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

type expansionCryptoDBProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p expansionCryptoDBProbe) ID() string { return p.id }

func (p expansionCryptoDBProbe) Semantic() {}

func (p expansionCryptoDBProbe) Flow() {}

func (p expansionCryptoDBProbe) Kinds() []syntax.NodeKind { return p.kinds }

func (p expansionCryptoDBProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }
