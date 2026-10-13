package filereadfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A result passed to a `string` parameter is reported on the call when the
// declaration that runs is known to declare it so (D1a).
func TestStringParameterOfTheDeclarationThatRuns(t *testing.T) {
	const (
		produce = "$v = file_get_contents($a); "
		typed   = "<?php class K { function take(string $s) {} static function make(string $s) {} } "
	)
	for _, tc := range []struct {
		name, source string
		version      phpversion.Version
		findings     int
	}{
		{"function", "<?php function take(string $s) {} function f($a) { " + produce + "take($v); }", phpversion.PHP85, 1},
		{"function below PHP 7", "<?php function take(string $s) {} function f($a) { " + produce + "take($v); }", phpversion.PHP56, 0},
		{"PHP function below PHP 7", "<?php function f($a) { " + produce + "return strlen($v); }", phpversion.PHP56, 1},
		{"function declared twice", "<?php if ($c) { function take(string $s) {} } else { function take($s) {} } function f($a) { " + produce + "take($v); }", phpversion.PHP85, 0},
		{"static method", typed + "function f($a) { " + produce + "K::make($v); }", phpversion.PHP85, 1},
		{"static method below PHP 7", typed + "function f($a) { " + produce + "K::make($v); }", phpversion.PHP56, 0},
		{"method of a typed receiver", typed + "function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 1},
		// `false` converts to text where null is refused: the call proves nothing.
		{"method read again after the call", typed + "function f($a, K $k) { " + produce + "$k->take($v); return strlen($v); }", phpversion.PHP85, 2},
		{"method of an unknown receiver", typed + "function f($a, $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"method of a class declared twice", "<?php if ($c) { class K { function take(string $s) {} } } else { class K { function take($s) {} } } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"method a class replaces with another type", typed + "class L extends K { function take($s) {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"method a class replaces with the same type", typed + "class L extends K { function take(string $s) {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 1},
		{"method a class replaces under another name", typed + "class L extends K { function take(string $t) {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"method a class replaces without the parameter", typed + "class L extends K { function take() {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"function parameter with a null default", "<?php function take(string $s = null) {} function f($a) { " + produce + "take($v); }", phpversion.PHP85, 0},
		{"function parameter with a text default", "<?php function take(string $s = 'null') {} function f($a) { " + produce + "take($v); }", phpversion.PHP85, 1},
		{"method parameter with a null default", "<?php final class K { function take(string $s = NULL) {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"static method parameter with a null default", "<?php class K { static function make(string $s = null) {} } function f($a) { " + produce + "K::make($v); }", phpversion.PHP85, 0},
		{"replacement with a null default", typed + "class L extends K { function take(string $s = null) {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"named argument, replacement declares it elsewhere", "<?php class K { function take(string $s, $t = 1) {} } class L extends K { function take($t = 1, string $s = '') {} } function f($a, K $k) { " + produce + "$k->take(s: $v); }", phpversion.PHP85, 1},
		{"named argument, replacement declares it untyped elsewhere", "<?php class K { function take(string $s, $t = 1) {} } class L extends K { function take($t = 1, $s = '') {} } function f($a, K $k) { " + produce + "$k->take(s: $v); }", phpversion.PHP85, 0},
		{"named argument, replacement collects it", "<?php class K { function take(string $s, $t = 1) {} } class L extends K { function take(...$all) {} } function f($a, K $k) { " + produce + "$k->take(s: $v); }", phpversion.PHP85, 0},
		{"positional argument, replacement declares the name elsewhere", "<?php class K { function take(string $s, $t = 1) {} } class L extends K { function take($t = 1, string $s = '') {} } function f($a, K $k) { " + produce + "$k->take($v); }", phpversion.PHP85, 0},
		{"method a class replaces, object made here", typed + "class L extends K { function take($s) {} } function f($a) { $k = new K; " + produce + "$k->take($v); }", phpversion.PHP85, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.version, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.ParseBest("probe.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}

// Where a label precedes the use, a guard written before it proves nothing
// by its position: only the paths the flow analysis follows count.
func TestGuardsBeforeALabel(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"guard jumped over", "<?php function f($a) { $v = file_get_contents($a); goto use_it; if ($v === false) { return; } use_it: return strlen($v); }", 1},
		{"jump into the guarded branch", "<?php function f($a, $c) { $v = file_get_contents($a); if ($c) { goto in; } if (is_string($v)) { in: return strlen($v); } }", 1},
		{"failure jumps to the use", "<?php function f($a) { $v = file_get_contents($a); if ($v === false) { goto fail; } return 1; fail: return strlen($v); }", 1},
		{"guard on every path", "<?php function f($a, $c) { again: $v = file_get_contents($a); if ($v === false) { return; } if ($c) { goto again; } return strlen($v); }", 0},
		{"label after the use", "<?php function f($a, $c) { $v = file_get_contents($a); if (is_string($v)) { echo strlen($v); } if ($c) { goto done; } echo 1; done: ; }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("probe.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
