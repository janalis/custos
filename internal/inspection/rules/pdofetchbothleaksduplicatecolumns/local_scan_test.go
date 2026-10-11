package pdofetchbothleaksduplicatecolumns

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestLocalScan covers inputs the shared resolvers still answer with a value,
// so the rule's own scan of the scope decides whether to report.
func TestLocalScan(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"PdoFetchBothLeaksDuplicateColumns"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"rows bound by reference after encoding", "<?php function f(PDOStatement $s) { $rows = $s->fetchAll(PDO::FETCH_BOTH); $json = json_encode($rows); foreach ($rows as &$row) { $row = 1; } return $json; }", 0},
		{"row field bound by reference after encoding", "<?php function f(PDOStatement $s) { $rows = $s->fetchAll(PDO::FETCH_BOTH); $json = json_encode($rows); foreach ($rows as ['name' => &$name]) { $name = 1; } return $json; }", 0},
		{"file-level rows bound by reference after encoding", "<?php $stmt = (new PDO('sqlite::memory:'))->query('SELECT 1'); $all = $stmt->fetchAll(PDO::FETCH_BOTH); echo json_encode($all); foreach ($all as &$one) { $one = 1; }", 0},
		{"rows iterated by value", "<?php function f(PDOStatement $s) { $rows = $s->fetchAll(PDO::FETCH_BOTH); foreach ($rows as $row) { } return json_encode($rows); }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("scan.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
