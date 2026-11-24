package infer_test

import "testing"

// A foreach key is int|string only for an array: an iterator may yield keys
// of any type, so a subject not known to be an array leaves its key unknown,
// and an array beside a Traversable adds its keys to the iterator's.
func TestForeachKeyKinds(t *testing.T) {
	check(t, `<?php
function f($untyped, \Traversable $it, iterable $items, ?\WeakMap $map, \WeakMap $weak, array $list) {
 foreach ($untyped as $k => $v) { t('untyped', $k); }
 foreach ($it as $k => $v) { t('traversable', $k); }
 foreach ($items as $k => $v) { t('iterable', $k); }
 foreach ($map ?? [] as $k => $v) { t('weakOrArray', $k); }
 foreach ($weak as $k => $v) { t('weak', $k); }
 foreach ($list as $k => $v) { t('array', $k); }
}
`, map[string]string{
		"untyped": "?unknown", "traversable": "?unknown", "iterable": "?unknown",
		"weakOrArray": "int|object|string", "weak": "object", "array": "int|string",
	})
}
