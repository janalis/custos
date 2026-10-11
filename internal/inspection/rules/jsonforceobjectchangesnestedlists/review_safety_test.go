package jsonforceobjectchangesnestedlists

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestReviewSafety pins list-order detection and the scope constructs that
// rewrite a tracked array without the flow engine following them.
func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"ordered list", `json_encode(['a' => [0 => 'x', 1 => 'y']], JSON_FORCE_OBJECT);`, 1},
		{"numeric string keys", `json_encode(['a' => ['0' => 'x', '1' => 'y']], JSON_FORCE_OBJECT);`, 1},
		{"overwrite keeps position", `json_encode(['a' => ['x', 'y', 0 => 'z']], JSON_FORCE_OBJECT);`, 1},
		{"reversed keys", `json_encode(['a' => [1 => 'x', 0 => 'y']], JSON_FORCE_OBJECT);`, 0},
		{"gap", `json_encode(['a' => [0 => 'x', 2 => 'y']], JSON_FORCE_OBJECT);`, 0},
		{"negative key", `json_encode(['a' => [-1 => 'x', 'y']], JSON_FORCE_OBJECT);`, 0},
		{"string key", `json_encode(['a' => ['x', 'k' => 'y']], JSON_FORCE_OBJECT);`, 0},
		{"deeper list behind map", `json_encode(['a' => [1 => ['x']]], JSON_FORCE_OBJECT);`, 1},
		{"top-level variable", `$v = ['a' => [1, 2]]; json_encode($v, JSON_FORCE_OBJECT);`, 1},
		{"top-level by-reference foreach", `$v = ['a' => [1, 2]]; foreach ($v as &$x) { $x = 1; } json_encode($v, JSON_FORCE_OBJECT);`, 0},
		{"keyed by-reference foreach", `function f() { $v = ['a' => [1, 2]]; foreach ($v as $k => &$x) { $x = 1; } return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"by-value foreach", `function f() { $v = ['a' => [1, 2]]; foreach ($v as $x) { echo count($x); } return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"by-value destructuring", `function f() { $v = ['a' => [1, 2]]; foreach ($v as [$x, $y]) { echo $x; } return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"reference destructuring", `function f() { $v = ['a' => [1, 2]]; foreach ($v as [&$x]) { $x = [9 => 1]; } return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"foreach element target", `function f($r) { $v = ['a' => [1, 2]]; foreach ($r as $v['a']) {} return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"foreach key target", `function f($r) { $v = ['a' => [1, 2]]; foreach ($r as $v['a'] => $x) {} return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"reference capture", `function f() { $v = ['a' => [1, 2]]; $g = function () use (&$v) { $v = 1; }; $g(); return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"value capture", `function f() { $v = ['a' => [1, 2]]; $g = function () use ($v) {}; return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"other reference capture", `function f() { $v = ['a' => [1, 2]]; $o = 1; $g = function () use (&$o) {}; return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"catch binding", `function f() { $v = ['a' => [1, 2]]; try { g(); } catch (E $v) {} return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"other catch binding", `function f() { $v = ['a' => [1, 2]]; try { g(); } catch (E $e) {} catch (F) {} return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"extract", `function f($i) { $v = ['a' => [1, 2]]; extract($i); return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"copy of rewritten", `function f() { $w = ['a' => [1, 2]]; foreach ($w as &$x) { $x = 1; } $v = $w; return json_encode($v, JSON_FORCE_OBJECT); }`, 0},
		{"copy cycle", `function f() { $w = ['a' => [1, 2]]; $v = $w; $w = $v; return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"plain copy", `function f() { $w = ['a' => [1, 2]]; $v = $w; return json_encode($v, JSON_FORCE_OBJECT); }`, 1},
		{"nested variable rewritten", `function f() { $w = [1, 2]; foreach ($w as &$x) { $x = 1; } return json_encode(['a' => $w], JSON_FORCE_OBJECT); }`, 0},
		{"nested variable intact", `function f() { $w = [1, 2]; return json_encode(['a' => $w], JSON_FORCE_OBJECT); }`, 1},
		{"other scope is separate", `function f() { $v = ['a' => [1, 2]]; return json_encode($v, JSON_FORCE_OBJECT); } function g($v) { foreach ($v as &$x) { $x = 1; } }`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("bad source: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d findings want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
