package infer_test

import "testing"

// A type check of an element narrows it, written `$a['k'] ?? null` to
// avoid the undefined key too, and also when nothing is known of the array.
// is_null() of `$a['k'] ?? null` holds when the key is missing: it does not.
func TestElementTypeChecks(t *testing.T) {
	check(t, `<?php
/** @param array<string, mixed> $a */
function run(array $a, $v) {
    if (is_string($a['k'] ?? null)) { t('coalesce', $a['k']); }
    if (\is_int($a['x'] ?? null) || \is_float($a['x'] ?? null)) { t('either', $a['x']); }
    if (is_null($a['n'] ?? null)) { t('isNull', $a['n']); }
    if (is_string($a['m'] ?? '')) { t('otherDefault', $a['m']); }
    if (is_array($v) && is_string($v['date'] ?? null)) { t('untyped', $v['date']); }
    if (!is_string($a['k'] ?? null)) { return; }
    t('early', $a['k']);
}
`, map[string]string{
		"coalesce":     "string",
		"either":       "float|int",
		"isNull":       "mixed",
		"otherDefault": "mixed",
		"untyped":      "string",
		"early":        "string",
	})
}

func TestNullableArrayOffsetWriteDropsNull(t *testing.T) {
	check(t, `<?php function f(?array $a, ?array $b, bool $write) {
 $a['k'] = 1; t('written', $a);
 if ($write) { $b['k'] = 1; } t('conditional', $b);
 }`, map[string]string{"written": "array", "conditional": "array|null"})
}

func TestNullableNonArrayOffsetWriteDropsNull(t *testing.T) {
	check(t, `<?php function f(?int $a, ?ArrayAccess $b) {
 $a['k'] = 1; t('scalar', $a);
 $b['k'] = 1; t('object', $b);
 }`, map[string]string{"scalar": "int|int[]", "object": `\ArrayAccess|int[]`})
}
