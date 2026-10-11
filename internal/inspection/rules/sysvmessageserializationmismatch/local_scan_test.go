package sysvmessageserializationmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestLocalScan covers inputs the shared resolvers still answer with a value,
// so the rule's own scan of the scope decides whether to report.
func TestLocalScan(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, Only: []string{"SysvMessageSerializationMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"send flag captured by reference before the calls", "<?php function f() { $q = msg_get_queue(35); $s = false; $set = function () use (&$s) { $s = true; }; if (msg_send($q, 3, 'hello', $s)) { msg_receive($q, 3, $t, 128, $m, true); } }", 0},
		{"send flag captured by reference later in the loop", "<?php function f() { $q = msg_get_queue(36); $s = false; foreach ([1, 2] as $round) { if (msg_send($q, 3, 'hello', $s)) { msg_receive($q, 3, $t, 128, $m, true); } $set = function () use (&$s) { $s = true; }; } }", 0},
		{"send flag captured by reference after the calls", "<?php function f() { $q = msg_get_queue(37); $s = false; if (msg_send($q, 3, 'hello', $s)) { msg_receive($q, 3, $t, 128, $m, true); } $set = function () use (&$s) { $s = true; }; }", 1},
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
