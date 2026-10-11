package arraymapmissingcallbackreturn

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewReplacementCounter(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"array_map(function($v){str_replace('x','y',$v);},$names);", 1},
		{"array_map(function($v){str_replace(search:'x',replace:'y',subject:$v);},$names);", 1},
		{"array_map(function($v)use(&$t){str_replace('x','y',$v,$n);$t+=$n;},$names);", 0},
		{"array_map(function($v)use(&$t){str_replace('x','y',$v,$t);},$names);", 0},
		{"array_map(function($v)use(&$t){str_replace('x','y',$v,count:$t);},$names);", 0},
		{"array_map(function($v)use(&$t){str_replace('x','y',subject:$v,COUNT:$t);},$names);", 0},
		{"array_map(function($v)use($a){str_replace(...$a);},$names);", 0},
		{"array_map(function($v){str_replace(...);},$names);", 0},
		{"array_map(function($v)use(&$t){str_replace('x','y',$v,$t);trim($v);},$names);", 1},
		{"array_map(function($v)use(&$t){trim($v);str_replace('x','y',$v,$t);},$names);", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ArrayMapMissingCallbackReturn"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
