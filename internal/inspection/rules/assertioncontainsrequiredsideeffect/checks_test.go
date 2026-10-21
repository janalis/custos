package assertioncontainsrequiredsideeffect

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"AssertionContainsRequiredSideEffect"}})
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
		{"\nassert(($result = 7) > 0); echo $result;\n", 1},
		{"\n$result = 7; assert($result > 0); echo $result;\n", 0},
		{"function f(){assert(($x=1)>0);echo $x;}", 1},
		{"function f($x){assert(($x=1)>0);echo $x;}", 0},
		{"assert(($a[0]=1)>0);echo $a;", 0},
		{"assert(($x=1)>0);$x=2;echo $x;", 0},
		{"assert(($x=1)>0);echo $x;$x=2;", 1},
		{"echo $x;assert(($x=1)>0);echo $x;", 0},
		{"function f(){global $x;assert(($x=1)>0);echo $x;}", 0},
		{"function f(){static $x;assert(($x=1)>0);echo $x;}", 0},
		{"assert(($x=1)>0);isset($x);", 0},
		{"assert(($x=1)>0);empty($x);", 0},
		{"assert(($x=1)>0);$y=$x??2;", 0},
		{"assert(($x=1)>0);if($b){echo $x;}", 0},
		{"assert(($x=1)>0);$f=function(){echo $x;};", 0},
		{"assert((function(){$x=1;return true;})());echo $x;", 0},
		{"assert(($x=&$y)>0);echo $x;", 0},
		{"assert(($x=1)>0);", 0},
		{"if(false){\nassert(($result = 7) > 0); echo $result;\n}", 0},
		{"\nassert(($result = 7) > 0); echo $result;\n $broken = ;", 0},
		{"assert(($x=1)>0);" + strings.Repeat("$y=1;", 513) + "echo $x;", 0}, // past the budget
		{"assert(($x=1)>0);unset($x);echo $x;", 0},
		{"assert(($x=1)>0);$alias=&$x;echo $x;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
