package parseurlfailuredereferenced

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// Where a label precedes the read, a guard written before it proves nothing
// by its position: only the paths the flow analysis follows count. A read
// after a statement that ends its list is reached through a label a `goto`
// names.
func TestGuardsBeforeALabel(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"guard jumped over", "<?php function f($a) { $u = parse_url($a); goto use_it; if ($u === false) { return; } use_it: echo $u['host']; }", 1},
		{"jump into the guarded branch", "<?php function f($a, $c) { $u = parse_url($a); if ($c) { goto in; } if ($u !== false) { in: echo $u['host']; } }", 1},
		{"jump over an early exit", "<?php function f($a, $c) { $u = parse_url($a); if ($c) { goto mid; } if ($u === false) { return; } mid: echo $u['host']; }", 1},
		{"failure jumps to the read", "<?php function f($a) { $u = parse_url($a); if ($u === false) { goto fail; } return; fail: echo $u['host']; }", 1},
		{"failure jumps out of a loop to the read", "<?php function f($l) { foreach ($l as $n) { $u = parse_url($n); if ($u === false) { goto out; } } return; out: echo $u['host']; }", 1},
		{"success jumps to the read", "<?php function f($a) { $u = parse_url($a); if ($u !== false) { goto ok; } return; ok: echo $u['host']; }", 0},
		{"failure jumps past the read", "<?php function f($a) { $u = parse_url($a); if ($u === false) { goto end; } echo $u['host']; end: return; }", 0},
		{"guard taken again on every turn", "<?php function f($a, $l) { again: $u = parse_url($a); if ($u === false) { $a = next($l); goto again; } echo $u['host']; }", 0},
		{"label nothing jumps to", "<?php function f($a) { $u = parse_url($a); return; dead: echo $u['host']; }", 0},
		{"guard by position, label and try", "<?php function f($a) { $u = parse_url($a); if (!isset($u['host'])) { return; } try { g(); } finally { h(); } goto ok; ok: echo $u['host']; }", 0},
		{"guard by position, label after the read and try", "<?php function f($a) { $u = parse_url($a); if (!isset($u['host'])) { return; } echo $u['host']; try { g(); } finally { h(); } goto done; done: ; }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("label.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
