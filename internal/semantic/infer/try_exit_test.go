package infer_test

import "testing"

// A return or throw inside a try block leaves the rest of the block for
// good: a definition before it does not reach the later statements of the
// block, but still reaches the catch block and the code after the try.
func TestExitInsideTryBlock(t *testing.T) {
	checkAnywhere(t, `<?php
class P {}
function t($label, $value) {}
function f($c) {
    $q = 1;
    try {
        $p = new P();
        if ($c) {
            unset($p);
            $q = 'a';
            return 1;
        }
        t('try', $p);
        if ($c) {
            $q = 2.5;
            throw new \Exception();
        }
    } catch (\Exception $e) {
        t('catch', $q);
    }
    t('after', $q);
}
`, map[string]string{"try": `\P`, "catch": "float|int<1, 1>|string<len 1, 1>", "after": "float|int<1, 1>|string<len 1, 1>"})
}
