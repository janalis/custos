package onlywritesonparameter

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// A read placed after a do-while counts unless the loop certainly never
// falls through. syntax.Terminates decides the literal jumps; the rule's own
// do-while check keeps the read reachable where Terminates still answers
// "never falls through" for a statement holding a do-while with a jump.
func TestReadAfterDoWhile(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want int
	}{
		// Decided by syntax.Terminates: the loop falls through.
		{"break", "do{if($ok){break;}return [];}while(false);", 0},
		{"continue", "do{if($ok){continue;}return [];}while(false);", 0},
		{"continue 2 from nested loop", "do{while($ok){continue 2;}return [];}while(false);", 0},
		{"break 2 from nested loop", "do{while($ok){break 2;}return [];}while(false);", 0},
		// Decided by the rule: a computed level may target the do-while, whose
		// condition then ends it.
		{"continue with computed level", "do{if($ok){continue $n;}return [];}while(false);", 0},
		// Deliberately cautious: any jump inside the do-while keeps what
		// follows reachable.
		{"continue in endless do-while", "do{if($ok){continue;}}while(true);", 0},
		{"break of a nested loop", "do{foreach($n as $x){break;}return [];}while(false);", 0},
		{"do-while inside a terminating block", "{do{if($ok){break;}}while(false);return [];}", 0},
		// No jump: the read is unreachable and the write is lost.
		{"endless do-while", "do{}while(true);", 1},
		{"body always returns", "do{return [];}while(false);", 1},
		{"jump only in a nested closure", "do{$h=function(){do{break;}while(true);};return $h;}while($ok);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"OnlyWritesOnParameter"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("reach.php", []byte("<?php function f(array $v,$ok,$n){$v[]=1;"+tc.body+"return $v;}"), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
