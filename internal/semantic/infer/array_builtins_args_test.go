package infer_test

import (
	"fmt"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

const arrayBuiltinsSrc = `<?php
/**
 * @param list<int> $l
 * @param array<string, string> $m
 * @param list<array{id: int, name: string}> $rows
 * @param resource $h
 * @param \Iterator<string, \DateTime> $it
 */
function f(array $l, array $m, array $rows, array $plain, $h, string $s, \Iterator $it, \Traversable $tr, $u, int $i, $flags) {
    t('merge', array_merge($l, $m)); t('mergeShapes', array_merge(['a' => 1, 5 => 'x'], ['a' => 'y', 7 => true]));
    t('replaceShapes', array_replace(['a' => 1, 5 => 'x'], [5 => true, 9 => 1.5]));
    t('mergeNone', array_merge()); t('replaceNone', array_replace()); t('mergePlain', array_merge($plain, $l));
    t('mergeSpread', array_merge(...$rows)); t('mergeNotArray', array_merge($u)); t('mergeEmpty', array_merge([], []));
    t('chunk', array_chunk($l, 2)); t('chunkKeys', array_chunk($m, 2, true)); t('chunkPlain', array_chunk($plain, 2)); t('chunkUnknown', array_chunk($u, 2));
    t('combine', array_combine($l, $m)); t('combinePlain', array_combine($plain, $plain)); t('combineUnknown', array_combine($u, $l));
    t('column', array_column($rows, 'name')); t('columnNull', array_column($rows, null)); t('columnIndexed', array_column($rows, 'id', 'name'));
    t('columnMissing', array_column($rows, 'other')); t('columnPlain', array_column($plain, 'id')); t('columnComputed', array_column($rows, $s));
    t('fill', array_fill(0, 3, 'x')); t('fillUnknown', array_fill(0, 3, $u));
    t('fillKeys', array_fill_keys($l, true)); t('fillKeysUnknown', array_fill_keys($l, $u));
    t('pad', array_pad($l, 5, 'x')); t('padPlain', array_pad($plain, 5, 'x')); t('padUnknown', array_pad($l, 5, $u));
    t('keys', array_keys($m)); t('keysList', array_keys(['a' => 1])); t('keysSearch', array_keys($l, 3)); t('keysPlain', array_keys($plain)); t('keysUnknown', array_keys($u));
    t('flip', array_flip($l)); t('countValues', array_count_values($l)); t('countValuesUnknown', array_count_values($u));
    t('intersectKey', array_intersect_key(['a' => 1, 'b' => 's'], $m)); t('diff', array_diff($l, [1])); t('diffUnknown', array_diff($u, $l));
    t('i2a', iterator_to_array($it)); t('i2aList', iterator_to_array($it, false)); t('i2aPlain', iterator_to_array($tr)); t('i2aUnknown', iterator_to_array($u));
    t('range', range(1, $i)); t('rangeStep', range(1, 10, 2)); t('rangeFloat', range(0, 1, 0.25)); t('rangeStr', range('a', 'e'));
    t('csv', fgetcsv($h)); t('strCsv', str_getcsv($s)); t('file', file($s)); t('glob', glob($s));
    t('split', preg_split('/,/', $s)); t('splitFlags', preg_split('/,/', $s, -1, PREG_SPLIT_NO_EMPTY | PREG_SPLIT_DELIM_CAPTURE));
    t('splitOffset', preg_split('/,/', $s, -1, \PREG_SPLIT_OFFSET_CAPTURE)); t('splitZero', preg_split('/,/', $s, -1, 0));
    t('splitDynamic', preg_split('/,/', $s, -1, $flags)); t('splitInt', preg_split('/,/', $s, -1, 4)); t('splitAnd', preg_split('/,/', $s, -1, PREG_SPLIT_NO_EMPTY & $flags));
    t('values', array_values(['a' => 1, 'b' => 'x'])); t('valuesEmpty', array_values([])); t('unique', array_unique(['a' => 1]));
    t('url', parse_url($s));
}
`

func TestArrayBuiltinArguments(t *testing.T) {
	checkShape(t, arrayBuiltinsSrc, map[string]string{
		"chunk":              "int[][]<non-empty array>",
		"chunkKeys":          "string[][]<non-empty array>",
		"chunkPlain":         "array[]<non-empty array>",
		"chunkUnknown":       "array",
		"column":             "string[]",
		"columnComputed":     "array",
		"columnIndexed":      "int[]",
		"columnMissing":      "array",
		"columnNull":         "array[]<array{id: int, name: string}>",
		"columnPlain":        "array",
		"combine":            "string[]",
		"combinePlain":       "array",
		"combineUnknown":     "array",
		"countValues":        "int[]",
		"countValuesUnknown": "array",
		"csv":                "non-empty false|null[]|string[]",
		"diff":               "int[]",
		"diffUnknown":        "array",
		"file":               "false|string[]",
		"fill":               "string[]",
		"fillKeys":           "true[]",
		"fillKeysUnknown":    "array",
		"fillUnknown":        "array",
		"flip":               "int[]",
		"glob":               "false|string[]",
		"i2a":                "\\DateTime[]",
		"i2aList":            "\\DateTime[]",
		"i2aPlain":           "mixed[]",
		"i2aUnknown":         "mixed[]",
		"intersectKey":       "array{a?: int, b?: string}",
		"keys":               "string[]",
		"keysList":           "non-empty string[]",
		"keysPlain":          "int[]|string[]",
		"keysSearch":         "int[]",
		"keysUnknown":        "int[]|string[]",
		"merge":              "int[]|string[]",
		"mergeEmpty":         "array{}",
		"mergeNone":          "array{}",
		"mergeNotArray":      "array",
		"mergePlain":         "array",
		"mergeShapes":        "array{a: string, 0: string, 1: true}",
		"mergeSpread":        "array",
		"pad":                "int[]|string[]",
		"padPlain":           "array",
		"padUnknown":         "array",
		"range":              "non-empty int[]",
		"rangeFloat":         "array",
		"rangeStep":          "non-empty int[]",
		"rangeStr":           "array",
		"replaceNone":        "array",
		"replaceShapes":      "array{a: int, 5: true, 9: float}",
		"split":              "false|string[]",
		"splitAnd":           "array[]|false|string[]<array{0: string, 1: int}>",
		"splitDynamic":       "array[]|false|string[]<array{0: string, 1: int}>",
		"splitFlags":         "false|string[]",
		"splitInt":           "array[]|false|string[]<array{0: string, 1: int}>",
		"splitOffset":        "array[]|false<array{0: string, 1: int}>",
		"splitZero":          "false|string[]",
		"strCsv":             "non-empty null[]|string[]",
		"unique":             "non-empty int[]",
		"url":                "array|false{scheme?: string, host?: string, port?: int, user?: string, pass?: string, path?: string, query?: string, fragment?: string}",
		"values":             "non-empty int[]|string[]",
		"valuesEmpty":        "array",
	})
	checkVer(t, phpversion.PHP73, false, arrayBuiltinsSrc, map[string]string{
		"chunk":     "int[][]|null<non-empty array>",
		"combine":   "false|string[]",
		"csv":       "non-empty false|null|null[]|string[]",
		"mergeNone": "array",
	})
}

// Merging shapes past the shape size keeps the elements' union.
func TestArrayMergeOverShapeCap(t *testing.T) {
	var a, b strings.Builder
	for i := range 20 {
		fmt.Fprintf(&a, "'a%d' => %d, ", i, i)
		fmt.Fprintf(&b, "'b%d' => 's', ", i)
	}
	checkShape(t, `<?php
t('merged', array_merge([`+a.String()+`], [`+b.String()+`]));
`, map[string]string{"merged": "non-empty int[]|string[]"})
}
