package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewPdoFetchIntoRetainsSharedRows(t *testing.T) {
	const (
		open  = "function f(PDO $pdo, PDOStatement $t, $x, $list){$o=new stdClass();$rows=[];$s=$pdo->query('SELECT name FROM people');"
		into  = "$s->setFetchMode(PDO::FETCH_INTO,$o);"
		fetch = "while($row=$s->fetch())"
	)
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"shared target", into + fetch + "{$rows[]=$row;}}", 1},
		{"statement parameter", "$t->setFetchMode(PDO::FETCH_INTO,$o);while($row=$t->fetch()){$rows[]=$row;}}", 1},
		{"guarded mode call", "if(!$s->setFetchMode(PDO::FETCH_INTO,$o)){return;}" + fetch + "{$rows[]=$row;}}", 1},
		{"conditional append", into + fetch + "{if($x){$rows[]=$row;}}}", 1},
		{"unbraced body", into + fetch + " $rows[]=$row;}", 1},
		{"statement after the append", into + fetch + "{$rows[]=$row;$x++;}}", 1},
		{"conditional break", into + fetch + "{$rows[]=$row;if($x){break;}}}", 1},
		{"mode change after the loop", into + fetch + "{$rows[]=$row;}$s->setFetchMode(PDO::FETCH_OBJ);}", 1},
		{"mode change on another statement", into + "$t->setFetchMode(PDO::FETCH_OBJ);" + fetch + "{$rows[]=$row;}}", 1},
		{"conditional mode replaced", "if($x){$s->setFetchMode(PDO::FETCH_OBJ);}" + into + fetch + "{$rows[]=$row;}}", 1},
		{"statement per outer iteration", "foreach($list as $q){$s=$pdo->query($q);" + into + fetch + "{$rows[]=$row;}}}", 1},
		{"break leaves an inner loop only", into + fetch + "{foreach($list as $q){$rows[]=$row;break;}}}", 1},
		{"break leaves a switch only", into + fetch + "{switch($x){case 1:$rows[]=$row;break;}}}", 1},
		{"dynamic method name", into + "$s->$x(PDO::FETCH_OBJ);" + fetch + "{$rows[]=$row;}}", 1},

		{"fresh target per iteration", "$s->setFetchMode(PDO::FETCH_INTO,new stdClass());" + fetch + "{$rows[]=$row;$s->setFetchMode(PDO::FETCH_INTO,new stdClass());}}", 0},
		{"conditional mode change", into + "if($x){$s->setFetchMode(PDO::FETCH_OBJ);}" + fetch + "{$rows[]=$row;}}", 0},
		{"mode change in an outer loop", "foreach($list as $q){" + fetch + "{$rows[]=$row;}$s->setFetchMode(PDO::FETCH_INTO,new stdClass());}}", 0},
		{"mode change in outer loop after proof", into + "do{" + fetch + "{$rows[]=$row;}$s->setFetchMode(PDO::FETCH_OBJ);}while($x);}", 0},
		{"mode change in outer for", into + "for(;;){" + fetch + "{$rows[]=$row;}$s->setFetchMode(PDO::FETCH_OBJ);}}", 0},
		{"mode change in outer while", into + "while($x){" + fetch + "{$rows[]=$row;}$s->setFetchMode(PDO::FETCH_OBJ);}}", 0},
		{"break after append", into + fetch + "{$rows[]=$row;break;}}", 0},
		{"return after append", into + fetch + "{$rows[]=$row;return $rows;}}", 0},
		{"throw after append", into + fetch + "{$rows[]=$row;throw new Exception();}}", 0},
		{"exit after append", into + fetch + "{$rows[]=$row;exit;}}", 0},
		{"break after parenthesised append", into + fetch + "{($rows[]=$row);break;}}", 0},
		{"break after guarded append", into + fetch + "{if($x){$rows[]=$row;}break;}}", 0},
		{"return inside an inner loop", into + fetch + "{foreach($list as $q){$rows[]=$row;return $rows;}}}", 0},
		{"break after a switch", into + fetch + "{switch($x){case 1:$rows[]=$row;}break;}}", 0},
		{"no mode call", fetch + "{$rows[]=$row;}}", 0},
		{"per-fetch mode override", into + "while($row=$s->fetch(PDO::FETCH_ASSOC)){$rows[]=$row;}}", 0},
		{"unknown per-fetch mode", into + "while($row=$s->fetch($x)){$rows[]=$row;}}", 0},
	} {
		for _, php := range []string{"5.3", "7.4", "8.0", "8.5"} {
			t.Run(tc.name+"/"+php, func(t *testing.T) {
				if got := reviewPdoFetchInto(t, php, open+tc.src); len(got) != tc.want {
					t.Fatalf("got %+v, want %d", got, tc.want)
				}
			})
		}
	}
}

