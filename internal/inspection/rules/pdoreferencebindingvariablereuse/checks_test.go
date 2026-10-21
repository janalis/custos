package pdoreferencebindingvariablereuse

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoReferenceBindingVariableReuse"}})
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
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"function f(PDO $p,$sql){$s=$p->prepare($sql);$v=1;$s->bindParam(':a',$v);$v=2;$s->bindParam(':b',$v);$s->execute();}", 0},
		{"function f(PDO $p,$args){$s=$p->prepare('SELECT :a,:b');$v=1;$s->bindParam(':a',$v);$v=2;$s->bindParam(':b',$v);$s->execute($args);}", 0},
		{"function f(PDO $p,$name){$s=$p->prepare('SELECT :a,:b');$v=1;$s->bindParam($name,$v);$s->execute();}", 0},
		{"function f(PDO $p,$v){$s=$p->prepare('SELECT :a,:b');$s->bindParam(':a',$v);$s->execute();}", 0},
		{"function f(PDO $p){$s=$p->prepare('SELECT :a,:b');$v='a';$s->bindParam(':a',$v);$v='b';$s->bindParam(':b',$v);$s->execute();}", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
