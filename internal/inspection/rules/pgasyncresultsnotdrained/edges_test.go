package pgasyncresultsnotdrained

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
		{"pg_send_query();", 0},
		{"function f($c){if($flag){pg_send_query($c,'SELECT 2');}}", 0},
		{"function f($c){if(pg_send_query($c,$sql)){pg_send_query($c,'SELECT 2');}}", 0},
		{"function f($c){if(pg_send_query($c,'')){pg_send_query($c,'SELECT 2');}}", 0},
		{"function f($c){if(pg_send_query($c,\"SELECT 'x';SELECT 2\")){pg_send_query($c,'SELECT 2');}}", 0},
		{"function f($c){if(pg_send_query($c,'SELECT 1;SELECT 2')){$x=other();pg_send_query($c,'SELECT 2');}}", 0},
		{"function f($c){if(pg_send_query($c,'SELECT 1'))pg_send_query($c,'SELECT 2');}", 0},
		{"function f($c){if(pg_send_query($c,'SELECT 1')){pg_get_result($c);pg_send_query($c,'SELECT 2');}}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PgAsyncResultsNotDrained"}})
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
