package infer_test

import "testing"

// Refined PHPDoc types keep the values they allow: int ranges and strings
// of at least one byte, alone or on a native int or string (with null when
// the native type admits it).
func TestRefinedDocTypes(t *testing.T) {
	checkShape(t, `<?php
/**
 * @param positive-int $p
 * @param non-negative-int $nn
 * @param non-zero-int $nz
 * @param int<1, 10> $r
 * @param int<min, -1> $neg
 * @param non-empty-string $s
 * @param numeric-string $num
 * @param positive-int $q
 * @param int $plain
 * @param int<10, 1> $reversed
 * @param int<1, many> $word
 */
function f(int $p, $nn, $nz, int $r, $neg, string $s, $num, ?int $q, int $plain, $reversed, $word) {
 t('p', $p); t('nn', $nn); t('nz', $nz); t('r', $r); t('neg', $neg); t('s', $s); t('num', $num);
 t('q', $q); t('plain', $plain); t('reversed', $reversed); t('word', $word);
}
`, map[string]string{
		"p": "int<1, max>", "nn": "int<0, max>", "nz": "int<min, -1>|int<1, max>", "r": "int<1, 10>",
		"neg": "int<min, -1>", "s": "string<len 1, max>", "num": "string<len 1, max>",
		"q": "int<1, max>|null", "plain": "int", "reversed": "int", "word": "int",
	})
}
