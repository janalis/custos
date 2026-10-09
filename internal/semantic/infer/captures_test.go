package infer_test

import (
	"fmt"
	"strings"
	"testing"
)

func TestCaptureSnapshots(t *testing.T) {
	check(t, `<?php
class Captured {}
function snapshots(bool $c, array $input, ?Captured $object) {
    $x = 1; $x = 's';
    $closure = function () use ($x) { t('closure', $x); return $x; };
    $arrow = fn() => $x;
    $x = [];
    t('callClosure', $closure()); t('callArrow', $arrow());
    if ($object instanceof Captured) {
        $narrow = function () use ($object) { t('narrow', $object); };
        $narrowArrow = fn() => $object;
        t('narrowArrow', $narrowArrow());
    }
    $before = 1;
    extract($input);
    $dynamic = function () use ($before) { t('dynamic', $before); };
    $dynamicArrow = fn() => $before;
    t('dynamicArrow', $dynamicArrow());
    $before = 'reset';
    $restored = fn() => $before;
    t('restored', $restored());
}

`, map[string]string{
		"closure": "string", "callClosure": "string", "callArrow": "string",
		"narrow": `\Captured`, "narrowArrow": `\Captured`,
		"dynamic": "?unknown", "dynamicArrow": "?unknown", "restored": "string",
	})
}

func TestCaptureBranchesArraysAndNesting(t *testing.T) {
	check(t, `<?php
function captures(bool $c, array $input) {
    $v = 1;
    if ($c) { $v = 's'; }
    $branch = function () use ($v) { t('branch', $v); };
    if ($c) { $maybe = 's'; }
    $partial = fn() => $maybe;
    t('partial', $partial());
    $a = ['x' => 1]; $a['y'] = 's';
    $array = function () use ($a) { t('array', $a); };
    $arrayArrow = fn() => $a;
    t('arrayArrow', $arrayArrow());
    $shape = ['x' => 1];
    $sealed = function () use ($shape) { t('sealed', $shape); };
    $value = 's';
    $nested = fn() => fn() => $value;
    t('nested', ($nested())());
    $shadow = fn(int $value) => $value;
    t('shadow', $shadow(1));
    $ref = function () use (&$value) { t('ref', $value); $value = 1; t('refSet', $value); };
    $missing = function () use ($undefined) { t('missing', $undefined); };
    $unknown = undeclared();
    $unknownArrow = fn() => $unknown;
    t('unknown', $unknownArrow());
}
`, map[string]string{
		"branch": "int|string", "partial": "null|string",
		"array": "int[]|string[]", "arrayArrow": "int[]|string[]", "sealed": "int[]",
		"nested": "string", "shadow": "int", "ref": "?unknown", "refSet": "int",
		"missing": "?unknown", "unknown": "?unknown",
	})
}

func TestArrowLocalMutations(t *testing.T) {
	checkAnywhere(t, `<?php
function arrows() {
    $x = null;
    $prefix = fn() => ++$x;
    $postfix = fn() => $x++;
    t('prefixArrow', $prefix()); t('postfixArrow', $postfix());
    t('outer', $x);
    $nested = fn() => fn() => ++$x;
    t('nestedPrefix', ($nested())());
    $before = 's';
    $assigned = fn() => [t('before', $before), $before = 1];
    $shape = ['k' => 1];
    $sealed = fn() => t('shape', $shape);
    $copy = function () use ($shape) { t('closureShape', $shape); };
}
`, map[string]string{
		"prefixArrow": "int", "postfixArrow": "null", "outer": "null", "nestedPrefix": "int",
		"before": "string", "shape": "int[]{k: int}", "closureShape": "int[]{k: int}",
	})
}

func TestCaptureLoopsAndMutation(t *testing.T) {
	checkAnywhere(t, `<?php
function captureLoop(bool $c) {
    $v = 1;
    while ($c) { $f = fn() => t('backedge', $v); $v = 's'; }
    $unknown = 1;
    while ($c) { $f = fn() => t('unknownBackedge', $unknown); $unknown = undeclared(); }
    $array = ['x' => 1];
    $alias =& $array;
    $f = fn() => t('aliased', $array);
    $unset = ['x' => 1]; unset($unset['x']);
    $f = fn() => t('unsetShape', $unset);
    for ($initialized = 's'; $c;) {}
    $f = fn() => t('initialized', $initialized);
}
$top = 1; $top = 's';
$global = fn() => t('global', $top);
if (rand(0, 1)) { $choice = 'first'; } else { $choice = 'second'; }
$globalChoice = fn() => t('globalBranch', $choice);
`, map[string]string{
		"backedge": "int|string", "unknownBackedge": "?unknown", "aliased": "int[]",
		"unsetShape": "int[]", "initialized": "string", "global": "string", "globalBranch": "string",
	})
}

func TestCaptureLimits(t *testing.T) {
	var source strings.Builder
	source.WriteString("<?php function limits() { $many = 1;\n")
	for i := 0; i < 513; i++ {
		fmt.Fprintf(&source, "$many = %d;\n", i)
	}
	source.WriteString("$f = fn() => t('definitions', $many); $deep = 1; $g = ")
	source.WriteString(strings.Repeat("fn() => ", 70))
	source.WriteString("t('depth', $deep); }")
	checkAnywhere(t, source.String(), map[string]string{"definitions": "?unknown", "depth": "?unknown"})
}

func TestArrowDefinitionDominance(t *testing.T) {
	checkAnywhere(t, `<?php
function arrowDefinitions(bool $c, string $s) {
    $x = 's';
    $f = fn() => [$x = 1, t('assignment', $x)];
    $g = fn() => [($c && ($x = 1)), t('conditional', $x)];
    $h = fn() => [($c ? ($x = 1) : 2), t('ternary', $x)];
    $m = fn() => [match ($c) { true => ($x = 1), default => 2 }, t('match', $x)];
    $out = fn() => [preg_match('/x/', $s, $matches), t('out', $matches)];
    $increments = fn() => [++$x, t('increment', $x)];
    $optional = fn() => [($c && ++$x), t('optionalIncrement', $x)];
    $arithmetic = fn() => [($x = 1) + 2, t('arithmetic', $x)];
    $condition = fn() => [($x = 1) ? 2 : 3, t('condition', $x)];
}
`, map[string]string{
		"assignment": "int", "conditional": "int|string", "ternary": "int|string", "match": "int|string",
		"out": "string[]", "increment": "float|int|string", "optionalIncrement": "float|int|string",
		"arithmetic": "int", "condition": "int",
	})
}
