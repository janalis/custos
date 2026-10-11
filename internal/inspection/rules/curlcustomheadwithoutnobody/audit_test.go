package curlcustomheadwithoutnobody

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
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");curl_setopt($h,CURLOPT_NOBODY,true);curl_setopt($h,CURLOPT_HTTPGET,true);curl_exec($h);", 1, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");curl_setopt($h,CURLOPT_HTTPGET,$unknown);curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$f=function()use($h){curl_setopt($h,CURLOPT_NOBODY,true);};$f();curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$f=fn()=>curl_setopt($h,CURLOPT_NOBODY,true);$f();curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");configure([\"handle\"=>$h]);curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$this->handle=$h;$this->configure();curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$b=$h;$f=function()use($b){curl_setopt($b,CURLOPT_NOBODY,true);};$f();curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$b=&$h;curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$n=\"h\";configure($$n);curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$k=@$h;curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$k=$h?:$s;curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$k=$h??$s;curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");if($h===false||!$h||empty($h)||!isset($h)){return;}curl_exec($h);", 1, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");$f=function()use($x){return $x;};$g=fn($h)=>1;curl_exec($h);", 1, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");curl_exec($h);$keep=[$h];", 1, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");function other(){$h=[$h];}curl_exec($h);", 1, -1},
		{"$o->h=curl_init();curl_setopt($o->h,CURLOPT_CUSTOMREQUEST,\"HEAD\");curl_exec($o->h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");configure(compact(\"h\"));curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");configure(get_defined_vars());curl_exec($h);", 0, -1},
		{"$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,\"HEAD\");curl_exec($h);configure(compact(\"h\"));", 1, -1},
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
