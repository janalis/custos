package infer_test

import "testing"

// A guard leaves the block through every branch of an if/else, a try, a
// switch or a call that never returns: past it, the condition is false.
func TestGuardsThatLeaveThroughCompoundStatements(t *testing.T) {
	checkWith(t, nil, `<?php
class Fail {
    public static function now(): never { throw new \Exception(); }
    public function soon(): never { throw new \Exception(); }
}
function stop(): never { exit(1); }
function ifElse(?int $a, bool $b) {
    if ($a === null) { if ($b) { return; } else { throw new \Exception(); } }
    t('ifElse', $a);
}
function tryBody(?int $a) {
    if ($a === null) { try { return; } catch (\Exception $e) { return; } }
    t('try', $a);
}
function switchAll(?int $a, int $m) {
    if ($a === null) { switch ($m) { case 1: return; default: throw new \Exception(); } }
    t('switch', $a);
}
function neverFunc(?int $a) {
    if ($a === null) { stop(); }
    t('neverFunc', $a);
}
function neverStatic(?int $a) {
    if ($a === null) { Fail::now(); }
    t('neverStatic', $a);
}
function neverMethod(?int $a, Fail $f) {
    if ($a === null) { $f->soon(); }
    t('neverMethod', $a);
}
function nestedNever(?int $a, bool $b) {
    if ($a === null) { if ($b) { stop(); } else { Fail::now(); } }
    t('nestedNever', $a);
}
function elseIfNever(?int $a, int $b) {
    if ($a === null) { if ($b > 1) { stop(); } elseif ($b) { stop(); } else { Fail::now(); } }
    t('elseIfNever', $a);
}
`, map[string]string{
		"elseIfNever": "int",
		"ifElse":      "int", "try": "int", "switch": "int",
		"neverFunc": "int", "neverStatic": "int", "neverMethod": "int", "nestedNever": "int",
	})
}

// Branches that may fall through keep the condition open.
func TestGuardsThatMayFallThrough(t *testing.T) {
	checkWith(t, nil, `<?php
class Maybe { public function soon(): void {} }
function ifNoElse(?int $a, bool $b) {
    if ($a === null) { if ($b) { return; } }
    t('ifNoElse', $a);
}
function catchFalls(?int $a) {
    if ($a === null) { try { return; } catch (\Exception $e) { } }
    t('catch', $a);
}
function switchNoDefault(?int $a, int $m) {
    if ($a === null) { switch ($m) { case 1: return; } }
    t('switch', $a);
}
function nullsafe(?int $a, ?Maybe $m) {
    if ($a === null) { $m?->soon(); }
    t('nullsafe', $a);
}
function stop(): never { exit(1); }
function elseFalls(?int $a, bool $b) {
    if ($a === null) { if ($b) { stop(); } else { $b = false; } }
    t('elseFalls', $a);
}
function elseIfFalls(?int $a, int $b) {
    if ($a === null) { if ($b > 1) { stop(); } elseif ($b) { $b = 0; } else { stop(); } }
    t('elseIfFalls', $a);
}
function voidCall(?int $a, Maybe $m) {
    if ($a === null) { $m->soon(); }
    t('void', $a);
}
`, map[string]string{
		"ifNoElse": "int|null", "catch": "int|null", "switch": "int|null",
		"nullsafe": "int|null", "void": "int|null",
		"elseFalls": "int|null", "elseIfFalls": "int|null",
	})
}
