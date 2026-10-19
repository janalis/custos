package filereadfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A call that can enter the function again assigns its static locals: a test
// made before such a call does not cover a use made after it. A guard form
// only the rule's own walk reads is held to that like one the flow reads.
func TestStaticLocalAcrossReentry(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const (
		static = "<?php function f($c, $o, $k) { static $p; $p = file_get_contents($c); "
		plain  = "<?php function f($c, $o, $k) { $p = file_get_contents($c); "
		strict = "if ($p === false) { return; } "
		walked = "if ($p == '') { return; } "
		use    = "echo strlen($p);"
	)
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"strict test, use", static + strict + use + " }", 0},
		{"strict test, call, use", static + strict + "$o->run(); " + use + " }", 1},
		{"strict test, call after the use", static + strict + use + " $o->run(); }", 0},
		{"call, strict test, use", static + "$o->run(); " + strict + use + " }", 0},
		{"walked test, use", static + walked + use + " }", 0},
		{"walked test, call, use", static + walked + "$o->run(); " + use + " }", 1},
		{"walked test, call after the use", static + walked + use + " $o->run(); }", 0},
		{"call, walked test, use", static + "$o->run(); " + walked + use + " }", 0},
		{"walked test, builtin call, use", static + walked + "strlen('a'); " + use + " }", 0},
		{"walked test, call on one path, use", static + walked + "if ($k) { $o->run(); } " + use + " }", 1},
		{"walked test, use and call in a loop", static + walked + "foreach ($k as $e) { " + use + " $o->run(); } }", 1},
		{"walked test and use in a loop with a call", static + "foreach ($k as $e) { " + walked + use + " $o->run(); } }", 0},
		{"call on one path, walked test, call on one path, use", static + "if ($k) { $o->run(); } " + walked + "if ($k) { $o->run(); } " + use + " }", 1},
		{"call on one path, walked test, use", static + "if ($k) { $o->run(); } " + walked + use + " }", 0},
		{"plain local, strict test, call, use", plain + strict + "$o->run(); " + use + " }", 0},
		{"plain local, walked test, call, use", plain + walked + "$o->run(); " + use + " }", 0},
		{"plain local, call on one path, walked test, use", plain + "if ($k) { $o->run(); } " + walked + use + " }", 0},
		{"tested in place, use", "<?php function f($c, $o) { static $p; if (($p = file_get_contents($c)) == '') { return; } echo strlen($p); }", 0},
		{"tested in place, call, use", "<?php function f($c, $o) { static $p; if (($p = file_get_contents($c)) == '') { return; } $o->run(); echo strlen($p); }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("static.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
