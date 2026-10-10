package catchvariableoverwriteslocal

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
		{"$x=\"ok\";try{throw new Exception();}catch(Exception $x){$x=\"ok\";}echo $x;", 0, -1},
		{"$x=\"ok\";try{throw new Exception();}catch(Exception $x){}finally{$x=\"ok\";}echo $x;", 0, -1},
		{"$x=\"ok\";try{throw new Exception();}catch(Exception $x){}finally{return;}echo $x;", 0, -1},
		{"$x=\"ok\";try{throw new Exception();}catch(Exception $x){foreach([\"ok\"] as $x){}}echo $x;", 0, -1},
		{"$e=\"original\";try {work();}catch(Exception $e){$f=function(){$e=\"restore\";};}echo $e;", 1, -1},
		{"$e=\"original\";try {work();}catch(Exception $e){$e++;}echo $e;", 0, -1},
		{"$e=\"original\";try {work();}catch(Exception $e){unset($e);}echo $e;", 0, -1},
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
