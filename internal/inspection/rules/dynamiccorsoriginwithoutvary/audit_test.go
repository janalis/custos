package dynamiccorsoriginwithoutvary

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
		{"$_SERVER[\"HTTP_ORIGIN\"]=\"trusted\";header(\"Access-Control-Allow-Origin: \".$_SERVER[\"HTTP_ORIGIN\"]);header(\"Cache-Control: public,max-age=60\");", 0, -1},
		// A Vary header sent before the origin header covers the same response.
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Origin');header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 0, -1},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: *');header('X-A: b');header('HTTP/1.1 200 OK');header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 0, -1},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Origin');header('Vary: Accept', false);header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 0, -1},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Origin');header('Access-Control-Allow-Origin: '.$o);header('Vary: Accept', false);header('Cache-Control: public');", 0, -1},
		{"$o=$_SERVER['HTTP_ORIGIN'];header($line);header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 0, -1},
		// A replaced or removed earlier Vary header no longer covers it.
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Origin');header('Vary: Accept');header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 1, 0},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Origin');header('Access-Control-Allow-Origin: '.$o);header('Vary: Accept');header('Cache-Control: public');", 1, 0},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Origin');header_remove('Vary');header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 1, 0},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('Vary: Accept', false);header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 1, 0},
		{"$o=$_SERVER['HTTP_ORIGIN'];header('X-A: b');header('Access-Control-Allow-Origin: '.$o);header('Cache-Control: public');", 1, 0},
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
