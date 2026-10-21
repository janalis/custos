package fiberreturnbeforetermination

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FiberReturnBeforeTermination"}})
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
		{"\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();\n", 1},
		{"\n$f = new Fiber(fn() => 7); $f->start(); $f->getReturn();\n", 0},
		{"$f=new Fiber(function(){Fiber::suspend();return 1;});$f->start();$f->getReturn();", 1},
		{"$f=new Fiber(fn()=>Fiber::suspend());$f->start();$f->resume();$f->getReturn();", 0},
		{"$f=new Fiber(function(){});$f->start();$f->getReturn();", 0},
		{"$f=new Fiber(function(){return 1;});$f->start();$f->getReturn();", 0},
		{"$f=new Fiber(fn()=>Fiber::suspend());$f->getReturn();", 0},
		{"if(false){\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();\n}", 0},
		{"\n$f = new Fiber(fn() => Fiber::suspend()); $f->start(); $f->getReturn();\n $broken = ;", 0},
		{"$f=new Fiber(fn($x)=>Fiber::suspend());$f->start();$f->getReturn();", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
