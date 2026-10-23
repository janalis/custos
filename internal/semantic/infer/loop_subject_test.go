package infer_test

import "testing"

// The array a foreach iterates and the initializers of a for loop are
// evaluated once, before the loop: what its body assigns later does not
// reach them.
func TestLoopSubjectBeforeBody(t *testing.T) {
	check(t, `<?php
function run(\ReflectionClass $c, bool $b) {
    $methods = $c->getMethods();
    foreach ($methods as $method) {
        t('value', $method);
        if ($b) { $methods = ['GET']; }
        t('inside', $methods);
    }
    $n = 3;
    for ($m = $n; $b; $b = $m) {
        t('init', $m);
        $n = 'x';
    }
}
`, map[string]string{
		"value":  `\ReflectionMethod`,
		"inside": `\ReflectionMethod[]|string[]`,
		"init":   "int",
	})
}
