package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewPdoPlaceholderProbe dispatches on method calls only, as the rule package does.
type reviewPdoPlaceholderProbe struct{ nativeCheckProbe }

func (reviewPdoPlaceholderProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func TestReviewPdoPlaceholderBindingMismatchCapturedStatement(t *testing.T) {
	const open = "function f(PDO $pdo){$s=$pdo->prepare('SELECT :a,:b');"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"closure binds the other marker", open + "$b=function() use ($s){$s->bindValue(':b',2);};$s->bindValue(':a',1);$b();$s->execute();}", 0},
		{"by-reference capture", open + "$b=function() use (&$s){$s->bindValue(':b',2);};$s->bindValue(':a',1);$b();$s->execute();}", 0},
		{"arrow function binds the other marker", open + "$s->bindValue(':a',1);array_map(fn($v)=>$s->bindValue(':b',$v),[2]);$s->execute();}", 0},
		{"nested arrow function", open + "$s->bindValue(':a',1);$g=fn()=>fn($v)=>$s->bindValue(':b',$v);$s->execute();}", 0},
		{"call inside a loop", open + "$s->bindValue(':a',1);while(more()){$s->execute();$b=fn()=>$s->bindValue(':b',2);}}", 0},
		{"top-level capture", "$pdo=new PDO('sqlite::memory:');$s=$pdo->prepare('SELECT :a,:b');$b=function() use ($s){$s->bindValue(':b',2);};$s->bindValue(':a',1);$b();$s->execute();", 0},
		{"closure capturing another variable", open + "$b=function() use ($pdo){};$s->bindValue(':a',1);$s->execute();}", 1},
		{"arrow function using another variable", open + "$b=fn($v)=>$pdo->quote($v);$s->bindValue(':a',1);$s->execute();}", 1},
		{"closure created after the call", open + "$s->bindValue(':a',1);$s->execute();return function() use ($s){return $s->fetch();};}", 1},
		{"arrow function created after the call", open + "$s->bindValue(':a',1);$s->execute();return fn()=>$s->fetch();}", 1},
		{"capture with explicit parameters", open + "$b=function() use ($s){$s->bindValue(':b',2);};$b();$s->execute(['a'=>1]);}", 1},
		{"capture with complete parameters", open + "$b=function() use ($s){};$s->execute(['a'=>1,'b'=>2]);}", 0},
		{"call inside a capturing closure", "function f(PDO $pdo){return function() use ($pdo){$s=$pdo->prepare('SELECT :a,:b');$s->bindValue(':a',1);$s->execute();};}", 1},
		{"top-level mismatch", "$pdo=new PDO('sqlite::memory:');$s=$pdo->prepare('SELECT :a,:b');$s->bindValue(':a',1);$s->execute();", 1},
		{"chained execute", "function f(PDO $pdo){$pdo->prepare('SELECT :a')->execute();}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewPdoPlaceholderProbe{nativeCheckProbe{"PdoPlaceholderBindingMismatch", CheckPdoPlaceholderBindingMismatch}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"PdoPlaceholderBindingMismatch"}})
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
