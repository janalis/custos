package infer_test

import "testing"

func TestPerKeyNarrowing(t *testing.T) {
	checkWith(t, nil, `<?php
class Foo {}
/**
 * @param array{a: ?Foo, b?: int|null, c: string|false} $s
 * @param array<string, Foo|null> $m
 * @param array<string, mixed>|null $opt
 */
function run(array $s, array $m, ?array $opt, $i) {
    if (isset($s['a'])) { t('issetA', $s['a']); t('issetShape', $s); }
    t('after', $s['a']);
    if (isset($m['x'])) { t('issetMap', $m['x']); t('otherKey', $m['y']); }
    if ($m['x'] !== null) { t('notNull', $m['x']); }
    if (null != $m['x']) { t('notNullLoose', $m['x']); }
    if (!empty($s['c'])) { t('notEmpty', $s['c']); }
    if ($s['c']) { t('truthy', $s['c']); }
    if (array_key_exists('b', $s)) { t('keyExists', $s); t('keyExistsVal', $s['b']); }
    if (isset($opt['k'])) { t('container', $opt); }
    t('ternary', isset($m['x']) ? $m['x'] : null);
    t('and', isset($m['x']) && $m['x'] instanceof Foo ? 1 : $m['x']);
    /** @var array<string, mixed> $d */
    $d = f();
    if (isset($d['n']) && is_numeric($d['n'])) { t('numeric', $d['n']); }
    if (!isset($m['x'])) { return; }
    t('guard', $m['x']);
}
function writes(array $s) {
    /** @var array<string, Foo|null> $m */
    $m = f();
    if (isset($m['x'])) { $m['x'] = null; t('rewritten', $m['x']); }
    if (isset($m['x'])) { $m[$GLOBALS['k']] = null; t('computed', $m['x']); }
    if (isset($m['x'])) { $m['y'] = null; t('otherWrite', $m['x']); }
    if (isset($m['x'])) { unset($m['x']); t('unset', $m['x']); }
    if (isset($m['x'])) { h($m); t('passed', $m['x']); }
    if (!isset($m['x'])) { $m['x'] = new Foo(); }
    t('defaulted', $m['x']);
    if (isset($m['x'])) { sort($m); t('byRef', $m['x']); }
}
function reassign() {
    /** @var array<string, Foo|null> $m */
    $m = f();
    if (isset($m['x'])) { $m = ['x' => null]; t('reassigned', $m['x']); }
}
class Holder {
    /** @var array<string, Foo|null> */
    private array $p = [];
    function m() {
        if (isset($this->p['x'])) { t('prop', $this->p['x']); }
        if (isset($this->p['x'])) { $this->reset(); t('propCall', $this->p['x']); }
    }
    function reset() {}
}
`, map[string]string{
		"issetA":       `\Foo`,
		"issetShape":   `array{a: \Foo, b?: int|null, c: false|string}`,
		"after":        `\Foo|null`,
		"issetMap":     `\Foo`,
		"otherKey":     `\Foo|null`,
		"notNull":      `\Foo`,
		"notNullLoose": `\Foo`,
		"notEmpty":     "string",
		"truthy":       "string",
		"keyExists":    `array{a: \Foo|null, b: int|null, c: false|string}`,
		"keyExistsVal": "int|null",
		"container":    "non-empty mixed[]",
		"ternary":      `\Foo|null`,
		"and":          `\Foo|int|null`,
		"guard":        `\Foo`,
		"numeric":      "float|int|string", // not int: is_numeric() on mixed
		"rewritten":    `\Foo|null`,
		"computed":     `\Foo|null`,
		"otherWrite":   `\Foo`,
		"unset":        `\Foo|null`,
		"reassigned":   "null",
		"passed":       `\Foo`,
		"byRef":        `\Foo|null`,
		"defaulted":    `\Foo`,
		"prop":         `\Foo`,
		"propCall":     `\Foo|null`,
	})
}
