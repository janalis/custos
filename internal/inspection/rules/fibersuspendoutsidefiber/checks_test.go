package fibersuspendoutsidefiber

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FiberSuspendOutsideFiber"}})
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
		{"\nif (Fiber::getCurrent() === null) { Fiber::suspend(); }\n", 1},
		{"\nif (Fiber::getCurrent() !== null) { Fiber::suspend(); }\n", 0},
		{"if(null===Fiber::getCurrent()){Fiber::suspend();}", 1},
		{"Fiber::suspend();", 0},
		{"if($b){Fiber::suspend();}", 0},
		{"if(Fiber::getCurrent()===null){$f=fn()=>Fiber::suspend();}", 0},
		{"if(Fiber::getCurrent()===null){}else{Fiber::suspend();}", 0},
		{"if(false){\nif (Fiber::getCurrent() === null) { Fiber::suspend(); }\n}", 0},
		{"\nif (Fiber::getCurrent() === null) { Fiber::suspend(); }\n $broken = ;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
