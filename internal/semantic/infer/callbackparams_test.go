package infer_test

import "testing"

// A closure written as a builtin's callback gets its untyped parameters
// from the array the builtin passes it.
func TestCallbackParametersFromBuiltins(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App;
class Foo {}
/**
 * @param list<Foo> $foos
 * @param array<string, int> $map
 */
function run(array $foos, array $map, array $plain) {
    array_map(fn($f) => t('map', $f), $foos);
    array_map(function ($f) { t('mapClosure', $f); }, $foos);
    array_map(fn($a, $b) => t('zip', $b), $foos, $map);
    array_map(fn($a, $b, $c) => t('zipExtra', $c), $foos, $map);
    array_map(callback: fn($f) => t('named', $f), array: $foos);
    array_filter($map, fn($v) => t('filter', $v));
    array_filter($map, fn($k) => t('filterKey', $k), ARRAY_FILTER_USE_KEY);
    array_filter($map, fn($v, $k) => t('filterBoth', $k), \ARRAY_FILTER_USE_BOTH);
    array_filter($map, fn($v, $k) => t('filterBothValue', $v), ARRAY_FILTER_USE_BOTH);
    array_filter($map, fn($v, $k) => t('filterDynamic', $v), $GLOBALS['mode']);
    array_filter($map, fn($v) => t('filterOtherMode', $v), 7);
    array_walk($map, function (&$v, $k, $extra) { t('walk', $v); t('walkKey', $k); t('walkExtra', $extra); }, 1);
    array_reduce($foos, fn($carry, $f) => t('reduce', $f) ?? t('carry', $carry));
    usort($foos, fn($a, $b) => t('usort', $b));
    uasort($foos, fn($a, $b, $c) => t('uasortExtra', $c));
    uksort($map, fn($a, $b) => t('uksort', $a));
    array_map(fn($p) => t('plain', $p), $plain);
    array_map(fn(int $d) => t('declared', $d), $foos);
    array_map(fn(...$all) => t('variadic', $all), $foos);
    array_map(/** @param string $s */ fn($s) => t('doc', $s), $foos);
    other(fn($o) => t('user', $o), $foos);
    $cb = fn($v) => t('stored', $v);
    $x = [fn($v) => t('notArg', $v)];
    array_filter(...[$map, fn($v) => t('spread', $v)]);
    array_map(fn($f) => t('dynamicName', $f), ...$GLOBALS['lists']);
    ('array_map')(fn($f) => t('notNamed', $f), $foos);
    \array_map(fn($f) => t('qualified', $f), $foos);
    array_map($foos, fn($f) => t('wrongSlot', $f));
    array_map(function ($a, $b) { t('singleSecond', $b); }, $foos);
    array_map(function ($a, $b) { t('namedZipArg', $b); }, $foos, arrays: $map);
    $GLOBALS['o']->m(fn($m) => t('method', $m), $foos);
}
function other(callable $c, array $xs) {}
`, map[string]string{
		"map": `\App\Foo`, "mapClosure": `\App\Foo`, "zip": "int|null", "zipExtra": "?unknown",
		"named": `\App\Foo`, "filter": "int", "filterKey": "int|string", "filterBoth": "int|string",
		"filterBothValue": "int", "filterDynamic": "?unknown", "filterOtherMode": "?unknown",
		"walk": "int", "walkKey": "int|string", "walkExtra": "?unknown",
		"reduce": `\App\Foo`, "carry": "?unknown", "usort": `\App\Foo`, "uasortExtra": "?unknown",
		"uksort": "int|string", "plain": "?unknown", "declared": "int", "variadic": "?unknown",
		"doc": "string", "user": "?unknown", "stored": "?unknown", "notArg": "?unknown",
		"spread": "?unknown", "dynamicName": "?unknown", "notNamed": "?unknown", "qualified": `\App\Foo`,
		"wrongSlot": "?unknown", "singleSecond": "?unknown", "namedZipArg": "?unknown", "method": "?unknown",
	})
}

// A namespaced function's body, inferred for callers in other files, types
// its callbacks through the global fallback of array_map().
func TestCallbackParametersAcrossFiles(t *testing.T) {
	checkWith(t, map[string]string{"lib.php": `<?php
namespace Lib;
class Foo {}
/** @param list<Foo> $a */
function first(array $a) { return array_map(fn($x) => $x, $a); }
`}, `<?php
t('lib', \Lib\first([]));
`, map[string]string{"lib": `\Lib\Foo[]`})
}
