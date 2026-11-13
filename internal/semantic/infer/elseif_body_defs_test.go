package infer_test

import "testing"

// A definition made earlier in the elseif body holding a read reaches it:
// only the other branches of the chain exclude the read.
func TestElseIfBodyKeepsItsDefinitions(t *testing.T) {
	checkAnywhere(t, `<?php
function f(bool $x, bool $y, bool $z) {
    $a = null;
    if ($x) {
    } elseif ($y) {
        if ($z) {
            $a = 1;
        }
        t('nested', $a);
    }
    $b = null;
    if ($x) {
        $b = 'x';
    } elseif ($y) {
        $b = 2;
        t('direct', $b);
    } elseif ($z) {
        t('other', $b);
    }
}
`, map[string]string{"nested": "int<1, 1>|null", "direct": "int<2, 2>", "other": "null"})
}
