package pdorepeatednamedmarkerwithoutemulation

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoRepeatedNamedMarkerWithoutEmulation"}})
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
		{"\n$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :n + :n');\n", 1},
		{"\n$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :a + :b');\n", 0},
		{"$p=new PDO('sqlite::memory:');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT :n+:n');", 0},
		{"$p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,true);$p->prepare('SELECT :n+:n');", 0},
		{"$p=new PDO('mysql:host=x');$p->prepare('SELECT :n+:n');", 0},
		{"$p=new PDO('mysql:host=x',null,null,[PDO::ATTR_EMULATE_PREPARES=>false]);$p->prepare('SELECT :n+:n');", 1},
		{"$p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare(\"SELECT ':n', :n\");", 0},
		{"$p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT :n+:N');", 0},
		{"if(false){\n$pdo = new PDO('mysql:host=localhost;dbname=sample'); $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false); $pdo->prepare('SELECT :n + :n');\n}", 0},
		{"function f(PDO $p){$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT :x+:x');}", 0},
		{"$p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare($sql);", 0},
		{"$p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);$p->prepare('SELECT $body$');", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
