package forgottendebugoutput

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestFixtureFiles: E8 skips files under a Fixtures directory only.
func TestFixtureFiles(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ForgottenDebugOutput"}})
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]int{
		"tests/Fixtures/dump_server.php": 0,
		`tests\Fixtures\dump_server.php`: 0,
		"tests/DumpServerTest.php":       1,
		"src/fixtures/dump.php":          1,
	} {
		got := e.Analyze(syntax.Parse(path, []byte("<?php var_dump($x);"), syntax.Options{}))
		if len(got) != want {
			t.Errorf("%s: got %d findings, want %d", path, len(got), want)
		}
	}
}
