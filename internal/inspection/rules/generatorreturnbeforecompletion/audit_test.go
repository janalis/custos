package generatorreturnbeforecompletion

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestAuditRegressions exercises independent valid programs and comment-preservation boundaries.
func TestAuditRegressions(t *testing.T) {
	for _, tc := range []struct {
		src         string
		want, fixes int
	}{
		{"$g=(function(){if(false)yield 1;return 2;})();$g->getReturn();", 0, -1},
		{"$g=(function(){yield from [];return 2;})();$g->getReturn();", 0, -1},
		{"$g=makeGenerator();$g->getReturn();", 0, -1},
		{"$g=(function(){})();$g->getReturn();", 0, -1},
		// The read is reported wherever the statement evaluates it first.
		{"$g=(function(){yield 1;return 2;})();$r=$g->getReturn();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();return $g->getReturn();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();echo $g->getReturn(), 1;", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$g->getreturn();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$g->GETRETURN();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$r=(int)($g->getReturn());", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$r=!$g->getReturn();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$g->getReturn()+$g->next();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$g->getReturn()?1:2;", 1, -1},
		{"$g=(function(){yield 1;return 2;})();var_dump($g->getReturn(), $g->next());", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$g?->getReturn();", 1, -1},
		{"$g=(function(){yield 1;return 2;})();return strval($g->getReturn()) . 'x';", 1, -1},
		{"$g=(function(){yield 1;return 2;})();$r=[$g->next(),$g->getReturn()];", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$a[$g->next()]=$g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$$n=$g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$g->next().$g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$c?$g->getReturn():0;", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$c?0:$g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();f($g->next(),$g->getReturn());", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$f($g->getReturn());", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$o->m($g->getReturn());", 0, -1},
		{"$g=(function(){yield 1;return 2;})();echo $g->next(), $g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$c&&$g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$f=function()use($g){return $g->getReturn();};", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$f=fn()=>$g->getReturn();", 0, -1},
		{"$g=(function(){yield 1;return 2;})();if($g->getReturn()){}", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$g->getReturn()->x;", 0, -1},
		{"$g=(function(){yield 1;return 2;})();$r=$c??$g->getReturn();", 0, -1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("bad regression source: %+v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings want %d: %+v", len(got), tc.want, got)
			}
			if tc.fixes >= 0 {
				count := 0
				for _, d := range got {
					count += len(d.Fixes)
				}
				if count != tc.fixes {
					t.Fatalf("got %d fixes want %d", count, tc.fixes)
				}
			}
		})
	}
}
