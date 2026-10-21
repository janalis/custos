package pdorepeatednamedmarkerwithoutemulation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewPdoRepeatedNamedMarkerWithoutEmulation(t *testing.T) {
	const off = "$p=new PDO('mysql:host=x','u','p',[PDO::ATTR_EMULATE_PREPARES=>false]);"
	const prep = "$p->prepare('SELECT :n+:n');"
	const on = "$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,true);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"constructor proof", off + prep, 1},
		{"inline connection", "(new PDO('mysql:host=x','u','p',[PDO::ATTR_EMULATE_PREPARES=>false]))->prepare('SELECT :n+:n');", 1},
		{"conditional enable", off + "if($f){" + on + "}" + prep, 0},
		{"conditional enable then other call", off + "if($f){" + on + "$p->getAttribute(PDO::ATTR_ERRMODE);}" + prep, 0},
		{"conditional enable inside function", "function f($f){" + off + "if($f){" + on + "}" + prep + "}", 0},
		{"short-circuit enable", off + "$f&&" + on + prep, 0},
		{"conditional direct query", off + "if($f){$p->setAttribute(PDO::MYSQL_ATTR_DIRECT_QUERY,true);}" + prep, 0},
		{"conditional dynamic attribute", off + "if($f){$p->setAttribute($a,true);}" + prep, 0},
		{"unconditional dynamic attribute", off + "$p->setAttribute($a,true);" + prep, 0},
		{"conditional computed constant name", off + "if($f){$p->setAttribute(PDO::{$a},true);}" + prep, 0},
		{"conditional unknown value", off + "if($f){$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,$v);}" + prep, 0},
		{"closure enable", off + "$s=function()use($p){" + on + "};$s();" + prep, 0},
		{"conditional disable keeps proof", off + "if($f){$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);}" + prep, 1},
		{"conditional other attribute", off + "if($f){$p->setAttribute(PDO::ATTR_ERRMODE,2);}" + prep, 1},
		{"conditional other method", off + "if($f){$p->getAttribute(PDO::ATTR_EMULATE_PREPARES);}" + prep, 1},
		{"conditional dynamic method", off + "if($f){$p->$m(PDO::ATTR_EMULATE_PREPARES,true);}" + prep, 0},
		{"arrow function enable", off + "$s=fn()=>$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,true);$s();" + prep, 0},
		{"conditional enable in method", "class C{function f($f){" + off + "if($f){" + on + "}" + prep + "}}", 0},
		{"other connection variable", off + "if($f){$q->setAttribute(PDO::ATTR_EMULATE_PREPARES,true);}" + prep, 1},
		{"property connection", off + "if($f){$this->p->setAttribute(PDO::ATTR_EMULATE_PREPARES,true);}" + prep, 1},
		{"proof restored after conditional", off + "if($f){" + on + "}$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);" + prep, 1},
		{"enable after prepare", off + prep + on, 1},
		{"enable after prepare in loop", off + "foreach($rows as $r){" + prep + "if($r){" + on + "}}", 0},
		{"enable after prepare in while", off + "while($f){" + prep + on + "}", 0},
		{"loop repeats the proof", "foreach($rows as $r){" + off + prep + on + "}", 1},
		{"later loop without prepare", off + prep + "foreach($rows as $r){" + on + "}", 1},
		{"other function", off + prep + "function g($p){" + on + "}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
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
