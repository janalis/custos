package infer_test

import "testing"

// Elements created by nested writes are arrays of what those writes store.
func TestNestedWritesCreateElements(t *testing.T) {
	checkAnywhere(t, `<?php
function f(array $rows) {
    $groups = [];
    $deep = [];
    $counted = [];
    foreach ($rows as $k => $row) {
        $groups[$k][] = 'name';
        $deep[$k]['a']['b'] = 1;
        $counted[$k][$row]++;
    }
    t('group', $groups[$k]);
    t('deep', $deep[$k]);
    t('counted', $counted[$k]);
    foreach ($groups as $group) { t('each', $group); }
    $shape = ['x' => 1];
    $shape['k']['y'] = 'v';
    t('literal', $shape['k']);
}
`, map[string]string{"group": `string[]`, "deep": `array`, "counted": `array`, "each": `string[]`, "literal": `string[]`})
}
