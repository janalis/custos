package generatorrewindafteradvance

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"GeneratorRewindAfterAdvance"}})
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
		{"\n$g = (function() { yield 7; yield 11; })(); $g->next(); $g->rewind();\n", 1},
		{"\n$g = (function() { yield 7; yield 11; })(); $g->rewind();\n", 0},
		{"function values(){yield 1;yield 2;}$g=values();$g->send(3);$g->rewind();", 1},
		{"$g=(function(){yield 1;})();$g->next();$g->rewind();", 0},
		{"$g=(function(){if($b){yield 1;}yield 2;})();$g->next();$g->rewind();", 0},
		{"$g=(function(){yield 1;return;})();$g->next();$g->rewind();", 0},
		{"$g=unknown();$g->next();$g->rewind();", 0},
		{"$g=$unknown;$g->rewind();", 0},
		{"$g=(function(){yield 1;yield 2;})();$g->next();$g->{$name}();", 0},
		{"if(false){\n$g = (function() { yield 7; yield 11; })(); $g->next(); $g->rewind();\n}", 0},
		{"\n$g = (function() { yield 7; yield 11; })(); $g->next(); $g->rewind();\n $broken = ;", 0},
		{"$g=(fn()=>1)();$g->rewind();", 0},
		{"$g=(function(){$x=1;yield 2;yield 3;})();$g->rewind();", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
