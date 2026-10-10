package httperrorbodyreadwithoutignoreerrors

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"fopen('https://a.test','r');", 1},
		{"$ctx=stream_context_create(['http'=>['ignore_errors'=>false]]);stream_context_set_option($ctx,'http','ignore_errors',true);file_get_contents('https://a.test',false,$ctx);", 0},
		{"strlen('x');", 0},
		{"file_get_contents($x);", 0},
		{"file_get_contents('file:///tmp/a');", 0},
		{"file_get_contents('http://%xx');", 0},
		{"file_get_contents('https://a.test',false,$ctx);", 0},
		{"file_get_contents('https://a.test',false,unknown());", 0},
		{"file_get_contents('https://a.test',false,stream_context_create());", 1},
		{"file_get_contents('https://a.test',false,stream_context_create($x));", 0},
		{"file_get_contents('https://a.test',false,stream_context_create([]));", 1},
		{"file_get_contents('https://a.test',false,stream_context_create(['http'=>$x]));", 0},
		{"file_get_contents('https://a.test',false,stream_context_create(['http'=>[]]));", 1},
		{"file_get_contents('https://a.test',false,stream_context_create(['http'=>['ignore_errors'=>$x]]));", 0},
		{"file_get_contents('https://a.test',false,stream_context_create(['http'=>['ignore_errors'=>false]]));", 1},
		{"file_get_contents('https://a.test',false,stream_context_create(['http'=>['ignore_errors'=>true]]));", 0},
		{"fopen('https://a.test','r',false,stream_context_create(['http'=>['ignore_errors'=>true]]));", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HttpErrorBodyReadWithoutIgnoreErrors"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edge.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
