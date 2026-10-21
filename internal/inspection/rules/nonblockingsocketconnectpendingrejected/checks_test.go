package nonblockingsocketconnectpendingrejected

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NonblockingSocketConnectPendingRejected"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\nif (socket_set_nonblock($s)) { if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); } }\n", 1},
		{"\nsocket_set_block($s); if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); }\n", 0},
		{"if(socket_set_nonblock($s)){if(!socket_connect($s,$h,80)){socket_last_error($s);return false;}}", 0},
		{"socket_set_nonblock($s);if(!socket_connect($s,$h,80)){return false;}", 1},
		{"if(socket_set_nonblock($s)){socket_set_block($s);if(!socket_connect($s,$h,80)){return false;}}", 0},
		{"if(false){\nif (socket_set_nonblock($s)) { if (!socket_connect($s, $host, 443)) { throw new RuntimeException(); } }\n}", 0},
		{"socket_connect($s,$h,80);", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
