package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// Guards that are not written as a branch: a statement that leaves on one
// outcome of a test, a call that never returns, a test kept in a boolean, a
// function that only returns for an array.
func TestGuardsOutsideBranches(t *testing.T) {
	const u = `$u=fopen($a,'r'); `
	for _, tc := range []struct {
		name, src         string
		nonFalse, nonNull bool
	}{
		{"or exit", u + `$u !== false or exit(1); probe($u);`, true, false},
		{"or die", u + `$u or die('x'); probe($u);`, true, true},
		{"logical or exit", u + `$u !== false || exit; probe($u);`, true, false},
		{"or throw", u + `$u or throw new E(); probe($u);`, true, true},
		{"and exit", u + `$u === false and exit; probe($u);`, true, false},
		{"logical and throw", u + `$u === false && throw new E(); probe($u);`, true, false},
		{"assignment or exit", u + `$u=fopen($a,'w') or die(); probe($u);`, true, true},
		{"parenthesised assignment or exit", u + `($u=fopen($a,'w')) || die(); probe($u);`, true, true},
		{"short ternary throw", u + `$u ?: throw new E(); probe($u);`, true, true},
		{"or exit on the wrong outcome", u + `$u === false or exit; probe($u);`, false, false},
		{"and exit on the wrong outcome", u + `$u !== false and exit; probe($u);`, false, false},
		{"or a call that returns", u + `$u !== false or report('x'); probe($u);`, false, false},
		{"or a builtin that returns", u + `$u !== false or error_log('x'); probe($u);`, false, false},
		{"long ternary throw", u + `$u ? 1 : throw new E(); probe($u);`, false, false},
		{"short ternary without exit", u + `$u ?: report(); probe($u);`, false, false},
		{"other operator", u + `$u !== false xor exit; probe($u);`, false, false},
		{"exit in a try", u + `try { $u !== false or throw new E(); } catch (E $e) {} probe($u);`, false, false},
		{"exit statement", u + `if ($u === false) { exit; } probe($u);`, true, false},
		// Calls that never return.
		{"declared never", `function fail(): never { throw new E; } ` + u + `if ($u === false) { fail(); } probe($u);`, true, false},
		{"declared never, or form", `function fail(): never { throw new E; } ` + u + `$u !== false or fail(); probe($u);`, true, false},
		{"declared never method", `class T { function fail(): never { throw new E; } function f($a) { ` + u + `if ($u === false) { $this->fail(); } probe($u); } }`, true, false},
		{"declared never static method", `class T { static function fail(): never { throw new E; } function f($a) { ` + u + `if ($u === false) { self::fail(); } probe($u); } }`, true, false},
		{"declared never method of another class", `class H { static function fail(): never { throw new E; } } ` + u + `if ($u === false) { H::fail(); } probe($u);`, true, false},
		{"declared void", `function fail(): void { error_log('x'); } ` + u + `if ($u === false) { fail(); } probe($u);`, false, false},
		{"unresolved function", u + `if ($u === false) { fail(); } probe($u);`, false, false},
		{"unresolved method", u + `if ($u === false) { $o->fail(); } probe($u);`, false, false},
		{"dynamic static call", u + `if ($u === false) { $c::fail(); } probe($u);`, false, false},
		{"dynamic method name", `class T { function f($a, $m) { ` + u + `if ($u === false) { $this->$m(); } probe($u); } }`, false, false},
		{"never call in another closure", `function fail(): never { throw new E; } ` + u + `$g = function () use ($u) { if ($u === false) { fail(); } }; probe($u);`, false, false},
		// A test kept in a boolean.
		{"stored test, failed exit", u + `$ok = $u !== false; if (!$ok) { return; } probe($u);`, true, false},
		{"stored test, branch", u + `$ok = ($u !== false); if ($ok) { probe($u); }`, true, false},
		{"stored failure", u + `$failed = $u === false; if ($failed) { return; } probe($u);`, true, false},
		{"stored predicate", u + `$ok = is_resource($u); if ($ok) { probe($u); }`, true, true},
		{"stored negation", u + `$bad = !$u; if ($bad) { return; } probe($u);`, true, true},
		{"stored conjunction", u + `$v = fopen($a,'w'); $ok = $u !== false && $v !== false; if ($ok) { probe($u); }`, true, false},
		{"stored isset", u + `$ok = isset($u); if ($ok) { probe($u); }`, false, true},
		{"stored emptiness", u + `$none = empty($u); if (!$none) { probe($u); }`, true, true},
		{"stored instance test", u + `$ok = $u instanceof X; if ($ok) { probe($u); }`, true, true},
		{"stored test, wrong branch", u + `$ok = $u !== false; if ($ok) { return; } probe($u);`, false, false},
		{"stored test of an earlier value", u + `$ok = $u !== false; $u = fopen($a,'w'); if (!$ok) { return; } probe($u);`, false, false},
		{"stored test, flag overwritten", u + `$ok = $u !== false; $ok = $c; if (!$ok) { return; } probe($u);`, false, false},
		{"stored test, flag overwritten on one path", u + `$ok = $u !== false; if ($c) { $ok = true; } if (!$ok) { return; } probe($u);`, false, false},
		{"stored test, value passed by reference", u + `$ok = $u !== false; change($u); if (!$ok) { return; } probe($u);`, false, false},
		{"stored test, value replaced by a holder", u + `$g = function () use (&$u) { $u = false; }; $ok = $u !== false; $g(); if (!$ok) { return; } probe($u);`, false, false},
		{"stored test of a parameter", `function f($u) { $ok = $u !== false; if (!$ok) { return; } probe($u); }`, false, false},
		{"stored test that assigns", u + `$ok = ($u = fopen($a,'w')) !== false; if (!$ok) { return; } probe($u);`, false, false},
		{"stored test of a variable variable", u + `$ok = $$n !== false; if (!$ok) { return; } probe($u);`, false, false},
		{"stored copy", u + `$ok = $u; if (!$ok) { return; } probe($u);`, false, false},
		{"stored test in another closure", u + `$ok = $u !== false; $g = function () use ($ok, $u) { if (!$ok) { return; } probe($u); };`, false, false},
		// Functions that only take arrays throw on false since PHP 8.
		{"array-only parameter", `$u=scandir($a); if (count($u) > 0) { probe($u); }`, true, true},
		{"array-only parameter, later statement", `$u=scandir($a); $n = in_array('x', $u); probe($u);`, true, true},
		{"nullable array parameter", `$u=scandir($a); $x = array_merge_recursive(...$u); $m = new ArrayIterator(); probe($u);`, false, false},
		{"scalar parameter", `$u=scandir($a); if (strlen($u) > 0) { probe($u); }`, false, false},
		{"mixed parameter", `$u=scandir($a); var_dump($u); probe($u);`, false, false},
		{"array-only parameter in a try", `$u=scandir($a); try { $n = count($u); } catch (TypeError $e) { $n = 0; } probe($u);`, false, false},
		{"array-only parameter in a try without catch", `$u=scandir($a); try { $n = count($u); } finally { echo 1; } probe($u);`, true, true},
		{"unresolved callee", `$u=scandir($a); tally($u); probe($u);`, false, false},
		{"method call", `$u=scandir($a); $o->count($u); probe($u);`, false, false},
		{"extra argument", `$u=scandir($a); $n = count([], 0, $u); probe($u);`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The statements run in a function, after the declarations.
			src := tc.src
			if at := strings.Index(src, "$u=fopen"); !strings.Contains(src, "function f(") && at >= 0 {
				src = src[:at] + "function t($a, $c, $o, $n, $m) { " + src[at:] + " }"
			} else if at := strings.Index(src, "$u=scandir"); at >= 0 {
				src = "function t($a, $c, $o, $n, $m) { " + src + " }"
			}
			e, f := environment(t, src, nil)
			x := argument(callsNamed(f, "probe")[0])
			if got := e.Excludes(x, "false"); got != tc.nonFalse {
				t.Errorf("non-false=%v value=%+v", got, e.Value(x))
			}
			if got := e.Excludes(x, "null"); got != tc.nonNull {
				t.Errorf("non-null=%v value=%+v", got, e.Value(x))
			}
		})
	}
}

