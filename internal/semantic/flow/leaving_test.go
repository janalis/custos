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
		// Loose comparisons with constants false and null equal.
		{"loose empty string", u + `if ($u == '') { return; } probe($u);`, true, true},
		{"loose zero", u + `if ($u == 0) { return; } probe($u);`, true, true},
		{"loose zero float", u + `if (0.0 == $u) { return; } probe($u);`, true, true},
		{"loose zero string", u + `if ($u == '0') { return; } probe($u);`, true, false},
		{"loose other string", u + `if ($u == 'x') { return; } probe($u);`, false, false},
		{"loose hexadecimal", u + `if ($u == 0x0) { return; } probe($u);`, false, false},
		{"loose different from empty", u + `if ($u != '') { probe($u); }`, true, true},
		{"loose true", u + `if ($u == true) { probe($u); }`, true, true},
		{"loose not true", u + `if ($u != true) { return; } probe($u);`, true, true},
		{"loose different from a variable", u + `if ($u != $a) { probe($u); }`, false, false},
		// Calls that never return.
		{"declared never", `function fail(): never { throw new E; } ` + u + `if ($u === false) { fail(); } probe($u);`, true, false},
		{"declared never, or form", `function fail(): never { throw new E; } ` + u + `$u !== false or fail(); probe($u);`, true, false},
		{"declared never method", `class T { function fail(): never { throw new E; } function f($a) { ` + u + `if ($u === false) { $this->fail(); } probe($u); } }`, true, false},
		{"declared never static method", `class T { static function fail(): never { throw new E; } function f($a) { ` + u + `if ($u === false) { self::fail(); } probe($u); } }`, true, false},
		{"declared never method of another class", `class H { static function fail(): never { throw new E; } } ` + u + `if ($u === false) { H::fail(); } probe($u);`, true, false},
		{"declared void", `function fail(): void { error_log('x'); } ` + u + `if ($u === false) { fail(); } probe($u);`, false, false},
		// PHPUnit's own failures, when PHPUnit is not indexed.
		{"unindexed PHPUnit fail", `class T extends \PHPUnit\Framework\TestCase { function f($a) { ` + u + `if ($u === false) { self::fail('x'); } probe($u); } }`, true, false},
		{"unindexed PHPUnit skip", `use PHPUnit\Framework\TestCase; class B extends TestCase {} class T extends B { function f($a) { ` + u + `if ($u === false) { $this->markTestSkipped(); } probe($u); } }`, true, false},
		{"unindexed PHPUnit incomplete", `class T extends \PHPUnit\Framework\TestCase { function f($a) { ` + u + `if ($u === false) { static::markTestIncomplete(); } probe($u); } }`, true, false},
		{"unindexed other parent", `class T extends \Other\TestCase { function f($a) { ` + u + `if ($u === false) { self::fail('x'); } probe($u); } }`, false, false},
		{"unindexed PHPUnit, other method", `class T extends \PHPUnit\Framework\TestCase { function f($a) { ` + u + `if ($u === false) { self::report('x'); } probe($u); } }`, false, false},
		{"unindexed PHPUnit, other receiver", `class T extends \PHPUnit\Framework\TestCase { function f($a, $o) { ` + u + `if ($u === false) { $o->fail('x'); } probe($u); } }`, false, false},
		{"no parent class", `class T { function f($a) { ` + u + `if ($u === false) { self::fail('x'); } probe($u); } }`, false, false},
		{"inheritance cycle", `class A extends B {} class B extends A { function f($a) { ` + u + `if ($u === false) { self::fail('x'); } probe($u); } }`, false, false},
		{"PHPUnit fail outside a class", u + `if ($u === false) { self::fail('x'); } probe($u);`, false, false},
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
		// A goto jumps to its label, before or after it.
		{"goto past the guard", u + `goto done; if ($u === false) { return; } done: probe($u);`, false, false},
		{"goto from the failing branch", u + `if ($u === false) { goto fail; } probe($u); return; fail: return;`, true, false},
		{"goto back to another attempt", u + `retry: $u = fopen($a,'w'); if ($u === false) { goto retry; } probe($u);`, true, false},
		{"goto back after a later write", u + `if ($u === false) { return; } again: probe($u); $u = fopen($a,'w'); goto again;`, false, false},
		{"goto out of two loops", u + `foreach ($n as $x) { foreach ($m as $y) { if ($u === false) { goto out; } probe($u); } } out: return;`, true, false},
		{"goto over a statement in the same block", u + `if ($u === false) { return; } goto done; $u = fopen($a,'w'); done: probe($u);`, true, false},
		{"goto to a label followed by a statement", u + `if ($c) { goto done; } if ($u === false) { return; } $x = 1; done: $x = 2; probe($u);`, false, false},
		{"goto beside a try", u + `if ($u === false) { return; } try { if ($c) { goto done; } } catch (E $e) { } done: probe($u);`, true, false},
		{"goto out of a try whose finally writes", u + `if ($u === false) { return; } try { if ($c) { goto done; } } finally { $u = fopen($a,'w'); } done: probe($u);`, false, false},
		{"goto out of a catch", u + `try { $x = $o->read(); } catch (E $e) { goto done; } finally { $u = fopen($a,'w'); } if ($u === false) { return; } done: probe($u);`, false, false},
		{"goto beside a try in a closure", u + `$g = function () { try { echo 1; } finally { echo 2; } }; if ($u === false) { goto fail; } probe($u); return; fail: return;`, true, false},
		{"goto to a label that is not there", u + `if ($u === false) { return; } if ($c) { goto nowhere; } probe($u);`, false, false},
		{"goto to a label declared twice", u + `if ($u === false) { goto fail; } probe($u); return; fail: return; fail: return;`, true, false},
		// A declare block runs its statements in place.
		{"exit in a declare block", u + `if ($u === false) { declare(ticks=1) { return; } } probe($u);`, true, false},
		{"declare block that continues", u + `if ($u === false) { declare(ticks=1) { echo 1; } } probe($u);`, false, false},
		{"declare statement", u + `declare(ticks=1); if ($u === false) { return; } probe($u);`, true, false},
		// A predicate around an assignment tests the value assigned.
		{"predicate around an assignment", u + `if (!is_resource($u = fopen($a,'w'))) { return; } probe($u);`, true, true},
		{"predicate around an assignment, ternary", u + `$x = is_resource($u = fopen($a,'w')) ? probe($u) : 0;`, true, true},
		{"predicate around a compound assignment", u + `if (!is_string($u .= 'x')) { return; } probe($u);`, false, false},
		// A bitwise operation on a test is zero only when the test failed, and
		// not zero only when it held.
		{"bitwise or of a failure test", `$u=scandir($a); if ($c | $u === false) { return; } probe($u);`, true, false},
		{"bitwise or of an assignment tested in place", `$u=scandir($a); if ($c | ($u = scandir($n)) === false) { return; } probe($u);`, true, false},
		{"bitwise or of two tests", `$u=scandir($a); if (($u === false) | ($u === null) | !$c) { return; } probe($u);`, true, true},
		{"bitwise or, test first", `$u=scandir($a); if (!$u | $c) { return; } probe($u);`, true, true},
		{"bitwise or, wrong outcome", `$u=scandir($a); if ($c | $u !== false) { probe($u); }`, false, false},
		{"bitwise or of a value", `$u=scandir($a); if ($c | $u) { return; } probe($u);`, false, false},
		{"bitwise or, left test then rewritten", `$u=scandir($a); if (($u === false) | ($u = scandir($n))) { return; } probe($u);`, false, false},
		{"bitwise and of a success test", `$u=scandir($a); if ($c & $u !== false) { probe($u); }`, true, false},
		{"bitwise and of a predicate", `$u=scandir($a); if (is_array($u) & isset($c)) { probe($u); }`, true, true},
		{"bitwise and, wrong outcome", `$u=scandir($a); if ($c & $u === false) { return; } probe($u);`, false, false},
		{"bitwise xor of a failure test", `$u=scandir($a); if ($c ^ $u === false) { return; } probe($u);`, false, false},
		// A value found in a list of literals is one of them.
		{"strict search in a list of literals", `$u=scandir($a); if (in_array($u, [1, 'a'], true)) { probe($u); }`, true, true},
		{"strict search of an assignment in a list of literals", `$u=scandir($a); if (in_array($u = scandir($n), [1, 2], true)) { probe($u); }`, true, true},
		{"strict search in a list with a constant", `$u=scandir($a); if (in_array($u, [1, false], true)) { probe($u); }`, false, false},
		{"loose search in a list of literals", `$u=scandir($a); if (in_array($u, [1, 2])) { probe($u); }`, false, false},
		{"strict search that failed", `$u=scandir($a); if (!in_array($u, [1, 2], true)) { probe($u); }`, false, false},
		// `match (true)` runs the first arm one of whose conditions is true:
		// the arms below it, and the default one, saw each condition fail.
		{"match true, after an arm that throws", u + `$x = match (true) { $u === false => throw new Exception('x'), default => 1 }; probe($u);`, true, false},
		{"match true, after an arm that exits", u + `$x = match (true) { $u === false => exit(1), default => 1 }; probe($u);`, true, false},
		{"match true, after an arm that never returns", `function stop(): never { exit; } ` + u + `$x = match (true) { $u === false => stop(), default => 1 }; probe($u);`, true, false},
		{"match true, after an arm that returns", u + `$x = match (true) { $u === false => strlen('a'), default => 1 }; probe($u);`, false, false},
		{"match true, every arm throws", u + `$x = match (true) { $u === false => throw new Exception('x'), default => throw new Exception('y') }; probe($u);`, false, false},
		{"match true, arm of a success test", u + `$x = match (true) { $u !== false => probe($u), default => 0 };`, true, false},
		{"match true, default after a failure test", u + `$x = match (true) { $u === false => 0, default => probe($u) };`, true, false},
		{"match true, default written first", u + `$x = match (true) { default => probe($u), $u === false => 0 };`, true, false},
		{"match true, arm below a failure test", u + `$x = match (true) { $u === false => 0, $c => probe($u), default => 1 };`, true, false},
		{"match true, arm of a truth test", u + `$x = match (true) { $u => probe($u), default => 0 };`, true, true},
		{"match true, default after a truth test", u + `$x = match (true) { $u => 0, default => probe($u) };`, false, false},
		{"match true, default after a negation", u + `$x = match (true) { !$u => 0, default => probe($u) };`, true, true},
		{"match true, arm with two conditions", u + `$x = match (true) { $c, $u !== false => probe($u), default => 0 };`, false, false},
		{"match true, arm with two tests of the value", u + `$x = match (true) { $u !== false, is_resource($u) => probe($u), default => 0 };`, true, false},
		{"match true, arm below an assignment tested in place", u + `$x = match (true) { ($u = fopen($a,'w')) === false => 0, $c => probe($u), default => 1 };`, true, false},
		{"match true, arm below an assignment", u + `if ($u === false) { return; } $x = match (true) { $c === ($u = fopen($a,'w')) => 0, default => probe($u) };`, false, false},
		{"match true, after the match", u + `$x = match (true) { $u === false => 0, default => 1 }; probe($u);`, false, false},
		{"match true, without arms", u + `$x = match (true) { }; probe($u);`, false, false},
		{"match true, arm that fails without a default", u + `$x = match (true) { $u !== false => 1 }; probe($u);`, true, false},
		{"match false, arm of a failure test", u + `$x = match (false) { $u === false => probe($u), default => 0 };`, true, false},
		{"match false, default after a success test", u + `$x = match (false) { $u !== false => 0, default => probe($u) };`, true, false},
		{"match false, default after a truth test", u + `$x = match (false) { $u => 0, default => probe($u) };`, false, false},
		{"match of a value", u + `$x = match ($c) { $u === false => 0, default => probe($u) };`, false, false},
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
		{"typed parameter of a method called on an object", `class K { function take(string $s) {} } function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, true},
		{"typed parameter of a method a class replaces", `class K { function take(string $s) {} } class L extends K { function take($s) {} } function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"typed parameter of a method a class replaces, object made here", `class K { function take(string $s) {} } class L extends K { function take($s) {} } function f($a) { $k = new K; $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, true},
		{"typed parameter of a method a class replaces, object made as the class bound late", `class K { function take(string $s) {} function f($a) { $k = new static; $u=fopen($a,'r'); $k->take($u); probe($u); } } class L extends K { function take($s) {} }`, false, false},
		{"typed parameter of a method a class replaces, object made on one path", `class K { function take(string $s) {} } class L extends K { function take($s) {} } function f($a, K $k) { if ($a) { $k = new K; } $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"typed parameter of a method a class replaces, object behind a reference", `class K { function take(string $s) {} } class L extends K { function take($s) {} } function f($a, &$r) { $k = new K; $r = &$k; $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"typed parameter of a method called on a null-safe link", `class K { function take(string $s) {} } function f($a, ?K $k) { $u=fopen($a,'r'); $k?->take($u); probe($u); }`, false, false},
		{"typed parameter of a method of a class declared twice", `if ($a) { class K { function take(string $s) {} } } else { class K { function take($s) {} } } function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"typed parameter of an inherited method", `class J { function take(string $s) {} } class K extends J {} function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, true},
		{"typed parameter of a method inherited from a class declared twice", `if ($a) { class J { function take(string $s) {} } } else { class J { function take($s) {} } } class K extends J {} function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"typed parameter of a method of a trait", `trait T { function take(string $s) {} } class K { use T; } function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, true},
		{"typed parameter of a method of a PHP class", `function f($a, DateTime $k) { $u=fopen($a,'r'); $k->format($u); probe($u); }`, false, false},
		{"typed parameter of an unknown method", `class K {} function f($a, K $k) { $u=fopen($a,'r'); $k->take($u); probe($u); }`, false, false},
		{"static method of a PHP class", `$u=scandir($a); $d = DateTime::createFromFormat('Y', $u); $x = Closure::fromCallable($u); probe($u);`, false, false},
		{"typed project parameter, variable assigned by a later argument", `function take(string $s, $t) {} $u=scandir($a); take($u, $u = scandir($n)); probe($u);`, false, false},
		{"typed project parameter, function declared twice", `if ($a) { function take(string $s) {} } else { function take($s) {} } $u=scandir($a); take($u); probe($u);`, false, false},
		{"typed parameter of a private static method called from another class", `class K { private static function take(string $s) {} static function __callStatic($n, $a) {} } class L { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } }`, false, false},
		{"typed parameter of a private static method called from outside", `class K { private static function take(string $s) {} static function __callStatic($n, $a) {} } $u=scandir($a); K::take($u); probe($u);`, false, false},
		{"typed parameter of a private static method called inside", `class K { private static function take(string $s) {} function f($a) { $u=fopen($a,'r'); self::take($u); probe($u); } }`, false, true},
		{"typed parameter of a protected static method called from a subclass", `class K { protected static function take(string $s) {} static function __callStatic($n, $a) {} } class M extends K {} class L extends M { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } }`, false, true},
		{"typed parameter of an inherited protected static method", `class K { protected static function take(string $s) {} } class L extends K { function f($a) { $u=fopen($a,'r'); parent::take($u); probe($u); } }`, false, true},
		{"typed parameter of a protected static method called from another class", `class K { protected static function take(string $s) {} static function __callStatic($n, $a) {} } class L { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } }`, false, false},
		{"typed parameter of a protected static method called from its parent", `class J { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } } class K extends J { protected static function take(string $s) {} static function __callStatic($n, $a) {} }`, false, false},
		{"typed parameter of a protected static method called from a subclass declared twice", `class K { protected static function take(string $s) {} static function __callStatic($n, $a) {} } if ($a) { class L extends K { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } } } else { class L { } }`, false, false},
		{"typed parameter of a protected static method called below an unknown class", `class K { protected static function take(string $s) {} static function __callStatic($n, $a) {} } class L extends Missing { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } }`, false, false},
		{"typed parameter of a protected static method called below a class that extends itself", `class K { protected static function take(string $s) {} static function __callStatic($n, $a) {} } class L extends L { function f($a) { $u=fopen($a,'r'); K::take($u); probe($u); } }`, false, false},
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
		{"offset of a loop target", `$u=scandir($a); foreach ($n as $o->seen[count($u)] => $m[count($u)]) { probe($u); }`, true, true},
		{"offset of a loop target, after the loop", `$u=scandir($a); foreach ($n as $o->seen[count($u)] => $m[count($u)]) { } probe($u);`, false, false},
		{"offset of a target assigned when null", `$u=scandir($a); $o->seen[count($u)] ??= 1; probe($u);`, true, true},
		{"value assigned to a target when null", `$u=scandir($a); $o->seen[1] ??= count($u); probe($u);`, false, false},
		{"name of a variable assigned when null", `$u=scandir($a); ${name(count($u))} ??= 1; probe($u);`, false, false},
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
	e := New(f, infer.NewEnv(f, names.New(f), ix, phpversion.PHP74, infer.Options{}), nil, Options{})
	if e.Excludes(argument(callsNamed(f, "probe")[0]), "false") {
		t.Fatal("PHP 7 count() accepted as a proof")
	}
}

// Before PHP 7 a scalar type names a class, and a refused argument raises an
// error a handler can recover from: the call returns.
func TestTypedParametersProveNothingBeforePHP7(t *testing.T) {
	for version, want := range map[phpversion.Version]bool{phpversion.PHP56: false, phpversion.PHP70: true} {
		f := syntax.Parse("test.php", []byte(`<?php function take(array $s) {} function t($a) { $u=scandir($a); take($u); probe($u); }`), syntax.Options{Version: version})
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		e := New(f, infer.NewEnv(f, names.New(f), ix, version, infer.Options{}), nil, Options{})
		if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "null"); got != want {
			t.Errorf("%v: non-null=%v", version, got)
		}
	}
}

// Before PHP 8 a property written to null makes an object of it: what a
// strict comparison decided of the variable no longer stands. A method call
// leaves it the value it is in every version.
func TestPropertyWritesChangeIdentityBeforePHP8(t *testing.T) {
	for src, want := range map[string]bool{
		`function probe($x) {} function t($a, $c, $m, $o) { $u = parse_url($a); if ($c === $m && ($u = parse_url($o)) === false) { return; } $c->v = 2; if ($c === $m) { probe($u); } }`:  false,
		`function probe($x) {} function t($a, $c, $m, $o) { $u = parse_url($a); if ($c === $m && ($u = parse_url($o)) === false) { return; } $c->add(2); if ($c === $m) { probe($u); } }`: true,
	} {
		f := syntax.Parse("test.php", []byte(`<?php `+src), syntax.Options{Version: phpversion.PHP74})
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		e := New(f, infer.NewEnv(f, names.New(f), ix, phpversion.PHP74, infer.Options{}), nil, Options{})
		if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "false"); got != want {
			t.Errorf("non-false=%v for %s", got, src)
		}
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
		snapshot = NewSnapshot(Extract(f, e.types.Index, e.types.PHP, snapshot, Options{}))
	}
	for name, never := range map[string]bool{"fail": true, "stop": true, "maybe": false, "gen": false, "loud": false, "h::quit": true, "h::bail": true, "h::open": true, "d": false} {
		if got := snapshot.files["test.php"].Summaries[name].Never; got != never {
			t.Errorf("%s never=%v", name, got)
		}
	}
	e = New(f, e.types, snapshot, Options{})
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
	sums := Extract(f, e.types.Index, e.types.PHP, nil, Options{}).Summaries
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
		{"predicate around the assignment", `$u = parse_url($a); if ($u || is_array($u = parse_url($o))) { probe($u); }`, true, true},
		{"predicate around the assignment, failure exit", `$u = parse_url($a); if (!$u && !is_array($u = parse_url($o))) { return; } probe($u); `, true, true},
		{"predicate around the assignment, null test", `$u = parse_url($a); if ($u !== false || !is_null($u = parse_url($o))) { probe($u); }`, false, false},
		{"predicate around the assignment, wrong outcome", `$u = parse_url($a); if ($u || is_array($u = parse_url($o))) { return; } probe($u);`, false, false},
		{"predicate of the project around the assignment", `$u = parse_url($a); if ($u || is_fine($u = parse_url($o))) { probe($u); }`, false, false},
		{"predicate around another argument", `$u = parse_url($a); if ($u || strpos($m, $u = parse_url($o))) { probe($u); }`, false, false},
		{"skipped when true", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { if ($c) { continue; } probe($u); } }`, true, false},
		{"skipped when true, wrong outcome", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { if (!$c) { continue; } probe($u); } }`, true, false},
		{"skipped when true, flag flipped", `foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { probe($u); } $c = !$c; }`, false, false},
		{"negated left operand", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } if ($c) { continue; } probe($u); }`, true, false},
		{"unknown left operand", `foreach ($l as $n) { if ($zz || ($u = parse_url($n)) !== false) { if ($zz) { continue; } probe($u); } }`, true, false},
		{"skipped when true, value assigned before the loop", `$u = parse_url($a); foreach ($l as $n) { if ($c || ($u = parse_url($n)) !== false) { if (!$c) { continue; } probe($u); } }`, false, false},
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
		// What the skipped operand leaves in place can depend on the left
		// operand across the turns of a loop: where the flag would skip the
		// assignment, the value an earlier turn left passed the test.
		{"skipped while a flag holds, failure keeps the flag", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); $c = !$c; }`, true, false},
		{"skipped while a flag holds, null test", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === null) { continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, null test of a value in place", `$u = 1; if ($u === null) { return; } foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === null) { continue; } probe($u); $c = !$c; }`, false, true},
		{"skipped while a flag holds, unguarded value before the loop", `$u = parse_url($a); foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, guarded value before the loop", `$u = parse_url($a); if ($u === false) { return; } foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); $c = !$c; }`, true, false},
		{"skipped while a flag holds, failure sets the flag", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { $c = true; continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, failure passes the flag on", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { change($c); continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, failure changes the flag's object", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { $c->v = 1; continue; } probe($u); }`, false, false},
		{"skipped while a flag holds, flag passed on after the read", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); change($c); }`, true, false},
		{"skipped while a flag holds, value replaced after the read", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); $u = parse_url($a); $c = !$c; }`, false, false},
		{"skipped while a flag holds, value replaced then flag kept", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); $u = parse_url($a); }`, false, false},
		{"skipped while a flag holds, value unset after the read", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); unset($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, import after the read", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); extract($o); }`, false, false},
		{"skipped while a flag holds, flag bound to another name", `$r = &$c; foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { $r = 1; continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, flag bound to another name, copy assigned", `$r = &$c; foreach ($l as $n) { if (!$c && ($u = $n) === false) { $r = 1; continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, copy assigned", `foreach ($l as $n) { if (!$c && ($u = $n) === false) { continue; } probe($u); $c = !$c; }`, true, false},
		{"skipped while a flag holds, flag bound later", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { $r = &$c; continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, static flag across a call", `static $s; $s = $c; foreach ($l as $n) { if (!$s && ($u = parse_url($n)) === false) { again(); continue; } probe($u); $s = !$s; }`, false, false},
		{"skipped while a flag holds, flag written by the operand", `foreach ($l as $n) { if (!$c && ($u = parse_url($c = $n)) === false) { continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped while a flag holds, left operand with an effect", `foreach ($l as $n) { if (!next($l) && ($u = parse_url($n)) === false) { continue; } probe($u); }`, false, false},
		{"skipped while an object is true", `$x = simplexml_load_string($a); foreach ($l as $n) { if (!$x && ($u = parse_url($n)) === false) { $x->addChild('x'); continue; } probe($u); }`, false, false},
		{"skipped while a flag holds, in a closure that captures", `$g = function () use ($u, $l, $c) { foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); $c = !$c; } };`, false, false},
		{"skipped while a flag holds, compared flag", `foreach ($l as $n) { if ($c !== 1 && ($u = parse_url($n)) === false) { continue; } probe($u); $c = 1; }`, true, false},
		{"skipped while a flag holds, included file", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } probe($u); include $a; }`, false, false},
		{"skipped while either flag holds", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { continue; } if (!$m && ($w = parse_url($n)) === false) { continue; } probe($u); $c = !$c; $m = !$m; }`, true, false},
		{"skipped while either flag holds, first value replaced", `foreach ($l as $n) { if (!$m && ($w = parse_url($n)) === false) { continue; } if (!$c && ($u = parse_url($n)) === false) { continue; } probe($w); $w = parse_url($a); $u = parse_url($a); $c = !$c; }`, false, false},
		{"skipped while either flag holds, second value replaced", `foreach ($l as $n) { if (!$m && ($w = parse_url($n)) === false) { continue; } if (!$c && ($u = parse_url($n)) === false) { continue; } probe($w); $u = parse_url($a); $c = !$c; $m = !$m; }`, true, false},
		{"skipped while a flag holds, flag made global", `foreach ($l as $n) { if (!$c && ($u = parse_url($n)) === false) { global $c; continue; } probe($u); $c = !$c; }`, false, false},
		{"skipped with a key of an array in place", `foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { return; } probe($u); foreach ((array) $l as $u => $x) { } }`, true, false},
		{"skipped with a key of anything in place", `foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { return; } probe($u); foreach ($l as $u => $x) { } }`, false, false},
		{"skipped with an element in place", `foreach ($l as $n) { if ($c && ($u = parse_url($n)) === false) { return; } probe($u); foreach ($l as $u) { } }`, false, false},
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
		// The identity of a value does not follow what its object holds.
		{"property of an identical variable written", `$u = parse_url($a); if ($c === $m && ($u = parse_url($o)) === false) { return; } $c->v = 2; $m->w[] = 3; if ($c === $m) { probe($u); }`, true, false},
		{"method called on an identical variable", `$u = parse_url($a); if ($c !== $m && ($u = parse_url($o)) === false) { return; } $c->add('x'); if ($c !== $m) { probe($u); }`, true, false},
		{"element of an identical variable written", `$u = parse_url($a); if ($c === $m && ($u = parse_url($o)) === false) { return; } $c['k'] = 2; if ($c === $m) { probe($u); }`, false, false},
		{"reference taken inside an identical variable", `$u = parse_url($a); if ($c === $m && ($u = parse_url($o)) === false) { return; } $r = &$c->v; if ($c === $m) { probe($u); }`, false, false},
		{"property of an identical variable passed on", `$u = parse_url($a); if ($c === $m && ($u = parse_url($o)) === false) { return; } fill($c->v); if ($c === $m) { probe($u); }`, false, false},
		{"property of the tested variable read", `$u = parse_url($a); if ($c && ($u = parse_url($o)) === false) { return; } $x = $c->v; if ($c) { probe($u); }`, true, false},
		{"constants spelled alike", `$u = parse_url($a); if ($c === done && ($u = parse_url($o)) === false) { return; } if ($c === DONE) { probe($u); }`, false, false},
		{"null spelled otherwise", `$u = parse_url($a); if ($c !== NULL && ($u = parse_url($o)) === false) { return; } if ($c !== \null) { probe($u); }`, true, false},
		// The truth of an object can change without a write to the variable.
		{"truth of an object", `$x = simplexml_load_string($a); $u = parse_url($a); if ($x && ($u = parse_url($o)) === false) { return; } $x->addChild('x'); if ($x) { probe($u); }`, false, false},
		// A call that enters the function again assigns its static variables.
		{"static variable across a call into the project", `static $u; $u = parse_url($a); if ($u === false) { return; } $o->again(); probe($u);`, false, false},
		{"static variable across a constructor of the project", `static $u; $u = parse_url($a); if ($u === false) { return; } $x = new Again(); probe($u);`, false, false},
		{"static variable across a PHP function", `static $u; $u = parse_url($a); if ($u === false) { return; } echo strlen('a'), count([]); probe($u);`, true, false},
		{"static variable across a PHP function given what may be an object", `static $u; $u = parse_url($a); if ($u === false) { return; } echo strlen($a); probe($u);`, false, false},
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
		{"static cache starting false", `static $u = false; if ($u === false) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, true, false},
		{"static cache starting false, read before it is filled", `static $u = false; again(); probe($u); $u = parse_url($a); if ($u === false) { $u = []; }`, false, false},
		{"static starting false, never filled on one path", `static $u = false; if ($c) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache starting from a constant", `static $u = START; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static cache starting from an array", `static $u = []; if (!$u) { $u = parse_url($a); if ($u === false) { $u = ['x']; } } again(); probe($u);`, true, false},
		{"static cache starting from a negative number", `static $u = -1; if ($u === -1) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, true, false},
		{"static cache starting from an expression", `static $u = !1; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		{"static declared twice, once starting false", `if ($c) { static $u = false; } else { static $u; } if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u);`, false, false},
		// null and true are not false: a cache emptied at the end stays one.
		{"static cache reset to null", `static $u; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u); $u = null;`, true, false},
		{"static cache reset to true", `static $u; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u); $u = TRUE;`, true, false},
		{"static cache reset to false", `static $u; if ($u === null) { $u = parse_url($a); if ($u === false) { $u = []; } } again(); probe($u); $u = false;`, false, false},
		// Each static variable is assumed on its own.
		{"second static breaks the assumption, first read", `static $u, $w; $u = parse_url($a); if ($u === false) { $u = []; } $w = parse_url($o); if ($w === false) { return; } again(); probe($u);`, true, false},
		{"second static breaks the assumption, second read", `static $u, $w; $u = parse_url($a); if ($u === false) { $u = []; } $w = parse_url($o); if ($w === false) { return; } again(); probe($w);`, false, false},
		{"four statics break the assumption in turn", `static $u, $w, $x, $y, $z; $u = parse_url($a); if ($u === false) { $u = []; } again(); probe($u); $w = false; again(); if ($c) { echo 1; } $x = false; again(); if ($c) { echo 1; } $y = false; again(); if ($c) { echo 1; } $z = false;`, false, false},
		{"three statics break the assumption in turn", `static $u, $w, $x; $u = parse_url($a); if ($u === false) { $u = []; } again(); probe($u); $w = false; again(); if ($c) { echo 1; } $x = false;`, true, false},
		// The statement reached again in a loop reads what the function left.
		{"static statement reached again with false in place", `while (true) { static $u; again(); probe($u); if ($c) { return; } $u = false; }`, false, false},
		{"static statement reached again with a value in place", `while (true) { static $u; again(); probe($u); if ($c) { return; } $u = []; }`, true, false},
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

// Code of the project runs without a call written where an object is
// converted to text, compared with some, iterated or read by offset, where a
// property is read or written, where a class loads. It can enter the
// function again like a call. Copying a value, testing its identity or its
// truth and computing with it run none, nor does anything done with a value
// that is no object.
func TestCodeThatRunsWithoutACallEntersTheFunctionAgain(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kept bool
	}{
		{`echo $o;`, false},
		{`$x = "v$o";`, false},
		{`foreach ($o as $v) { }`, false},
		{`$x = $o->p;`, false},
		{`$x = $o?->p;`, false},
		{`$o->p = 1;`, false},
		{`$o->p['k'] = 1;`, false},
		{`[$o->p] = [1];`, false},
		{`K::$p = 1;`, false},
		{`$x = K::$p;`, false},
		{`$x = $o[0];`, false},
		{`$o[0] = 1;`, false},
		{`$x = isset($o[0]);`, false},
		{`$x = clone $o;`, false},
		{`unset($x);`, true},
		{`$x = 'text'; unset($x);`, true},
		{`unset($o);`, false},
		{`unset($o[0]);`, false},
		{`unset($o->p);`, false},
		// What a destructuring or a reference rebinds is released.
		{`$x = $o; [$x] = [1];`, false},
		{`$x = $o; [[$x]] = [[1]];`, false},
		{`$x = 'text'; [$x] = [1];`, true},
		{`$x = $o; $x = &$l;`, false},
		{`$x = 'text'; $x = &$l;`, true},
		{`$x = count($o);`, false},
		{`$x = strlen($o);`, false},
		{`$x = json_encode($o);`, false},
		{`trigger_error('x');`, false},
		{`$x = class_exists('K');`, false},
		{`$x = Unknown::X;`, false},
		{`$x = $o == 'x';`, false},
		{`$x = 'x' < $o;`, false},
		{`$x = $o <=> $c;`, false},
		{`$x = isset($o->p);`, false},
		{`$x = (string) $o;`, false},
		{`$x = Fiber::suspend();`, false},
		{`$x = $o . '';`, false},
		{`$x = 'a'; $x .= $o;`, false},
		{`$l['k'] .= $o;`, false},
		{`[$x] = $o;`, false},
		{`$x = [...$o];`, false},
		{`$x = max(...$o);`, false},
		{`$x = $$o;`, false},
		{`print $o;`, false},
		{`switch ($o) { case 'x': break; }`, false},
		{`switch ($c) { case $o: break; }`, false},
		{`$x = new DateTime($o);`, false},
		{`$x = new DateTimeZone(...$o);`, false},
		{`$x = implode(',', $l);`, false},
		{`$x = [$o]; sort($x);`, false},
		{`$x = [$o]; echo $x[0];`, false},
		{`$x = $l['a']['b'];`, false},
		{`function g(): \Generator { yield 1; } $g = g(); $g->current();`, false},
		{`$x = new ArrayIterator([]); $x->current();`, true},
		{`class It extends ArrayIterator { function current(): mixed { return again(); } } $x = new ArrayIterator([]); $x->current();`, false},
		{`function it(\Iterator $i) { static $u; $u = parse_url($i); if ($u === false) { return; } $i->next(); probe($u); }`, false},
		{`$x = $o + 1;`, true},
		{`$x = -$o;`, true},
		{`$x = (int) $o;`, true},
		{`$x = strtoupper('a');`, true},
		{`$x = new ArrayObject([]);`, true},
		{`$x = new ArrayObject($o);`, true},
		{`$x = $o->m(...);`, true},
		{`$x = $o;`, true},
		{`$x = $o === $c;`, true},
		{`$x = !$o;`, true},
		{`$x = $o ? 1 : 2;`, true},
		{`$x = $o ?? 1;`, true},
		{`$x = $o && $c;`, true},
		{`$x = $o == null;`, true},
		{`$x = false != $o;`, true},
		{`$x = $o < 3;`, true},
		{`$x = $o > -1.5;`, true},
		{`$x = $o instanceof K;`, true},
		{`$x = is_array($o);`, true},
		{`$x = array_key_exists('k', $o);`, true},
		{`$x = K::class;`, true},
		{`$x = K::class . 'a'; echo K::class, $o::class;`, true},
		{`echo K::NAME;`, false},
		{`$x = [$o, 1];`, true},
		{`$x = [$o]; $n = count($x);`, true},
		{`$x = fn () => $o->p;`, true},
		{`$x = function () use ($o) { echo $o; };`, true},
		{`$n = 1; $x = "v$n" . $n; echo $n, 'a';`, true},
		{`$n = (string) $o; echo $o;`, false},
		{`$n = [1, 'a' => ['b']]; foreach ($n as $k => $v) { echo $k, $v; }`, true},
		{`$n = [1, $o]; foreach ($n as $k => $v) { echo $k; }`, true},
		{`$n = [1, $o]; foreach ($n as $k => $v) { echo $v; }`, false},
		{`$n = explode(',', $a); echo $n[0], count($n), implode('-', $n);`, false},
		{`if (!is_string($a)) { return; } $n = explode(',', $a); echo $n[0] . count($n) . implode('-', $n); sort($n);`, true},
		{`$n = array_merge([1], $l); echo $n[0];`, false},
		{`$n = array_merge([1], [2]); echo $n[0];`, true},
		{`$n = func_get_args(); echo $n[0];`, false},
		{`$n = explode(',', 'a', 2, 3); echo $n[0];`, false},
		{`$x = new ArrayObject([]); $n = $x->getArrayCopy(); echo $n[0];`, false},
		{`$x = new ArrayObject([]); $n = $x->count(); echo $n;`, true},
		{`preg_match('/a/', 'a', $n); echo $n[0]; $n[] = 1; echo $n[1];`, true},
		{`preg_match('/a/', 'a', $n); $n[] = $o; echo $n[1];`, false},
		{`$n = []; $n[] = 1; $n['k'] = 'a'; echo $n['k'];`, true},
		{`$n = 1; $n++; $n += 2; $n .= 'a'; echo $n;`, true},
		{`$n = [1]; $n += $l; echo $n[0];`, false},
		{`$n = $o; $n ??= 1; echo $n;`, false},
		{`$n = (array) $o; echo $n[0];`, false},
		{`$n = (array) 'a'; echo $n[0];`, true},
		{`$n = @$o; echo $n;`, false},
		{`if (is_int($o)) { echo $o; }`, true},
		{`if (is_array($o)) { $x = $o[0]; }`, true},
		{`if (is_array($o)) { echo $o[0]; }`, false},
		{`if ($o === 'a') { echo $o; }`, true},
		{`echo $_GET['a'], $_SESSION['a'] ?? 1, $GLOBALS['a'] ?? 1;`, false},
		{`$x = $GLOBALS['a'] ?? 1; $y = isset($GLOBALS['a']); $GLOBALS['b'] = 1; echo $_GET['a']['b'];`, true},
		// A function PHP calls can make objects.
		{`function it($a, $o) { static $u; $n = array_map(function ($i) { return new K; }, [1]); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, false},
		{`function it($a, $o) { static $u; $n = array_map('make', [1]); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, false},
		{`function it($a, $o) { static $u; $n = array_map($o, [1]); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, false},
		{`function it($a, $o) { static $u; $n = array_map('strtoupper', ['a']); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, true},
		{`function it($a, $o) { static $u; $n = array_map('array_values', [[1]]); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, false},
		{`function it($a, $o) { static $u; $n = array_filter([1, 2]); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, true},
		{`function it($a, $o) { static $u; $n = [1]; usort($n, 'strcmp'); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, true},
		{`function it($a, $o) { static $u; $n = [1]; usort($n, $o); $u = parse_url($a); if ($u === false) { return; } echo $n[0]; probe($u); }`, false},
		// Looking for a value compares each element with it.
		{`if (is_array($o)) { $x = array_keys($o); }`, true},
		{`if (is_array($o)) { $x = array_keys($o, 'x'); }`, false},
		// An element that is taken apart may be an object.
		{`foreach ([[1]] as [$x]) { }`, true},
		{`if (is_array($o)) { foreach ($o as [$x]) { } }`, false},
		{`if (is_array($o)) { [$x] = $o; }`, true},
		{`if (is_array($o)) { [[$x]] = $o; }`, false},
		{`[[$x]] = [[1]];`, true},
		// A value that is replaced may be an object nothing else holds.
		{`$x = new ArrayObject([]); $x = 1;`, false},
		{`$x = [1]; $x = 1; $y = $x; $y = 2;`, true},
		{`$o = null;`, false},
		{`if (is_array($o)) { $o = null; }`, false},
		{`function it($c) { return function () use ($c) { static $u; $u = parse_url($c); if ($u === false) { return; } $x = 1; probe($u); }; }`, false},
		{`try { $n = 1; } catch (Exception $n) { } catch (Error) { }`, true},
		{`try { $n = 1; } catch (Exception $o) { }`, true},
		{`try { work(); } catch (Exception $o) { }`, false},
		// A key is a number or text when an array is iterated.
		{`foreach ($o as $k => $v) { echo $k; }`, false},
		{`foreach ([1] as $k => $v) { echo $k; }`, true},
		// A constant of the project may be an object.
		{`$x = PHP_EOL; echo $x, E_ALL, \PHP_INT_SIZE;`, true},
		{`$x = SOME_CONSTANT; echo $x;`, false},
		{`$x = (object) ['a' => 1]; echo $x;`, false},
		// A declared property of the object itself is read without code.
		{`class P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, true},
		{`class P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this?->p; probe($u); } }`, true},
		{`class P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $y = $this->p['k']; probe($u); } }`, false},
		{`final class P extends Q { public $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } } class Q { }`, true},
		{`class P { public $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		{`class P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->q; probe($u); } }`, false},
		{`class P { private $p = 'a'; function it($a, $n) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->$n; probe($u); } }`, false},
		{`class P { private $p = 'a'; function it($a, P $o) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $o->p; probe($u); } }`, false},
		{`class P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $g = function () { return $this->p; }; $g = fn () => $this->p; probe($u); } }`, true},
		{`class P { private $p = 'a'; function it($a) { $g = function () use ($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); }; } }`, false},
		{`class P { private $p = 'a'; function __get($n) { } function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		{`class P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } } class Q extends P { function __get($n) { } }`, false},
		{`class P extends Q { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		{`class P extends Q { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } } class Q extends P { }`, false},
		{`class P { use Q; private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		{`class P extends Q { function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } } class Q { protected $p = 'a'; }`, false},
		{`class P { private static $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		{`class P { private $p { get => 'a'; } function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		{`trait P { private $p = 'a'; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $this->p; probe($u); } }`, false},
		// A constant of a class that is loaded loads nothing.
		{`class P { const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::C + P::C; probe($u); } }`, true},
		{`class P extends Q { function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = parent::C; probe($u); } } class Q { const C = 'a'; }`, true},
		{`final class P { const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = static::C; probe($u); } }`, true},
		{`class P { final const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = static::C; probe($u); } }`, true},
		{`class P { const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = static::C; probe($u); } }`, false},
		{`class P { const C = K::D; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::C; probe($u); } }`, false},
		{`class P { const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::D; probe($u); } }`, false},
		{`class P { const C = 1; function it($a, $n) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::{$n}; probe($u); } }`, false},
		{`class P { const C = 1; function it($a, $n) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = $n::C; probe($u); } }`, false},
		{`class P { const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = Q::C; probe($u); } } class Q { const C = 1; }`, false},
		{`class P extends Q { const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::C; probe($u); } }`, false},
		{`enum P { case A; const C = 1; function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::C; probe($u); } }`, false},
		{`function it($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = self::C; probe($u); }`, false},
		{`$x = $_SESSION['a'] ?? 1;`, false},
		{`$_SESSION['a'] = 1;`, false},
		{`$n = 1; $x = "a{$o}b$n";`, false},
		{`$n = 1; $x = "a{$n}b$n" . <<<T
			$n
			T;`, true},
		{`$_GET['k'] = $o; echo $_GET['k'];`, false},
		{`$_GET['k'] = 1; echo $_GET['k'];`, true},
		{`switch ($c) { case 1: break; case null: break; }`, true},
		{`switch (true) { case $o: break; }`, true},
		{`$x = match ($o) { 'a' => 1, default => 2 };`, true},
		{`echo $nothing; $x = $nothing[0]; foreach ($nothing as $v) { }`, true},
		// The guarded variable itself is no object.
		{`echo $u['host'], strlen($u['path']); foreach ($u as $k => $v) { echo $k, $v; } $x = "p$u[port]" . $u['user'];`, true},
		{`other(); echo $u['host']; if ($u === false) { return; } echo $u['path'], $u['user'];`, true},
	} {
		t.Run(tc.src, func(t *testing.T) {
			src := tc.src
			if !strings.Contains(src, "function it(") {
				src = `function t($a, $c, $l, $o) { static $u; $u = parse_url($a); if ($u === false) { return; } ` + src + ` probe($u); }`
			}
			e, f := summarized(t, `class K { static $p; } function probe($x) {} `+src)
			calls := callsNamed(f, "probe")
			if got := e.Excludes(argument(calls[len(calls)-1]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
}

// A static variable the function gives nothing but values that are no
// object is none after a call that enters the function again: reading it
// runs no code.
func TestStaticVariablesThatNeverHoldAnObject(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kept bool
	}{
		{`static $u; $u = parse_url($a);`, true},
		{`static $u = []; $u = parse_url($a); $u['k'] = 1; $u += [2]; $n = strlen($u['k']); $x = [$u]; foreach ($u as $v) { } $g = function () use ($u) { };`, true},
		{`static $u; $u = $c ? parse_url($a) : ($u ?: []);`, true},
		{`static $u; $u = parse_url($a); if ($c) { $u = $o; }`, false},
		{`static $u; $u = parse_url($a); if ($c) { $u = make(); }`, false},
		{`static $u; $u = parse_url($a); $u['k'] = $o;`, false},
		{`static $u; $u = parse_url($a); $u += $l;`, false},
		{`static $u; $u = parse_url($a); $u ??= $o;`, false},
		{`static $u; $u = parse_url($a); $u .= $o; $u['k'] .= 'a'; $u++;`, true},
		{`static $u; $u = parse_url($a); if ($c) { global $u; }`, false},
		{`static $u; $u = parse_url($a); $r = &$u;`, false},
		{`static $u; $u = parse_url($a); fill($u);`, false},
		{`static $u; $u = parse_url($a); sort($u);`, false},
		{`static $u; $u = parse_url($a); [$u] = $l;`, false},
		{`static $u; $u = parse_url($a); foreach ($l as $u) { break; }`, false},
		{`static $u; $u = parse_url($a); $g = function () use (&$u) { };`, false},
		{`static $u; $u = parse_url($a); if ($c) { unset($u); }`, false},
		{`static $u; $u = parse_url($a); try { } catch (E $u) { }`, false},
		{`static $u = K::START; $u = parse_url($a);`, false},
		{`static $u; $u = array_merge(parse_url($a), $l);`, false},
		{`static $u; $u = array_merge(parse_url($a), ['k' => 1]);`, true},
		{`static $u; $u = $c ? parse_url($a) : [1, $o];`, false},
		{`static $u; $u = (array) $o;`, false},
		{`static $u; $u = compact('o');`, false},
		{`static $u; $u = parse_url($a) ?? [1];`, true},
		{`static $u; $u = parse_url($a) ?? $l;`, false},
		{`static $u; $u = [1] + parse_url($a);`, true},
		{`static $u; $u = [1] + $l;`, false},
		{`static $u; $u = ($x = parse_url($a));`, true},
		{`static $u; $u = ($x = $o);`, false},
		{`static $u; $u = ($x .= 'a');`, true},
		{`static $u; $u = ($x ??= parse_url($a));`, false},
		{`static $u; $u = parse_url($a); $x = [[$u], 1];`, true},
		{`static $u; $u = parse_url($a); echo count([$u]);`, true},
		{`static $u; $u = parse_url($a); [[$u]] = $l;`, false},
		{`static $u; $u = parse_url($a); list($u) = $l;`, false},
		{`static $u; $u = parse_url($a); foreach ($l as [$u]) { break; }`, false},
		{`static $u; $u = parse_url($a); foreach ($l as [$u] => $x) { break; }`, false},
		{`static $u; $u = parse_url($a); $x = max(...$u);`, false},
		{`static $u; $u = array_merge(...$l);`, false},
		{`static $u; $u = (array) parse_url($a);`, true},
		{`static $u; $u = @$o;`, false},
		// A write the scope does not name may give it any value.
		{`static $u; $u = parse_url($a); $$c = $o;`, false},
		{`static $u; $u = parse_url($a); extract($l);`, false},
		{`static $u; $u = parse_url($a); parse_str($c);`, false},
		{`static $u; $u = parse_url($a); parse_str($c, $x);`, true},
		{`static $u; $u = parse_url($a); eval($c);`, false},
		{`static $u; $u = parse_url($a); include $c;`, false},
		{`static $u; $u = (object) parse_url($a);`, false},
		{`static $u = PHP_INT_SIZE; $u = parse_url($a);`, true},
		{`static $u = SOME_CONSTANT; $u = parse_url($a);`, false},
		{`static $u; $u = parse_url($a); $u = $x++;`, true},
	} {
		t.Run(tc.src, func(t *testing.T) {
			// The read of an element follows a call that enters the function
			// again: it runs code of the project unless the variable is known
			// to hold no object.
			e, f := environment(t, `function probe($x) {} function t($a, $c, $l, $o) { `+tc.src+` again(); if ($u === false) { return; } echo $u['host']; probe($u); }`, nil)
			if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
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
		{"function that copies and tests its argument", `function leaf($x) { $y = $x; if ($y === null || !$y) { return [$y, 2]; } return $y + 1; }`, true},
		{"function that converts text", `function leaf(int $x) { $s = 'a'; return "x$x" . $s . strlen($s); }`, true},
		{"function that returns its argument as text", `function leaf($x): string { return $x; }`, false},
		{"function that returns text", `function leaf($x): string { return 'a'; }`, true},
		{"function that returns its argument", `function leaf($x): ?array { return $x; }`, true},
		{"function that releases its argument", `function leaf($x) { unset($x); }`, false},
		{"function that echoes its argument", `function leaf($x) { echo $x; }`, false},
		{"function that iterates its argument", `function leaf($x) { foreach ($x as $v) { } }`, false},
		{"function that writes a property", `function leaf($x) { $x->p = 1; }`, false},
		{"function that writes an offset", `function leaf($x) { $x[0] = 1; }`, false},
		{"function that writes an offset of an array", `function leaf(array $x) { $x[0] = 1; return $x[0] + count($x); }`, true},
		{"method that returns its argument as text", `class L { static function m($x): string { return $x; } } function leaf($x) { return L::m($x); }`, false},
		{"closure that returns its argument as text", `function leaf($x) { return function ($y): string { return $y; }; }`, true},
		{"function that reads a property", `function leaf($x) { return $x->p; }`, false},
		{"function that calls into the project", `function leaf($x) { return other($x); }`, false},
		{"function that makes an object", `function leaf($x) { return new Thing($x); }`, false},
		{"function that runs a callback", `function leaf($x) { return array_map($x, []); }`, false},
		{"function declared conditionally", `if (!function_exists('leaf')) { function leaf($x) { return 1; } }`, false},
		{"generator", `function leaf($x) { yield other($x); }`, true},
		{"generator declared conditionally", `if (!function_exists('leaf')) { function leaf($x) { yield 1; } }`, false},
		{"function taking text", `function leaf(string $x) { return 1; }`, false},
		{"function taking text or nothing", `function leaf(?string $x) { return 1; }`, false},
		{"function taking a callable", `function leaf(int|callable $x) { return 1; }`, false},
		{"function taking a number", `function leaf(int $x) { return 1; }`, true},
		{"function taking text, given text", `function leaf($x) { return inner('a'); } function inner(string $s) { return 1; }`, true},
		{"function with a default that loads a class", `function leaf($x, $y = Thing::K) { return 1; }`, false},
		{"function with a default that makes an object", `function leaf($x, $y = new Thing) { return 1; }`, false},
		{"function with a constant default", `function leaf($x, $y = PHP_INT_SIZE) { return 1; }`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := summarized(t, tc.src+` function probe($x) {} function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } leaf($a); probe($u); }`)
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

// In a namespace, a constant defined under a computed name may be the one an
// unqualified name reads instead of PHP's own, and may be an object.
func TestComputedDefinesShadowConstants(t *testing.T) {
	const body = `function probe($x) {} function t($a) { static $u; $u = parse_url($a); if ($u === false) { return; } $x = PHP_EOL; echo $x; probe($u); }`
	for _, tc := range []struct {
		src  string
		kept bool
	}{
		{`namespace App; ` + body, true},
		{`namespace App; ` + body + ` define('App\OTHER', 1);`, true},
		{`namespace App; ` + body + ` define($name, 1);`, false},
		{`namespace App; ` + body + ` define($name, 1); define('App\OTHER', 2);`, false},
		{`namespace App; ` + body + ` define(...$args);`, false},
		{`namespace App; ` + body + ` define();`, false},
		{`namespace App; ` + body + ` define(constant_name: 'App\OTHER', value: 1);`, false},
		{`namespace App; ` + body + ` function define($n, $v) {} define($name, 1);`, true},
	} {
		t.Run(tc.src[len(body)+15:], func(t *testing.T) {
			e, f := summarized(t, tc.src)
			calls := callsNamed(f, "probe")
			if got := e.Excludes(argument(calls[len(calls)-1]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
}
