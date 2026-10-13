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
		// `continue` in a `for` loop runs the step, not the end of the body.
		{"for with a step, continue before the guard", u + `for ($i = 0; $i < 3; $i++) { if ($c) { continue; } if ($u === false) { return; } probe($u); }`, true, false},
		{"for with a step, guard continues", `for ($i = 0; $i < 3; $i++) { $u=fopen($a,'r'); if ($u === false) { continue; } probe($u); }`, true, false},
		{"for with a step, finally in the body", u + `for ($i = 0; $i < 3; $i++) { try { if ($c) { continue; } } finally { echo 1; } if ($u === false) { return; } probe($u); }`, true, false},
		{"for with a step, continue from an inner loop", u + `for ($i = 0; $i < 3; $i++) { foreach ($n as $x) { if ($x) { continue 2; } } if ($u === false) { return; } probe($u); }`, true, false},
		{"for with a step, rewritten before continue", u + `if ($u === false) { return; } for ($i = 0; $i < 3; $i++) { if ($c) { $u = fopen($a,'w'); continue; } probe($u); }`, false, false},
		// A jump over several levels out of a try runs its finally block.
		{"continue 2 out of a try with finally", u + `foreach ($n as $i) { foreach ($n as $j) { try { if ($c) { continue 2; } } finally { echo 1; } } } probe($u);`, false, false},
		{"continue 2 out of a try with finally, guarded before", u + `if ($u === false) { return; } foreach ($n as $i) { foreach ($n as $j) { try { if ($c) { continue 2; } } finally { echo 1; } } } probe($u);`, true, false},
		{"break 2 out of a try with a finally that reassigns", u + `if ($u === false) { return; } foreach ($n as $i) { foreach ($n as $j) { try { if ($c) { break 2; } } finally { $u = fopen($a,'w'); } } } probe($u);`, false, false},
		{"break 2 in a closure inside a try with finally", u + `if ($u === false) { return; } foreach ($n as $i) { try { $g = function () { foreach ([1] as $x) { foreach ([2] as $y) { break 2; } } }; } finally { echo 1; } } probe($u);`, true, false},
		{"break 1 out of a try with finally, outside any loop level above", u + `if ($u === false) { return; } foreach ($n as $i) { try { if ($c) { break 1; } } finally { echo 1; } } probe($u);`, true, false},
		// A loop that always repeats is left through a jump only.
		{"failing branch loops for ever", u + `if ($u === false) { for (;;) { } } probe($u);`, true, false},
		{"failing branch loops while true", u + `if ($u === false) { while (true) { echo 1; } } probe($u);`, true, false},
		{"failing branch loops, do-while true", u + `if ($u === false) { do { echo 1; } while (TRUE); } probe($u);`, true, false},
		{"failing branch loop can break", u + `if ($u === false) { while (true) { if ($c) { break; } } } probe($u);`, false, false},
		{"failing branch loops on a condition", u + `if ($u === false) { while ($c) { echo 1; } } probe($u);`, false, false},
		// A declare block runs its statements in place.
		{"exit in a declare block", u + `if ($u === false) { declare(ticks=1) { return; } } probe($u);`, true, false},
		{"declare block that continues", u + `if ($u === false) { declare(ticks=1) { echo 1; } } probe($u);`, false, false},
		{"declare statement", u + `declare(ticks=1); if ($u === false) { return; } probe($u);`, true, false},
		// A predicate around an assignment tests the value assigned.
		{"predicate around an assignment", u + `if (!is_resource($u = fopen($a,'w'))) { return; } probe($u);`, true, true},
		{"predicate around an assignment, ternary", u + `$x = is_resource($u = fopen($a,'w')) ? probe($u) : 0;`, true, true},
		{"predicate around a compound assignment", u + `if (!is_string($u .= 'x')) { return; } probe($u);`, false, false},
		// `switch (false)` runs the first case that is not true.
		{"switch false, case leaves", u + `switch (false) { case $u: return; } probe($u);`, true, true},
		{"switch false, case body", u + `switch (false) { case $u: probe($u); }`, false, false},
		{"switch false, failed comparison", u + `switch (false) { case $u === false: probe($u); }`, true, false},
		{"switch false, later case", u + `switch (false) { case $u: break; case $c: probe($u); }`, true, true},
		// A typed parameter of a project function does not take null.
		{"typed project parameter", `function take(string $s) {} $u=scandir($a); take($u); probe($u);`, false, true},
		{"typed project parameter, union", `function take(array|false $s) {} $u=scandir($a); take($u); probe($u);`, false, true},
		{"nullable project parameter", `function take(?string $s) {} $u=scandir($a); take($u); probe($u);`, false, false},
		{"project parameter admitting null", `function take(string|null $s) {} $u=scandir($a); take($u); probe($u);`, false, false},
		{"mixed project parameter", `function take(mixed $s) {} $u=scandir($a); take($u); probe($u);`, false, false},
		{"untyped project parameter", `function take($s) {} $u=scandir($a); take($u); probe($u);`, false, false},
		{"optional project parameter", `function take(string $s = null) {} $u=scandir($a); take($u); probe($u);`, false, false},
		{"reference project parameter", `function take(string &$s) {} $u=scandir($a); take($u); probe($u);`, false, false},
		{"typed project parameter in a try", `function take(string $s) {} $u=scandir($a); try { take($u); } catch (TypeError $e) {} probe($u);`, false, false},
		{"typed static method parameter", `class K { static function take(string $s) {} } $u=scandir($a); K::take($u); probe($u);`, false, true},
		{"typed parameter of a method bound late", `class K { static function take(string $s) {} function f($a) { $u=fopen($a,'r'); static::take($u); probe($u); } }`, false, false},
		{"typed parameter of a method called on an object", `class K { function take(string $s) {} } function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"static method of a PHP class", `$u=scandir($a); $d = DateTime::createFromFormat('Y', $u); $x = Closure::fromCallable($u); probe($u);`, false, false},
		{"typed project parameter, variable assigned by a later argument", `function take(string $s, $t) {} $u=scandir($a); take($u, $u = scandir($n)); probe($u);`, false, false},
		{"typed project parameter, function declared twice", `if ($a) { function take(string $s) {} } else { function take($s) {} } $u=scandir($a); take($u); probe($u);`, false, false},
		{"typed parameter of a private static method called from another class", `class K { private static function take(string $s) {} static function __callStatic($n, $a) {} } class L { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } }`, false, false},
		{"typed parameter of a private static method called from outside", `class K { private static function take(string $s) {} static function __callStatic($n, $a) {} } $u=scandir($a); K::take($u); probe($u);`, false, false},
		{"typed parameter of a private static method called inside", `class K { private static function take(string $s) {} function f($a) { $u=fopen($a,'r'); self::take($u); probe($u); } }`, false, true},
		{"typed parameter of a static method of a class declared twice", `if ($a) { class K { static function take(string $s) {} } } else { class K { static function take($s) {} } } $u=scandir($a); K::take($u); probe($u);`, false, false},
		{"typed parameter of an unknown static method", `class K {} $u=scandir($a); K::take($u); probe($u);`, false, false},
		// Positions PHP may skip prove nothing.
		{"null coalescing assignment", `$u=scandir($a); $c ??= count($u); probe($u);`, false, false},
		{"null-safe call argument", `$u=scandir($a); $o?->add(count($u)); probe($u);`, false, false},
		{"call chained after a null-safe one", `$u=scandir($a); $o?->items(array_keys($u))->go(); probe($u);`, false, false},
		{"call chained after a null-safe property", `$u=scandir($a); $o?->p->add(count($u)); probe($u);`, false, false},
		{"offset of a null-safe property", `$u=scandir($a); $x = $o?->list[count($u)]; probe($u);`, false, false},
		{"static call chained after a null-safe property", `$u=scandir($a); $x = $o?->p::make(count($u)); probe($u);`, false, false},
		{"static call on a property", `$u=scandir($a); $x = $o->p::make(count($u)); $n = count($u); probe($u);`, true, true},
		{"dynamic property name", `$u=scandir($a); $x = $o->{name(count($u))}; probe($u);`, true, true},
		{"dynamic method name", `$u=scandir($a); $x = $o->{name(count($u))}(); probe($u);`, true, true},
		{"dynamic null-safe method name", `$u=scandir($a); $x = $o?->{name(count($u))}(); probe($u);`, false, false},
		{"dynamic static method name", `$u=scandir($a); $x = K::{name(count($u))}(); probe($u);`, true, true},
		{"dynamic static method name after a null-safe link", `$u=scandir($a); $x = $o?->k::{name(count($u))}(); probe($u);`, false, false},
		{"dynamic class of an object", `$u=scandir($a); $x = new (name(count($u)))(); probe($u);`, true, true},
		{"offset of a written property", `$u=scandir($a); $o->seen[count($u)] = 1; probe($u);`, true, true},
		{"key of a destructuring", `$u=scandir($a); [count($u) => $x] = $n; probe($u);`, true, true},
		{"offset of a loop target", `$u=scandir($a); foreach ($n as $o->seen[count($u)] => $m[count($u)]) { probe($u); }`, false, false},
		{"call as the base of a written element", `$u=scandir($a); $o->bag(count($u))[0] = 1; probe($u);`, true, true},
		{"key of a destructuring, after its value", `$u=scandir($a); [count($u) => $x] = [$u = scandir($n)]; probe($u);`, true, true},
		{"offset of an unset element", `$u=scandir($a); unset($o->seen[count($u)], $m[0][count($u)], K::$t[count($u)]); probe($u);`, true, true},
		{"dynamic variable written", `$u=scandir($a); ${name(count($u))} = 1; probe($u);`, false, false},
		{"dynamic static property class", `$u=scandir($a); ($o->k)::$t[count($u)] = 1; probe($u);`, true, true},
		{"dynamic null-safe property name", `$u=scandir($a); $x = $o?->{name(count($u))}; probe($u);`, false, false},
		{"never static call chained after a null-safe property", `class H { static function fail(): never { throw new E; } } function f($a, $o) { $u=fopen($a,'r'); if ($u === false) { $o?->h::fail(); } probe($u); }`, false, false},
		// A class using a trait can replace its methods: only an abstract one binds it.
		{"declared never trait method", `trait Tr { function fail(): never { throw new E; } function f($a) { $u=fopen($a,'r'); if ($u === false) { $this->fail(); } probe($u); } }`, false, false},
		{"declared never static trait method", `trait Tr { static function fail(): never { throw new E; } function f($a) { $u=fopen($a,'r'); if ($u === false) { self::fail(); } probe($u); } }`, false, false},
		{"declared never static trait method, late bound", `trait Tr { static function fail(): never { throw new E; } function f($a) { $u=fopen($a,'r'); if ($u === false) { static::fail(); } probe($u); } }`, false, false},
		{"abstract never trait method", `trait Tr { abstract function fail(): never; function f($a) { $u=fopen($a,'r'); if ($u === false) { $this->fail(); } probe($u); } }`, true, false},
		{"never trait method called in the class using it", `trait Tr { function fail(): never { throw new E; } } class C { use Tr; function f($a) { $u=fopen($a,'r'); if ($u === false) { $this->fail(); } probe($u); } }`, true, false},
		{"offset of a static property after a null-safe link", `$u=scandir($a); $x = $o?->q::$tbl[count($u)]; probe($u);`, false, false},
		{"call chained after a static call after a null-safe link", `$u=scandir($a); $x = $o?->q::make()->r(count($u)); probe($u);`, false, false},
		{"offset of a class constant after a null-safe link", `$u=scandir($a); $x = $o?->q::C[count($u)]; probe($u);`, false, false},
		{"offset of a static property", `$u=scandir($a); $x = Reg::$tbl[count($u)]; probe($u);`, true, true},
		{"returning trait method", `trait Tr { private function fail() { throw new E; } function f($a) { $u=fopen($a,'r'); if ($u === false) { $this->fail(); } probe($u); } }`, false, false},
		{"offset of a property", `$u=scandir($a); $x = $o->list[count($u)]; probe($u);`, true, true},
		{"method call argument", `$u=scandir($a); $o->p->add(count($u)); probe($u);`, true, true},
		{"null-safe never method", `class T { function fail(): never { throw new E; } } function f($a, ?T $b) { $u=fopen($a,'r'); if ($u === false) { $b?->fail(); } probe($u); }`, false, false},
		{"never function made a closure", `function fail(): never { throw new E; } ` + u + `if ($u === false) { fail(...); } probe($u);`, false, false},
		{"never method made a closure", `class H { static function fail(): never { throw new E; } } ` + u + `if ($u === false) { H::fail(...); } probe($u);`, false, false},
		{"method call", `$u=scandir($a); $o->count($u); probe($u);`, false, false},
		{"extra argument", `$u=scandir($a); $n = count([], 0, $u); probe($u);`, false, false},
		// A switch compares loosely: a subject that missed `false` or `null`
		// is neither, one that matched `true` is neither.
		{"switch false label leaves", u + `switch ($u) { case false: return; } probe($u);`, true, true},
		{"switch null label leaves", u + `switch ($u) { case NULL: return; } probe($u);`, true, true},
		{"switch default after a false label", u + `switch ($u) { case false: break; default: probe($u); }`, true, true},
		{"switch case below a false label", u + `switch ($u) { case false: break; case 5: probe($u); }`, true, true},
		{"switch true label", u + `switch ($u) { case true: probe($u); }`, true, true},
		{"switch case above a false label", u + `switch ($u) { case 5: probe($u); break; case false: break; }`, false, false},
		{"switch false label falls through", u + `switch ($u) { case false: case 5: probe($u); }`, false, false},
		{"switch false label continues", u + `switch ($u) { case false: break; } probe($u);`, false, false},
		{"switch true label falls out", u + `switch ($u) { case true: break; } probe($u);`, false, false},
		{"switch label expression beside a false label", u + `switch ($u) { case change($u): break; case false: return; } probe($u);`, false, false},
		{"switch subject expression", u + `switch (!$u) { case false: return; } probe($u);`, false, false},
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

// What a skipped position would have assigned is not known to be there.
func TestSkippedPositionsAssignNothingForSure(t *testing.T) {
	for _, src := range []string{
		`function f($x){ $n=1; $x ??= ($n = $_GET['a']); probe($n); }`,
		`function f($o){ $n=1; $o?->m($n = -1); probe($n); }`,
	} {
		e, f := environment(t, src, nil)
		if _, resolved := e.Resolve(argument(callsNamed(f, "probe")[0])); resolved {
			t.Errorf("%s: value resolved", src)
		}
	}
	e, f := environment(t, `function f($x){ $n=1; $x ??= ($n = $_GET['a']); probe($n); }`, nil)
	if !e.Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
		t.Error("request data of a skipped position lost")
	}
}

// A body is only the one a call runs when the callable is declared
// unconditionally and outside a trait.
func TestBodiesThatMayNotBeTheOnesCalled(t *testing.T) {
	e, f := environment(t, `namespace N { function nf() { throw new \E; }
		if (!function_exists('N\\cond')) { function cond() { throw new \E; } }
		trait Tr { private function tq() { throw new \E; } }
		final class C { private function cq() { throw new \E; } }
		if (PHP_VERSION_ID > 1) { final class D { private function dq() { throw new \E; } } } }`, nil)
	sums := Extract(f, e.types.Index, e.types.PHP, nil).Summaries
	for name, never := range map[string]bool{`n\nf`: true, `n\cond`: false, `n\tr::tq`: false, `n\c::cq`: true, `n\d::dq`: false} {
		if sum, ok := sums[name]; !ok || sum.Never != never {
			t.Errorf("%s never=%v declared=%v (%v)", name, sum.Never, ok, len(sums))
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

// A short-circuit operation whose right operand assigns and tests a value
// leaves, where either operand decided, a value that passed the test or one
// the skipped operand left in place. A later test of what the left operand
// read, while those values are in place, tells which.
func TestLaterTestOfTheLeftOperandSettlesASkippedAssignment(t *testing.T) {
	const retry = `do { $n = next($l); } while ($n !== false && ($u = parse_url($n)) === false); `
	for _, tc := range []struct {
		name, src         string
		nonFalse, nonNull bool
	}{
		{"retry until success or end of input", retry + `if ($n === false) { return; } probe($u);`, true, false},
		{"retry, same test", retry + `if ($n !== false) { probe($u); }`, true, false},
		{"retry, negated test", retry + `if (!($n !== false)) { return; } probe($u);`, true, false},
		{"retry, wrong outcome", retry + `if ($n !== false) { return; } probe($u);`, false, false},
		{"retry, no later test", retry + `probe($u);`, false, false},
		{"retry, test variable rewritten", retry + `$n = next($l); if ($n === false) { return; } probe($u);`, false, false},
		{"retry, test variable passed by reference", retry + `change($n); if ($n === false) { return; } probe($u);`, false, false},
		{"retry, test variable rewritten on one path", retry + `if ($c) { $n = next($l); } if ($n === false) { return; } probe($u);`, false, false},
		{"retry, another variable tested", retry + `if ($m === false) { return; } probe($u);`, false, false},
		{"retry, another comparison", retry + `if ($n === null) { return; } probe($u);`, false, false},
		{"retry, loose test of a strict operand", retry + `if ($n == false) { return; } probe($u);`, false, false},
		{"retry, constant first", retry + `if (false === $n) { return; } probe($u);`, false, false},
		{"retry, relational test", retry + `if ($n > 0) { probe($u); }`, false, false},
		{"retry, test of an element", retry + `if ($n['k'] === false) { return; } probe($u);`, false, false},
		{"retry, test of two literals", retry + `if (1 === 2) { return; } probe($u);`, false, false},
		{"retry, value copied before the test", retry + `$w = $u; if ($n === false) { return; } probe($w);`, false, false},
		{"retry, value replaced before the test", retry + `$u = parse_url($a); if ($n === false) { return; } probe($u);`, false, false},
		{"retry, test stored in a flag", retry + `$done = $n === false; if ($done) { return; } probe($u);`, true, false},
		{"loose operand", `do { $n = next($l); } while ($n != false && ($u = parse_url($n)) === false); if ($n == false) { return; } probe($u);`, false, false},
		{"loose operand, object changed between", `$u = parse_url($a); if ($c == $m && ($u = parse_url($o)) === false) { return; } $c->v = 2; if ($c == $m) { probe($u); }`, false, false},
		{"literal operand", `do { $n = next($l); } while ($n !== 'end' && ($u = parse_url($n)) === false); if ($n === 'end') { return; } probe($u);`, true, false},
		{"another literal", `do { $n = next($l); } while ($n !== 'end' && ($u = parse_url($n)) === false); if ($n === 'END') { return; } probe($u);`, false, false},
		{"constant operand", `do { $n = next($l); } while ($n !== DONE && ($u = parse_url($n)) === false); if ($n === DONE) { return; } probe($u);`, true, false},
		{"constant operand, global name", `do { $n = next($l); } while ($n !== DONE && ($u = parse_url($n)) === false); if ($n === \DONE) { return; } probe($u);`, false, false},
		{"constant against a literal", `do { $n = next($l); } while ($n !== DONE && ($u = parse_url($n)) === false); if ($n === 'DONE') { return; } probe($u);`, false, false},
		{"constant-first operand", `do { $n = next($l); } while (false !== $n && ($u = parse_url($n)) === false); if (false === $n) { return; } probe($u);`, true, false},
		{"two variables", `do { $n = next($l); } while ($n !== $m && ($u = parse_url($n)) === false); if ($n === $m) { return; } probe($u);`, true, false},
		{"two variables, one rewritten", `do { $n = next($l); } while ($n !== $m && ($u = parse_url($n)) === false); $m = 1; if ($n === $m) { return; } probe($u);`, false, false},
		{"two variables, one unknown", `do { $n = next($l); } while ($zz !== $n && ($u = parse_url($n)) === false); if ($zz === $n) { return; } probe($u);`, false, false},
		{"null test", `do { $n = next($l); } while ($n !== false && ($u = parse_url($n)) === null); if ($n === false) { return; } probe($u);`, false, true},
		{"truth test", `do { $n = next($l); } while ($n !== false && !($u = parse_url($n))); if ($n === false) { return; } probe($u);`, true, true},
		{"skipped when true", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { if ($c) { continue; } probe($u); } }`, true, false},
		{"skipped when true, wrong outcome", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { if (!$c) { continue; } probe($u); } }`, false, false},
		{"skipped when true, flag flipped", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { probe($u); } $c = !$c; }`, false, false},
		{"negated left operand", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } if ($c) { continue; } probe($u); }`, true, false},
		{"unknown left operand", `foreach ($l as $n) { if ($zz || ($u = parse_url($n)) !== false) { if ($zz) { continue; } probe($u); } }`, false, false},
		{"left operand with an effect", `foreach ($l as $n) { if (next($l) || ($u = parse_url($n)) !== false) { if (next($l)) { continue; } probe($u); } }`, false, false},
		{"guarded value skipped over", `$u = parse_url($a); foreach ($l as $x) { $n = next($l); if ($n !== false) { probe($u); } if ($n !== false && ($u = parse_url($n)) === false) { break; } }`, false, false},
		{"two operations", `$u = parse_url($o); if ($c) { if (!($m || ($u = parse_url($a)) !== false)) { return; } } else { if (!($m || ($u = parse_url($n)) !== false)) { return; } } if ($m) { return; } probe($u);`, false, false},
		// The value a skipped operand leaves in place may itself be known:
		// a variable nothing assigned is null, not false.
		{"skipped with nothing assigned", `foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, true, false},
		{"skipped with nothing assigned, either branch", `if ($c) { if (!($m || ($u = parse_url($a)) !== false)) { return; } } else { if (!($m || ($u = parse_url($n)) !== false)) { return; } } probe($u);`, true, false},
		{"skipped with a guarded value in place", `$u = parse_url($a); if ($u === false) { return; } foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, true, false},
		{"skipped with a null-guarded value in place", `$u = parse_url($a); if ($u === null) { return; } if ($c && ($u = parse_url($o)) === null) { return; } probe($u);`, false, true},
		{"skipped with an unguarded value in place", `$u = parse_url($a); foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"skipped with a failed value of an earlier turn", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { probe($u); } $c = !$c; }`, false, false},
		{"skipped in a closure that captures", `$g = function () use ($u, $l, $c) { foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); } };`, false, false},
		{"skipped in a closure that captures nothing", `$g = function ($l, $c) { foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); } };`, true, false},
		{"skipped in an arrow function", `$g = fn () => ($c && ($u = parse_url($a)) === false) ? 0 : probe($u);`, false, false},
		{"skipped after an import", `extract($o); foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"skipped after a dynamic write", `$$m = false; foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"skipped after a dynamic name was passed on", `known($$m); foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"skipped after a reference to a dynamic name", `$r = &${'u'}; $r = parse_url($a); foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"skipped after an assignment inside a written offset", `$o->p[$u = parse_url($a)] = 1; foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"skipped after a dynamic write on one path", `if ($a) { $$m = false; } foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { break; } probe($u); }`, false, false},
		{"assigned and guarded on one path", `if ($c) { $u = parse_url($a); if ($u === false) { return; } } probe($u);`, true, false},
		{"assigned and guarded on the other path", `if ($c) { echo 1; } else { $u = parse_url($a); if ($u === false) { return; } } probe($u);`, true, false},
		{"assigned and guarded in a loop left before", `do { if ($c) { break; } $u = parse_url($a); if ($u === false) { return; } } while ($m); probe($u);`, true, false},
		{"assigned and guarded in a try", `try { $u = parse_url($a); if ($u === false) { return; } risky(); } catch (E $e) { } probe($u);`, false, false},
		{"assigned and guarded in one arm", `$x = match (true) { $c => ($u = parse_url($a)) === false ? exit() : 1, default => 0 }; probe($u);`, false, false},
		{"static variable with two producers across a call", `static $u; $u = parse_url($a); if ($c) { $u = scandir($a); } if ($u === false) { return; } again(); probe($u);`, false, false},
		{"assigned on one path, unguarded", `if ($c) { $u = parse_url($a); } probe($u);`, false, false},
		// The truth of a variable can change without a write to it.
		{"property of the tested variable written", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $c->v = 2; if ($c) { probe($u); }`, false, false},
		{"element of a property of the tested variable appended", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $c->v[] = 2; if ($c) { probe($u); }`, false, false},
		{"method called on the tested variable", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $c->addChild('x'); if ($c) { probe($u); }`, false, false},
		{"method called on a property of the tested variable", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $c->list->add('x'); if ($c) { probe($u); }`, false, false},
		{"property of the tested variable passed on", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } fill($c->v); if ($c) { probe($u); }`, false, false},
		{"reference taken inside the tested variable", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $r = &$c->v; if ($c) { probe($u); }`, false, false},
		{"property of the tested variable read", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $x = $c->v; if ($c) { probe($u); }`, true, false},
		{"constants spelled alike", `$u = parse_url($a); if ($c === done && ($u = parse_url($o)) === false) { return; } if ($c === DONE) { probe($u); }`, false, false},
		{"null spelled otherwise", `$u = parse_url($a); if ($c !== NULL && ($u = parse_url($o)) === false) { return; } if ($c !== \null) { probe($u); }`, true, false},
		// The truth of an object can change without a write to the variable.
		{"truth of an object", `$x = simplexml_load_string($a); $u = parse_url($a); if ($x && ($u = parse_url($o)) === false) { return; } $x->addChild('x'); if ($x) { probe($u); }`, false, false},
		// A call that enters the function again assigns its static variables.
		{"static variable across a call into the project", `static $u; $u = parse_url($a); if ($u === false) { return; } $o->again(); probe($u);`, false, false},
		{"static variable across a constructor of the project", `static $u; $u = parse_url($a); if ($u === false) { return; } $x = new Again(); probe($u);`, false, false},
		{"static variable across a PHP function", `static $u; $u = parse_url($a); if ($u === false) { return; } echo strlen($a); probe($u);`, true, false},
		{"static variable guarded after the call", `static $u; $u = parse_url($a); $o->again(); if ($u === false) { return; } probe($u);`, true, false},
		{"static variable never assigned", `static $s; $u = parse_url($a); if ($s && ($u = parse_url($o)) === false) { return; } again(); if ($s) { probe($u); }`, true, false},
		// A static variable that is never false where the function calls out
		// or ends is not false after a call either.
		{"static variable read only", `static $u; if ($u === false) { return; } $x = [$l, $u . 'x']; $y = $u['k']; again(); probe($u);`, true, false},
		{"static variable stepped", `static $u; if ($u === false) { return; } again(); probe($u); $u++;`, false, false},
		{"static variable element written", `static $u; if ($u === false) { return; } again(); probe($u); $u['k']['j'] = 1;`, false, false},
		{"static variable passed on", `static $u; if ($u === false) { return; } again(); probe($u); fill($u);`, false, false},
		{"static variable bound by reference", `static $u; if ($u === false) { return; } again(); probe($u); $r = &$u;`, false, false},
		{"static variable copied and passed to a call", `static $u; if ($u === false) { return; } $w = $u; again(); probe($u);`, true, false},
		{"static cache with a fallback", `static $u; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, true, false},
		{"static cache that throws on failure", `static $u; if ($u === null) { $u = parse_url($a); if ($u === false) { throw new E(); } } again(); probe($u);`, false, false},
		{"static cache that returns on failure", `static $u; if ($u === null) { $u = parse_url($a); if ($u === false) { return; } } again(); probe($u);`, false, false},
		{"static cache with a division before the fallback", `static $u; if ($u === null) { $u = parse_url($a); $x = 1 % $c; if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache with an output before the fallback", `static $u; if ($u === null) { $u = parse_url($a); echo $c; if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache with a loop before the fallback", `static $u; if ($u === null) { $u = parse_url($a); foreach ($l as $x) { } if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache with an unset before the fallback", `static $u; if ($u === null) { $u = parse_url($a); unset($c); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache with a negation before the fallback", `static $u; if ($u === null) { $u = parse_url($a); $x = !($u === null) && true ? [1, 'a' => 2] : null; if ($u === false) { $u = []; } } again(); probe($u);`, true, false},
		{"static cache with a call before the fallback", `static $u; if ($u === null) { $u = parse_url($a); note(); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache starting false", `static $u = false; if ($u === false) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache starting from a constant", `static $u = START; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache starting from an array", `static $u = []; if (!$u) { $u = parse_url($a); if ($u === false) { $u = ['x']; } } again(); probe($u);`, true, false},
		{"static cache starting from an expression", `static $u = -1; if ($u === -1) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache starting null", `static $u = null; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, true, false},
		{"static cache bound by reference", `static $u; $r = &$u; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static variable across a closure made from a call", `static $u; $u = parse_url($a); if ($u === false) { return; } $g = $o->again(...); probe($u);`, true, false},
		{"static test stored before a call", `static $u; $u = parse_url($a); $ok = $u !== false; $o->again(); if (!$ok) { return; } probe($u);`, false, false},
		{"static tested before a call in the same condition", `static $u; $u = parse_url($a); if ($u === false || !$o->again()) { return; } probe($u);`, false, false},
		{"static variable captured by reference", `static $u; if ($u === false) { return; } $g = function () use (&$u) { }; again(); probe($u);`, false, false},
		{"static test variable across a call into the project", `static $s; $s = $c; $u = parse_url($a); if ($s && ($u = parse_url($o)) === false) { return; } again(); if ($s) { probe($u); }`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The probe takes its argument by value: a loop reads it again.
			e, f := environment(t, "function probe($x) {} function t($a, $c, $l, $m, $o) { "+tc.src+" }", nil)
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

// A call to a function that runs no code of the project but its own cannot
// enter its caller again.
func TestQuietCalleesLeaveStaticVariables(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		nonFalse  bool
	}{
		{"function that calls nothing", `function leaf($x) { return 1; }`, true},
		{"function that calls a PHP function", `function leaf($x) { return strlen('a') + PHP_INT_SIZE; }`, true},
		{"function that calls a quiet function", `function inner() { return 1; } function leaf($x) { return inner(); }`, true},
		{"function that converts its argument", `function leaf($x) { return "x$x"; }`, false},
		{"function that reads a property", `function leaf($x) { return $x->p; }`, false},
		{"function that calls into the project", `function leaf($x) { return other($x); }`, false},
		{"function that makes an object", `function leaf($x) { return new Thing($x); }`, false},
		{"function that runs a callback", `function leaf($x) { return array_map($x, []); }`, false},
		{"function declared conditionally", `if (!function_exists('leaf')) { function leaf($x) { return 1; } }`, false},
		{"generator", `function leaf($x) { yield 1; }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := summarized(t, tc.src+` function probe($x) {} function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } leaf(1); probe($u); }`)
			calls := callsNamed(f, "probe")
			if got := e.Excludes(argument(calls[len(calls)-1]), "false"); got != tc.nonFalse {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
	// A method is the one called when no class can replace it.
	for src, want := range map[string]bool{
		`class K { private function leaf() { return 1; } function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $this->leaf(); probe($u); } }`:                                                      true,
		`class K { function leaf() { return 1; } function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $this->leaf(); probe($u); } }`:                                                              false,
		`class K { static function leaf() { return 1; } function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } self::leaf(); probe($u); } }`:                                                        true,
		`class K { static function leaf() { return 1; } function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } static::leaf(); probe($u); } }`:                                                      false,
		`class K { private function leaf() { return 1; } function __call($n, $a) { again(); } } function t($a, K $k) { static $u; $u = parse_url($a); if ($u === false) { return; } $k->leaf(); probe($u); }`:              false,
		`class K { private static function leaf() { return 1; } static function __callStatic($n, $a) { again(); } } function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } K::leaf(); probe($u); }`: false,
	} {
		e, f := summarized(t, `function probe($x) {} `+src)
		if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "false"); got != want {
			t.Errorf("non-false=%v for %s", got, src)
		}
	}
}
