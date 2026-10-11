package loopclosurecapturesreference

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewCounterWritesAreNotReads pins which occurrences of the captured
// counter count as reads inside the deferred closure.
func TestReviewCounterWritesAreNotReads(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want int
	}{
		{"plain read", "return $i;", 1},
		{"plain write", "$i=10;", 0},
		{"compound write", "$i+=1;", 1},
		{"assignment source", "$x=$i;", 1},
		{"short destructuring", "[$i]=[10];", 0},
		{"list destructuring", "list($i)=[10];", 0},
		{"nested destructuring", "[[$i],$x]=[[10],1];", 0},
		{"keyed destructuring value", "['a'=>$i]=['a'=>10];", 0},
		{"destructuring key", "[$i=>$x]=[1,2,3];", 1},
		{"array value element", "$x=[$i];", 1},
		{"array passed to call", "foo([$i]);", 1},
		{"spread element", "$x=[...$i];", 1},
		{"foreach value", "foreach([7] as $i){}", 0},
		{"foreach key", "foreach([7] as $i=>$v){}", 0},
		{"foreach destructuring", "foreach([[7]] as [$i]){}", 0},
		{"foreach destructuring key", "foreach([[7]] as [$i=>$v]){}", 1},
		{"foreach subject", "foreach($i as $v){}", 1},
		{"foreach body", "foreach([7] as $v){echo $i;}", 1},
		{"unset", "unset($i);", 0},
		{"write then read", "[$i]=[10]; return $i;", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := "<?php $jobs=[]; for($i=0;$i<3;$i++){ $jobs[]=function() use (&$i){" + tc.body + "}; } foreach($jobs as $job){echo $job();}"
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"LoopClosureCapturesReference"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
			if len(got) != tc.want {
				t.Errorf("%s: got %d findings, want %d", tc.body, len(got), tc.want)
			}
		})
	}
}
