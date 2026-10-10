package infer_test

import "testing"

// var_export()/print_r() return a string only with a true $return.
func TestPrintReturn(t *testing.T) {
	checkAnywhere(t, `<?php
function f($x, bool $b) {
    t('ve', var_export($x, true));
    t('ve0', var_export($x));
    t('veFalse', var_export($x, false));
    t('veVar', var_export($x, $b));
    t('pr', print_r($x, true));
    t('pr0', print_r($x));
    t('named', print_r($x, return: true));
}
`, map[string]string{"ve": "string", "ve0": "null", "veFalse": "null", "veVar": "null|string", "pr": "string", "pr0": "true", "named": "string"})
}

// thecodingmachine/safe wrappers are typed like the builtin, without the
// false (and preg_ null) they throw instead of returning.
func TestSafeFunctions(t *testing.T) {
	checkAnywhere(t, `<?php
namespace Safe {
/** @return string|array|null */
function preg_replace($p, $r, $s, int $l = -1, &$c = null) {}
/** @return string */
function file_get_contents(string $f) {}
function json_encode($v, int $f = 0) {}
/** @return int */
function strlen2($s) {}
/** @return \DateTime */
function substr($s, $o) {}
function no_such_builtin($x) {}
function str_replace($a, $b, $c) {}
}
namespace Safe\Sub { function strlen($s) {} }
namespace App {
use function Safe\preg_replace;
use function Safe\json_encode;
function f(string $s, array $a, $u = null) {
    t('pr', preg_replace('/x/', 'y', $s));
    t('prA', preg_replace('/x/', 'y', $a));
    t('fgc', \Safe\file_get_contents('x'));
    t('je', json_encode($a));
    t('better', \Safe\substr($s, 1));
    t('notBuiltin', \Safe\no_such_builtin(1));
    t('nested', \Safe\Sub\strlen('x'));
    t('unknownSubject', \Safe\str_replace('a', 'b', $u));
}
}
`, map[string]string{"pr": "string", "prA": "array", "fgc": "string", "je": "string", "better": `\DateTime`, "notBuiltin": "null", "nested": "null", "unknownSubject": "null"})
}

func TestArgumentDependentBuiltinReturns(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $f, $flags) {
    t('pe', pathinfo($f, PATHINFO_EXTENSION));
    t('pa', pathinfo($f));
    t('pv', pathinfo($f, $flags));
    t('g0', gettimeofday());
    t('g1', gettimeofday(true));
    t('gf', gettimeofday(false));
    t('gv', gettimeofday($flags));
}
`, map[string]string{
		"pe": "string",
		"pa": "array",
		"pv": "array|string",
		"g0": "array",
		"g1": "float",
		"gf": "array",
		"gv": "float|int[]",
	})
}
