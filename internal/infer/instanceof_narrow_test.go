package infer_test

import "testing"

func TestNegativeInstanceofDropsSubtypes(t *testing.T) {
	checkWith(t, nil, `<?php
interface PathInterface {}
class Path implements PathInterface {}
class Other {}
function run(string|Path|Other $p) {
    if ($p instanceof PathInterface) { return; }
    t('rest', $p);
}
`, map[string]string{"rest": `\Other|string`})
}

func TestArrayRandKeyCount(t *testing.T) {
	checkWith(t, nil, `<?php
function run(array $a, int $n) {
    t('one', array_rand($a));
    t('literal', array_rand($a, 1));
    t('many', array_rand($a, $n));
}
`, map[string]string{"one": "int|string", "literal": "int|string", "many": "array|int|string"})
}

func TestReassignmentInsideGuard(t *testing.T) {
	checkWith(t, nil, `<?php
class Finder { public function find(): string|false { return ''; } }
function f(?array $php, Finder $finder) {
    if (null === $php) {
        $php = $finder->find();
        t('inner', $php);
        $php = false === $php ? null : [$php];
        t('replaced', $php);
    }
    t('after', $php);
}
`, map[string]string{"inner": "false|string", "replaced": "null|string[]{0: string}", "after": "array|null|string[]"})
}

func TestTypeCheckOnMixedMember(t *testing.T) {
	checkWith(t, nil, `<?php
function g(): mixed { return 1; }
function f(string|null $s, $b) {
    if ($b) { $s = g(); }
    if (!is_scalar($s)) { return; }
    t('scalar', $s);
}
`, map[string]string{"scalar": "bool|false|float|int|string|true"})
}
