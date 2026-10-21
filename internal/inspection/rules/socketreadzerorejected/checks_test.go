package socketreadzerorejected

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SocketReadZeroRejected"}})
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
		{"\nif (!$data = socket_read($s, 1024, PHP_BINARY_READ)) { throw new RuntimeException(); }\n", 1},
		{"\n$data = socket_read($s, 1024, PHP_BINARY_READ); if ($data === false) { throw new RuntimeException(); }\n", 0},
		{"if(socket_read($s,10,PHP_BINARY_READ)==false){return false;}", 1},
		{"if(!socket_read($s,10,PHP_NORMAL_READ)){return false;}", 0},
		{"if(!socket_read($s,10,PHP_BINARY_READ)){echo 'empty';}", 0},
		{"if(false){\nif (!$data = socket_read($s, 1024, PHP_BINARY_READ)) { throw new RuntimeException(); }\n}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