// Before PHP 8 a function given false for an array warns and returns.
func TestArrayOnlyParametersProveNothingBeforePHP8(t *testing.T) {
	f := syntax.Parse("test.php", []byte(`<?php $u=scandir($a); if (count($u) > 0) { probe($u); }`), syntax.Options{Version: phpversion.PHP74})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	e := New(f, infer.NewEnv(f, names.New(f), ix, phpversion.PHP74), nil)
	if e.Excludes(argument(callsNamed(f, "probe")[0]), "false") {
		t.Fatal("PHP 7 count() accepted as a proof")
	}
}

// A callable every path of which leaves does not return; the summaries carry
// that across functions and files, for callables no subclass can replace.
func TestCallablesThatAlwaysLeave(t *testing.T) {
	const helpers = `function fail($m) { throw new E($m); }
		function stop($m) { if ($m) { exit(1); } fail($m); }
		function maybe($m) { if ($m) { throw new E($m); } }
		function gen($m) { yield 1; throw new E($m); }
		function loud($m) { echo $m; }
		class H {
			private function quit() { throw new E; }
			final public static function bail() { exit(1); }
			public function open() { throw new E; }
			function a($a) { $u=fopen($a,'r'); if ($u === false) { $this->quit(); } probe($u); }
			function b($a) { $u=fopen($a,'r'); if ($u === false) { self::bail(); } probe($u); }
			function c($a) { $u=fopen($a,'r'); if ($u === false) { $this->open(); } probe($u); }
		}
		function d($a) { $u=fopen($a,'r'); if ($u === false) { fail('x'); } probe($u); }
		function e($a) { $u=fopen($a,'r'); if ($u === false) { stop('x'); } probe($u); }
		function f($a) { $u=fopen($a,'r'); if ($u === false) { maybe('x'); } probe($u); }
		function g($a) { $u=fopen($a,'r'); if ($u === false) { gen('x'); } probe($u); }
		function h($a) { $u=fopen($a,'r'); if ($u === false) { loud('x'); } probe($u); }
		function i($a) { $u=fopen($a,'r'); $u !== false or fail('x'); probe($u); }`
	e, f := environment(t, helpers, nil)
	snapshot := NewSnapshot()
	for range 3 {
		snapshot = NewSnapshot(Extract(f, e.types.Index, e.types.PHP, snapshot))
	}
	for name, never := range map[string]bool{"fail": true, "stop": true, "maybe": false, "gen": false, "loud": false, "h::quit": true, "h::bail": true, "h::open": true, "d": false} {
		if got := snapshot.files["test.php"].Summaries[name].Never; got != never {
			t.Errorf("%s never=%v", name, got)
		}
	}
	e = New(f, e.types, snapshot)
	want := []bool{true, true, false, true, true, false, false, false, true}
	for i, call := range callsNamed(f, "probe") {
		if got := e.Excludes(argument(call), "false"); got != want[i] {
			t.Errorf("probe %d: non-false=%v", i, got)
		}
	}
	// Without the summaries nothing is known about the helpers.
	e, f = environment(t, helpers, nil)
	for i, call := range callsNamed(f, "probe") {
		if e.Excludes(argument(call), "false") {
			t.Errorf("probe %d guarded without summaries", i)
		}
	}
}

func TestCompoundOnly(t *testing.T) {
	for declared, want := range map[string][2]bool{
		"": {false, false}, "array": {true, false}, "?array": {true, true}, "array|null": {true, true}, "Countable|array": {true, false},
		"iterable": {true, false}, "callable": {true, false}, "object": {true, false}, "array|string": {false, false}, "mixed": {false, false},
		"array|false": {false, false}, "string": {false, false}, `\Traversable|array`: {true, false},
	} {
		if compound, nullable := compoundOnly(declared); compound != want[0] || nullable != want[1] {
			t.Errorf("compoundOnly(%q)=%v,%v", declared, compound, nullable)
		}
	}
}
