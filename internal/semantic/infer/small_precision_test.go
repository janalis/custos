package infer_test

import "testing"

// A callback parameter declared `?array` takes the elements passed as one
// declared `array` does; a closure declaring only `: array` returns what
// its body builds; `(array)` keeps what it can of its operand;
// array_replace_recursive() keeps the elements of its arrays.
func TestSmallArrayPrecision(t *testing.T) {
	checkAnywhere(t, `<?php
/**
 * @param list<array{lat: float}>|null $places
 * @param list<int> $ids
 * @param int|string|array<array-key, int|string> $id
 * @param array<string, array{a: int}> $x
 * @param array<string, array{b: string}> $y
 */
function run(?array $places, array $ids, $id, ?string $s, $object, array $x, array $y, array $plain, array $lists) {
    array_map(static function (?array $place) { t('nullable', $place); return 1; }, $places ?? []);
    array_map(static function (array $place) { t('notNull', $place); return 1; }, $places ?? []);
    t('arrow', array_map(fn ($i): array => ['y' => $i], $ids));
    t('arrowNullable', array_map(fn ($i): ?array => $i ? ['y' => $i] : null, $ids));
    t('arrowDoc', array_map(/** @return list<int> */ fn ($i): array => [$i], $ids));
    t('cast', (array) $id);
    t('castNullable', (array) $s);
    t('castObject', (array) $object);
    t('replaceRecursive', array_replace_recursive($x, $y));
    t('replaceUnknown', array_replace_recursive($x, $object));
    t('replacePlain', array_replace_recursive($x, $plain));
    t('replaceSpread', array_replace_recursive($x, ...$lists));
    t('replaceNamed', array_replace_recursive(array: $x));
}
`, map[string]string{
		"nullable":         "array{lat: float}",
		"notNull":          "array{lat: float}",
		"arrow":            "int[][]<array{y: int}>",
		"arrowNullable":    "int[][]|null[]<array{y: int}>",
		"arrowDoc":         "int[][]<array>",
		"cast":             "int[]|string[]",
		"castNullable":     "string[]",
		"castObject":       "array",
		"replaceRecursive": "array[]<non-empty array{a?: int, b?: string}>",
		"replaceUnknown":   "array",
		"replacePlain":     "array",
		"replaceSpread":    "array",
		"replaceNamed":     "array",
	})
}
