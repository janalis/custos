package infer_test

import "testing"

// is_array() narrows iterable and callable to an array of unknown elements,
// next to what the other members narrow to: a literal array assigned on one
// path does not make the elements known.
func TestIsArrayNarrowsIterableAndCallable(t *testing.T) {
	checkAnywhere(t, `<?php
function f(iterable $it, callable $cb, iterable|string $mix, iterable|int $plain, iterable $only, bool $c) {
    if ($c) {
        $it = ['x', 'y'];
        $cb = ['x', 'y'];
        $mix = ['x'];
    }
    if (\is_array($it)) { t('iterable', $it); }
    if (\is_array($cb)) { t('callable', $cb); }
    if (\is_array($mix)) { t('mixed', $mix); }
    if (\is_array($plain)) { t('plain', $plain); }
    if (\is_array($only)) { t('only', $only); }
}
`, map[string]string{
		"iterable": "array|string[]",
		"callable": "array|string[]",
		"mixed":    "array|string[]",
		"plain":    "array",
		"only":     "array",
	})
}
