package pdoboundvalueassumedlive

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewPdoBoundValueAssumedLive(t *testing.T) {
	const open = "function f(PDOStatement $s, $name, $pos){$n=1;"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"no rebind", "$s->bindValue(':n',$n);$n=2;$s->execute();}", 1},
		{"same spelling", "$s->bindValue(':n',$n);$n=2;$s->bindValue(':n',$n);$s->execute();}", 0},
		{"colon then bare", "$s->bindValue(':n',$n);$n=2;$s->bindValue('n',$n);$s->execute();}", 0},
		{"bare then colon", "$s->bindValue('n',$n);$n=2;$s->bindValue(':n',$n);$s->execute();}", 0},
		{"bare rebind by reference", "$s->bindValue(':n',$n);$n=2;$s->bindParam('n',$n);$s->execute();}", 0},
		{"other name", "$s->bindValue(':n',$n);$n=2;$s->bindValue('m',$n);$s->execute();}", 1},
		{"name through local", "$k='n';$s->bindValue(':n',$n);$n=2;$s->bindValue($k,$n);$s->execute();}", 0},
		{"same position", "$s->bindValue(1,$n);$n=2;$s->bindValue(1,$n);$s->execute();}", 0},
		{"other position", "$s->bindValue(1,$n);$n=2;$s->bindValue(2,$n);$s->execute();}", 1},
		{"position against name", "$s->bindValue(1,$n);$n=2;$s->bindValue(':n',$n);$s->execute();}", 1},
		{"name against position", "$s->bindValue(':n',$n);$n=2;$s->bindValue(1,$n);$s->execute();}", 1},
		{"unknown later name", "$s->bindValue(':n',$n);$n=2;$s->bindValue($name,$n);$s->execute();}", 0},
		{"unknown later position", "$s->bindValue(1,$n);$n=2;$s->bindValue($pos,$n);$s->execute();}", 0},
		{"unknown first name", "$s->bindValue($name,$n);$n=2;$s->bindValue(':n',$n);$s->execute();}", 0},
		{"unknown first name without rebind", "$s->bindValue($name,$n);$n=2;$s->execute();}", 1},
		{"rebind before the write", "$s->bindValue(':n',$n);$s->bindValue('n',$n);$n=2;$s->execute();}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzeReview(t, open+tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PdoBoundValueAssumedLive"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
