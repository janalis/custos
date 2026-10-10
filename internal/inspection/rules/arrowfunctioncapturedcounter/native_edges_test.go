package arrowfunctioncapturedcounter

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"$o=new stdClass(); $f=fn()=>++$o->n; echo $f(),$f();", 0},
		{"$n=0; ($f=fn()=>++$n); echo $f(),$f();", 0},
		{"$n=0; $f=fn()=>1; echo $f(),$f();", 0},
		{"$f=fn($n)=>++$n; echo $f(1),$f(1);", 0},
		{"$n=[]; $f=fn()=>++$n; echo $f(),$f();", 0},
		{"$n=0; array_map(fn()=>++$n,$xs);", 0},
		{"$n=0; $obj->cb=fn()=>++$n;", 0},
		{"$n=0; $f=fn()=>++$n;", 0},
		{"$n=0; $f=fn()=>++$n; return $f();", 0},
		{"$n=0; $f=fn()=>++$n; echo 'label',$f();", 0},
		{"$n=0; $f=fn()=>++$n; echo other(),other();", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrowFunctionCapturedCounter"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
