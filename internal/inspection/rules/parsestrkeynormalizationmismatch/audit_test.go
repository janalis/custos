package parsestrkeynormalizationmismatch

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
		{"parse_str(\"a.b=x\",$out);$out[\"a.b\"]=\"new\";", 0, -1},
		{"parse_str(\"a.b=x\",$out);unset($out[\"a.b\"]);", 0, -1},
		{"parse_str(\"a.b=x\",$out);echo count($out=[\"a.b\"=>\"ok\"]),$out[\"a.b\"];", 0, -1},
		{"parse_str(\"%20a.b=x\",$out);echo $out[\" a.b\"];", 0, -1},
		{"parse_str(\"a.b=x\",$q);echo $q[\"n\"][\"a.b\"];", 0, -1},
		// Review: a nested scope owns its own variable of the same name.
		{"parse_str(\"a.b=x\",$q);$pick=fn($q)=>$q[\"a.b\"];", 0, -1},
		{"parse_str(\"a.b=x\",$q);$pick=fn()=>$q[\"a.b\"];", 0, -1},
		{"parse_str(\"a.b=x\",$q);$pick=function($q){return $q[\"a.b\"];};", 0, -1},
		{"function f(){parse_str(\"a.b=x\",$q);return $q[\"a.b\"];}", 1, 1},
		{"$f=fn($s)=>[parse_str(\"a.b=x\",$q),$q[\"a.b\"]];", 0, -1},
		// Review: reference bindings create the dimension instead of reading it.
		{"parse_str(\"a.b=x\",$q);$slot=&$q[\"a.b\"];", 0, -1},
		{"parse_str(\"a.b=x\",$q);$refs=[&$q[\"a.b\"]];", 0, -1},
		{"parse_str(\"a.b=x\",$q);foreach($q[\"a.b\"] as &$item){}", 0, -1},
		{"parse_str(\"a.b=x\",$q);$copy=$q[\"a.b\"];", 1, 1},
		{"parse_str(\"a.b=x\",$q);$map=[$q[\"a.b\"]=>&$other];", 1, 1},
		{"parse_str(\"a.b=x\",$q);$list=[$q[\"a.b\"]];", 1, 1},
		{"parse_str(\"a.b=x\",$q);foreach($q[\"a.b\"] as $item){}", 1, 1},
		// Review: a fetch used as the key of a by-reference access is still read.
		{"parse_str(\"a.b=x\",$q);$slot=&$cache[$q[\"a.b\"]];", 1, 1},
		{"parse_str(\"a.b=x\",$q);$refs=[&$cache[$q[\"a.b\"]]];", 1, 1},
		{"parse_str(\"a.b=x\",$q);foreach($cache[$q[\"a.b\"]] as &$item){}", 1, 1},
		{"parse_str(\"a.b=x\",$q);$slot=&$q[\"a.b\"][\"c\"];", 0, -1},
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
