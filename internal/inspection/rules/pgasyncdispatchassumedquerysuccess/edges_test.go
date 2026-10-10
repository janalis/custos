package pgasyncdispatchassumedquerysuccess

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f($c){if(pg_send_query($c,'SELECT 3')) echo 'ok';}", 0},
		{"function f($c){if(pg_send_query($c,'SELECT 3')){echo 'ok';}}", 1},
		{"function f($c){if(pg_send_query($c,'SELECT 3')){$r=pg_get_result($c);return true;}}", 0},
		{"function f($c){if(pg_send_query($c,'SELECT 3')){return false;}}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PgAsyncDispatchAssumedQuerySuccess"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
