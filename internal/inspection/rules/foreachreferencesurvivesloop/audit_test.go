package foreachreferencesurvivesloop

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
		{"$xs=[];foreach($xs as &$x){}$x=9;", 0, -1},
		{"$xs=[1];foreach($xs as &$x){unset($x);}$x=9;", 0, -1},
		{"foreach($a as &$v){$f=function(){unset($v);};}$v=3;", 1, -1},
		{"foreach($a as &$v){$v=&$other;}$v=3;", 0, -1},
		// Fresh array values: nothing else can observe the leftover reference.
		{"foreach([1,2] as &$v){}$v=3;", 0, -1},
		{"function f(){return [1,2];}foreach(f() as &$v){}$v=3;", 0, -1},
		{"foreach(array_keys($a) as &$v){}$v=3;", 0, -1},
		// Reference or spread items, by-reference returns and unresolved or
		// non-array calls keep the report.
		{"foreach([&$p,&$q] as &$v){}$v=3;", 1, -1},
		{"foreach([...$a] as &$v){}$v=3;", 1, -1},
		{"function &f(){static $s=[1];return $s;}foreach(f() as &$v){}$v=3;", 1, -1},
		{"function f(){yield 1;}foreach(f() as &$v){}$v=3;", 1, -1},
		{"foreach(g() as &$v){}$v=3;", 1, -1},
		{"foreach($o->$m() as &$v){}$v=3;", 1, -1},
		{"foreach(K::$m() as &$v){}$v=3;", 1, -1},
		{"$xs=[1,2];foreach($xs as &$v){}$v=3;", 1, -1},
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
