package infer_test

import "testing"

// Conditions guarding a try statement hold in its body, catches and finally;
// guards inside the try body do not reach its catches.
func TestNarrowingAcrossTry(t *testing.T) {
	checkWith(t, nil, `<?php
function g(): void {}
function outer(?int $a) {
    if ($a === null) { return; }
    try { t('body', $a); } catch (\Exception $e) { t('catch', $a); } finally { t('finally', $a); }
}
function inIf(?int $a) {
    if ($a !== null) {
        try { g(); } catch (\Exception $e) { t('ifCatch', $a); }
    }
}
function inner(?int $a) {
    try {
        if ($a === null) { return; }
        g();
        t('innerBody', $a);
    } catch (\Exception $e) {
        t('innerCatch', $a);
    }
}
`, map[string]string{
		"body": "int", "catch": "int", "finally": "int", "ifCatch": "int",
		"innerBody": "int", "innerCatch": "int|null",
	})
}
