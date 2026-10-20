package infer_test

import "testing"

// A variable passed by reference to a user callable holds afterwards the
// type the parameter's `@param-out` documents, else its `@param`.
func TestByRefParamDocFallback(t *testing.T) {
	check(t, `<?php
/** @param int[] $xs */
function fill(array &$xs) {}
/**
 * @param int[] $xs
 * @param-out string[] $xs
 */
function convert(array &$xs) {}
function undocumented(array &$xs) {}
/** @param mixed $xs */
function anything(&$xs) {}
final class Box {
    /** @param array<string, int> $totals */
    public function add(array &$totals) {}
}
function f(Box $box) {
    $a = ['k' => 'v']; fill($a); t('param', $a);
    $b = []; convert($b); t('out', $b);
    $c = [1]; undocumented($c); t('none', $c);
    $d = 1; anything($d); t('mixed', $d);
    $e = []; $box->add($e); t('method', $e);
}
`, map[string]string{"param": "int[]", "out": "string[]", "none": "?unknown", "mixed": "?unknown", "method": "int[]"})
}
