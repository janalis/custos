package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

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

// A statement that leaves on one outcome of a test narrows what follows.
func TestStatementsLeavingOnAnOutcome(t *testing.T) {
	checkWith(t, nil, `<?php
function stop(): never { exit(1); }
function orExit(?int $a) { $a !== null or exit; t('orExit', $a); }
function orThrow(?int $a) { $a || throw new \Exception(); t('orThrow', $a); }
function andExit(?int $a) { $a === null and exit(1); t('andExit', $a); }
function shortTernary(?int $a) { $a ?: throw new \Exception(); t('short', $a); }
function elseExit(?int $a) { $a ? null : exit(1); t('elseExit', $a); }
function thenThrow(?int $a) { !$a ? throw new \Exception() : null; t('thenThrow', $a); }
function neverCall(?int $a) { $a !== null or stop(); t('never', $a); }
function bothLeave(?int $a) { $a ? exit(1) : exit(2); }
function noLeave(?int $a) { $a or print('x'); t('noLeave', $a); }
`, map[string]string{
		"orExit": "int", "orThrow": "int", "andExit": "int", "short": "int", "elseExit": "int",
		"thenThrow": "int", "never": "int", "noLeave": "int|null",
	})
}

// Before PHP 8.0 throw is a statement: an `or throw` written there (a parse
// error the file recovers from) leaves nothing.
func TestThrowExpressionBeforePHP8(t *testing.T) {
	src := `<?php
function f(?int $a) { $a || throw new \Exception(); u($a); }`
	for _, tc := range []struct {
		php  phpversion.Version
		want string
	}{{phpversion.PHP74, "int|null"}, {phpversion.PHP80, "int"}} {
		f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, tc.php, infer.Options{})
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "u" {
					if got := env.TypeOf(c.Args.Args[0].(*syntax.Arg).Value).String(); got != tc.want {
						t.Errorf("%v: got %s want %s", tc.php, got, tc.want)
					}
				}
			}
			return true
		})
	}
}

// Statements at file level guard what follows them like those of a function.
func TestFileLevelGuards(t *testing.T) {
	checkAnywhere(t, `<?php
$r = realpath($d);
if (!$r) { return; }
t('afterExit', $r);
$s = realpath($d);
$s !== false or exit(1);
t('afterOr', $s);
`, map[string]string{"afterExit": "string", "afterOr": "string"})
}
