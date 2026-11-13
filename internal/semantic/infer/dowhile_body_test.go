package infer_test

import "testing"

// A value stored in the body of a do-while loop comes back to the top of
// the body only once the condition held; not through an inner loop.
func TestDoWhileBodyBackEdge(t *testing.T) {
	checkAnywhere(t, `<?php
function next_page(): ?string { return null; }
function f(array $xs) {
    $r = 'a';
    do {
        t('do', $r);
        $r = next_page();
    } while ($r);
    $s = 'a';
    do {
        foreach ($xs as $x) {
            t('inner', $s);
            $s = next_page();
        }
    } while ($s);
}
`, map[string]string{"do": "string", "inner": "null|string"})
}
