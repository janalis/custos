package fiberstartedtwice

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FiberStartedTwice"}})
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
		{"\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();\n", 1},
		{"\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->resume();\n", 0},
		{"$f=new Fiber(fn()=>1);if($b){$f->start();}$f->start();", 0},
		{"$f=new Fiber(fn()=>1);$f->start();$f=new Fiber(fn()=>1);$f->start();", 0},
		{"$f=new Fiber(fn()=>1);$f->start();unknown($f);$f->start();", 0},
		{"function f(Fiber $f){$f->start();$f->start();}", 0},
		{"if(false){\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();\n}", 0},
		{"\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->start();\n $broken = ;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
