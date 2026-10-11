package mysqliunbufferedresultblocksnextquery

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
		{"$db=new mysqli();$r=$db->query(\"UPDATE t SET v=1\",MYSQLI_USE_RESULT);$db->query(\"SELECT 1\");", 0, -1},
		{"$db=new mysqli();$r=$db->query($sql,MYSQLI_USE_RESULT);$db->query(\"SELECT 2\");", 0, -1},
		{"$db=new mysqli();$r=$db->query(\"\",MYSQLI_USE_RESULT);$db->query(\"SELECT 2\");", 0, -1},
		// Review: a use of the result earlier in the querying statement may consume or free it.
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);if($r->fetch_all()&&($x=$db->query(\"SELECT 2\"))){echo 1;}", 0, -1},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$x=[$r->free(),$db->query(\"SELECT 2\")];", 0, -1},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$db->query(drain($r));", 0, -1},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$r=&$db->query(\"SELECT 2\");", 0, -1},
		{"$db=new mysqli();$$n=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$db->query(\"SELECT 2\");", 0, -1},
		// Reassigning the variable, or using the result only after the query, leaves the result outstanding.
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$r=$db->query(\"SELECT 2\");", 1, 0},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);if($db->query(\"SELECT 2\")&&$r->fetch_all()){echo 1;}", 1, 0},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);if($other&&$db->query(\"SELECT 2\")){echo 1;}", 1, 0},
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
