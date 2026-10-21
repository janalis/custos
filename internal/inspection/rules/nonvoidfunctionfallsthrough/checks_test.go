package nonvoidfunctionfallsthrough

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NonVoidFunctionFallsThrough"}})
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
		{"\nfunction countItems(bool $ready): int { if ($ready) return 7; }\n", 1},
		{"\nfunction countItems(bool $ready): int { if ($ready) return 7; return 0; }\n", 0},
		{"function f():int{}", 1},
		{"function f():void{}", 0},
		{"function f():mixed{}", 0},
		{"function f():never{throw new Error();}", 0},
		{"function f(){}", 0},
		{"interface X{function f():int;}", 0},
		{"$f=function():int{};", 1},
		{"class X{function f():int{if(false){return 1;}}}", 1},
		{"function f():int{if(true){return 1;}}", 0},
		{"function f():int{if(false){return 1;}else{return 2;}}", 0},
		{"function f():int{if($b){return 1;}else{$x=2;}}", 1},
		{"function f():int{if($b){return 1;}elseif($c){return 2;}}", 0},
		{"function f():int{unknown();}", 0},
		{"function f():int{if(unknown()){return 1;}}", 0},
		{"function f():int{while($b){return 1;}}", 0},
		{"function f():int{throw new Error();}", 0},
		{"function f():int{yield 1;}", 0},
		{"function f():int{echo 'x';;}", 1},
		{"if(false){\nfunction countItems(bool $ready): int { if ($ready) return 7; }\n}", 0},
		{"\nfunction countItems(bool $ready): int { if ($ready) return 7; }\n $broken = ;", 0},
		{"function f():int{" + strings.Repeat(";", 513) + "}", 0}, // past the budget
		{"function f():int{$x=fn()=>unknown();}", 1},
		{"function f():int{if(1){return 1;}}", 0},
		{"function f():int{if(1===1){return 1;}}", 0},
		{"function f($x):int{if($x===$x){return 1;}}", 0},
		{"function f($x):int{if($x>0){return 1;}}", 1},
		{"function f($x,$y):int{if($x&&$y){return 1;}}", 0},
		{"function f($x):int{if(!$x){return 1;}}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
