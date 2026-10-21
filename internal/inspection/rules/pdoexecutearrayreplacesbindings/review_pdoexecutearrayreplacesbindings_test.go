package pdoexecutearrayreplacesbindings

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewPdoExecuteArrayReplacesBindings(t *testing.T) {
	const open = "function f(PDO $pdo){$s=$pdo->prepare("
	for _, tc := range []struct {
		name, src string
		want      int
		fixable   bool
	}{
		{"named marker bound", open + "'SELECT :id');$s->bindValue(':id',7);$s->execute([]);}", 1, true},
		{"marker bound without colon", open + "'SELECT :id');$s->bindValue('id',7);$s->execute([]);}", 1, true},
		{"positional marker bound", open + "'SELECT ?');$s->bindValue(1,7);$s->execute([]);}", 1, true},
		{"bindParam on marker", open + "'SELECT :id');$s->bindParam(':id',$v);$s->execute([]);}", 1, true},
		{"one of two markers bound", open + "'SELECT :a,:b');$s->bindValue(':a',7);$s->execute([]);}", 1, true},
		{"comment in argument list withholds fix", open + "'SELECT :id');$s->bindValue(':id',7);$s->execute(/* none */[]);}", 1, false},
		{"sql without markers", open + "'SELECT 1 AS v');$s->bindValue(':id',7);$s->execute([]);}", 0, false},
		{"binding names absent marker", open + "'SELECT :id');$s->bindValue(':other',7);$s->execute([]);}", 0, false},
		{"one binding names absent marker", open + "'SELECT :id');$s->bindValue(':id',7);$s->bindValue(':other',7);$s->execute([]);}", 0, false},
		{"positional binding out of range", open + "'SELECT ?');$s->bindValue(2,7);$s->execute([]);}", 0, false},
		{"zero position", open + "'SELECT ?');$s->bindValue(0,7);$s->execute([]);}", 0, false},
		{"unknown binding name", open + "'SELECT :id');$s->bindValue($name,7);$s->execute([]);}", 0, false},
		{"same-named function is not a binding", open + "'SELECT :id');bindValue($s,':id');$s->execute([]);}", 0, false},
		{"no prior binding", open + "'SELECT :id');$s->execute([]);}", 0, false},
		{"non-empty array", open + "'SELECT :id');$s->bindValue(':id',7);$s->execute([':id'=>1]);}", 0, false},
		{"no array", open + "'SELECT :id');$s->bindValue(':id',7);$s->execute();}", 0, false},
		{"unknown sql", open + "$sql);$s->bindValue(':id',7);$s->execute([]);}", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := New()
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"PdoExecuteArrayReplacesBindings"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
			if tc.want == 1 && (len(got[0].Fixes) > 0) != tc.fixable {
				t.Fatalf("fixable = %v, want %v", len(got[0].Fixes) > 0, tc.fixable)
			}
		})
	}
}
