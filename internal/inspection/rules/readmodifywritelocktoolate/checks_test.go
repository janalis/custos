package readmodifywritelocktoolate

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ReadModifyWriteLockTooLate"}})
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
		{"file_put_contents(nextPath(),file_get_contents(nextPath()).'x',LOCK_EX);", 0},
		{"file_put_contents('file',file_get_contents('file').'x',LOCK_EX);", 1},
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"file_put_contents('/tmp/f',function(){return 'x';},LOCK_EX);", 0},
		{"file_put_contents('/tmp/f','x',unknownFlags());file_put_contents('/tmp/f');", 0},
		{"file_put_contents('/tmp/f','x',LOCK_EX);", 0},
		{"file_put_contents(filename:'file',flags:LOCK_EX);", 0},
		{"file_put_contents('file',['entry'=>'value'],LOCK_EX);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