// Below PHP 8.0 the statement is proven through its PDO::query() origin only.
func TestReviewPdoFetchIntoRetainsSharedRowsLegacyOrigin(t *testing.T) {
	const loop = "$s->setFetchMode(PDO::FETCH_INTO,$o);while($row=$s->fetch()){$rows[]=$row;}}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"query", "function f(PDO $pdo){$o=new stdClass();$s=$pdo->query('SELECT 1');" + loop, 1},
		{"top level", "$o=new stdClass();$pdo=new PDO($dsn);$s=$pdo->query('SELECT 1');$s->setFetchMode(PDO::FETCH_INTO,$o);while($row=$s->fetch()){$rows[]=$row;}", 1},
		{"earlier write replaced", "function f(PDO $pdo,$x){$o=new stdClass();$s=$x;$s=$pdo->query('SELECT 1');" + loop, 1},
		{"unknown origin", "function f($s){$o=new stdClass();" + loop, 0},
		{"unknown factory", "function f($pdo){$o=new stdClass();$s=$pdo->query('SELECT 1');" + loop, 0},
		{"other factory method", "function f(PDO $pdo){$o=new stdClass();$s=$pdo->lastInsertId();" + loop, 0},
		{"not a call", "function f(PDO $pdo,$x){$o=new stdClass();$s=$x;" + loop, 0},
		{"conditional replacement", "function f(PDO $pdo,$x){$o=new stdClass();$s=$pdo->query('SELECT 1');if($x){$s=$x;}" + loop, 0},
		{"reference binding", "function f(PDO $pdo,$x){$o=new stdClass();$s=&$x;" + loop, 0},
		{"compound write", "function f(PDO $pdo,$x){$o=new stdClass();$s=$pdo->query('SELECT 1');$s.=$x;" + loop, 0},
		{"destructured", "function f(PDO $pdo,$x){$o=new stdClass();$s=$pdo->query('SELECT 1');list($s)=$x;" + loop, 0},
		{"captured by reference", "function f(PDO $pdo,$x){$o=new stdClass();$s=$pdo->query('SELECT 1');$g=function()use(&$s){};" + loop, 0},
		{"user override", "class P extends PDO{function query($q){return new Other();}} function f(P $pdo){$o=new stdClass();$s=$pdo->query('SELECT 1');" + loop, 0},
		{"user lookalike", "class Q{function query($q){return new Other();}} function f(Q $pdo){$o=new stdClass();$s=$pdo->query('SELECT 1');" + loop, 0},
		{"dynamic factory", "function f(PDO $pdo,$m){$o=new stdClass();$s=$pdo->$m('SELECT 1');" + loop, 0},
		{"property receiver", "function f(PDO $pdo,$h){$o=new stdClass();$h->s=$pdo->query('SELECT 1');$h->s->setFetchMode(PDO::FETCH_INTO,$o);while($row=$h->s->fetch()){$rows[]=$row;}}", 0},
		{"other method", "function f(PDO $pdo){$o=new stdClass();$s=$pdo->query('SELECT 1');$s->setFetchMode(PDO::FETCH_INTO,$o);while($row=$s->fetchObject()){$rows[]=$row;}}", 0},
		{"dynamic method", "function f(PDO $pdo,$m){$o=new stdClass();$s=$pdo->query('SELECT 1');$s->setFetchMode(PDO::FETCH_INTO,$o);while($row=$s->$m()){$rows[]=$row;}}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := reviewPdoFetchInto(t, "7.4", tc.src); len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func reviewPdoFetchInto(t *testing.T, php, src string) []int {
	t.Helper()
	const id = "PdoFetchIntoRetainsSharedRows"
	// Dispatch on assignments only, as the rule package does.
	check := func(ctx *analysis.Context, n syntax.Node, message string) {
		if _, ok := n.(*syntax.Assign); ok {
			CheckPdoFetchIntoRetainsSharedRows(ctx, n, message)
		}
	}
	e, err := analysis.NewEngine([]analysis.Rule{nativeCheckProbe{id, check}}, analysis.Config{Only: []string{id}, PHP: phpversion.MustParse(php)})
	if err != nil {
		t.Fatal(err)
	}
	var starts []int
	for _, finding := range e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{})) {
		starts = append(starts, int(finding.Span.Start))
	}
	return starts
}
