package infer_test

import "testing"

// A write below an element changes it along the keys written: a literal key
// of its shape holds the value written below it, an append or a computed
// key joins its elements, and the elements it does not change keep theirs.
func TestNestedWritesKeepShapes(t *testing.T) {
	checkAnywhere(t, `<?php
class Edition {} class File { public function edition(): Edition { return new Edition; } public function id(): int { return 1; } }
class Verse {} class Form {}
/** @param list<File> $files */
function groups(array $files) {
    $groups = [];
    foreach ($files as $file) {
        $id = $file->id();
        if (!isset($groups[$id])) {
            $groups[$id] = ['edition' => $file->edition(), 'files' => []];
        }
        $groups[$id]['files'][] = $file;
    }
    foreach ($groups as $group) { t('group', $group); }
}
/** @param list<Form> $forms */
function annotated(array $forms, Verse $v, string $k) {
    /** @var array<string, array{verse: Verse, forms: array<int, Form>}> $c */
    $c = [];
    foreach ($forms as $i => $form) {
        $c[$k] ??= ['verse' => $v, 'forms' => []];
        $c[$k]['forms'][$i] = $form;
    }
    foreach ($c as $entry) { t('entry', $entry); }
}
function others(array $plain, $object, string $k) {
    $p = [$k => $plain];
    $p[$k]['x']['y'] = 1;
    t('plain', $p[$k]);
    $o = [$k => $object];
    $o[$k]['x'] = 1;
    t('object', $o[$k]);
    /** @var array<string, array{a: int, ...}> $open */
    $open = [];
    $open[$k]['b']['c'] = 1;
    t('open', $open[$k]);
    $s = ['', ''];
    $s[0][1]['x'] = 1;
    t('stringBelow', $s[0]);
}
function readBack(array $rows) {
    $refs = [];
    $outcome = ['count' => 0, 'edited' => []];
    foreach ($rows as [$key, $id, $row]) {
        $refs[$key] ??= ['id' => $id, 'pairs' => 0];
        $refs[$key]['id'] = min($refs[$key]['id'], $id);
        ++$refs[$key]['pairs'];
        $outcome['edited'][] = $row;
        ++$outcome['count'];
    }
    foreach ($refs as $ref) { t('readBack', $ref); }
    foreach (['count', 'edited'] as $k) { t('computedRead', $outcome[$k]); }
}
/**
 * @param array<string, callable> $cbs
 * @param array<string, bool> $flags
 */
function unusual(array $cbs, array $flags, string $k, $y) {
    $flags[$k]['x'] = 1;
    t('bool', $flags[$k]);
    $cbs[$k]['x'] = 1;
    t('callable', $cbs[$k]);
    $d = [];
    [$d[$k]['x'], $other] = [1, 2];
    t('destructured', $d[$k]);
    $r = [];
    $r[$k]['x'] = &$y;
    t('reference', $r[$k]);
}
/** @param array<int, array<int, list<int>>> $m */
function rewrite(array $m, array $keys) {
    foreach ($keys as [$a, $b]) {
        t('loopRead', $m[$a]);
        $copy = $m[$a][$b];
        sort($copy);
        $m[$a][$b] = $copy;
    }
}
`, map[string]string{
		// Writes into an array that started empty may also create the
		// element: its other keys are optional. A typed array's elements
		// are taken as present.
		"group":  `array{edition?: \Edition, files: \File[]}`,
		"entry":  `array{verse: \Verse, forms: \Form[]}`,
		"plain":  "array",
		"object": "?unknown",
		// A key whose old value is unknown holds what is written: unknown
		// below it; a scalar keeps its type (PHP throws on a write below).
		"open":         "array{a: int, ...}",
		"stringBelow":  "string",
		"computedRead": "array|float|int",
		// What the loop wrote back is what it read: nothing new.
		"loopRead": "int[][]<array>",
		// Values computed from the element itself are unknown: its keys stay.
		"readBack": "non-empty array|mixed[]{id?: ?unknown, pairs?: ?unknown}",
		// A callable may be an array that stops being one; destructuring
		// and references leave arrays of unknown values.
		"callable":     "?unknown",
		"bool":         "bool|int[]{x: int}", // false becomes an array, true keeps its type
		"destructured": "array",
		"reference":    "array",
	})
}
