package infer_test

import "testing"

// An empty array literal adds no element: iterating a typed array or `[]`
// gives the typed array's elements.
func TestEmptyArrayAlternatives(t *testing.T) {
	check(t, `<?php
class Foo {}
/** @return list<Foo> */
function found(): array { return []; }
/** @param list<Foo>|null $maybe */
function run(?array $maybe, bool $c) {
    foreach ($maybe ?? [] as $a) { t('coalesce', $a); }
    foreach ($c ? found() : [] as $b) { t('ternary', $b); }
    foreach (found() ?: [] as $d) { t('elvis', $d); }
    $list = $c ? found() : [];
    t('list', $list);
    foreach ($list as $e) { t('assigned', $e); }
    $none = $c ? [] : null;
    t('none', $none);
}
`, map[string]string{
		"coalesce": `\Foo`,
		"ternary":  `\Foo`,
		"elvis":    `\Foo`,
		"list":     `\Foo[]`,
		"assigned": `\Foo`,
		"none":     "array|null",
	})
}
