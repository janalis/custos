package iteratorcountchangesrequiredposition

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IteratorCountChangesRequiredPosition"}})
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
		{"\n$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();\n", 1},
		{"\n$it = new ArrayIterator([7, 11]); iterator_count($it); $it->rewind(); echo $it->current();\n", 0},
		{"$it=new ArrayIterator([1,2]);iterator_count($it);echo $it->key();", 1},
		{"$it=new ArrayIterator([1,2]);iterator_count($it);$it->seek(0);echo $it->current();", 0},
		{"$it=new ArrayIterator($items);iterator_count($it);echo $it->current();", 0},
		{"function f(ArrayIterator $it){iterator_count($it);echo $it->current();}", 0},
		{"$it=new ArrayIterator([1,2]);if($b){iterator_count($it);}echo $it->current();", 0},
		{"if(false){\n$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();\n}", 0},
		{"\n$it = new ArrayIterator([7, 11]); iterator_count($it); echo $it->current();\n $broken = ;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
