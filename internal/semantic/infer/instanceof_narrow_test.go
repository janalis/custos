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

// An elseif chain on instanceof narrows the join past it.
func TestElseIfInstanceofJoin(t *testing.T) {
	checkAnywhere(t, `<?php
class Lang {} class Code {} class Sub extends Code {}
interface Marker {}
/**
 * @param string|Code $code
 * @param string|Code $c2
 * @param string|Code $c3
 */
function f($code, $c2, $c3, $x) {
    if ($code instanceof Lang) { $z = 1; } elseif ($code instanceof Code) { $code = 'x'; }
    t('join', $code);
    if ($c2 instanceof Code) { $c2 = 'y'; } else { $w = 1; }
    t('withElse', $c2);
    if ($c3 instanceof Sub) { $q = 1; } elseif ($x) { $c3 = $x; }
    t('unknownWrite', $c3);
    if ($c3 instanceof Code) { foo(); } elseif ($x) { $c3 .= 'a'; }
    t('otherWrite', $c3);
    if ($c2 instanceof Marker) { t('iface', $c2); }
    if ($x instanceof Code) { return; } elseif ($x) { throw new \Exception(); }
    t('leaves', $x);
}
`, map[string]string{
		"join": "string", "withElse": "string", "unknownWrite": "?unknown", "otherWrite": "?unknown",
		"iface": `\Marker|string`, "leaves": "?unknown",
	})
}

// A string failing is_numeric() is still a string.
func TestNegatedIsNumericOnString(t *testing.T) {
	checkAnywhere(t, `<?php
function e(string $v, int|string $w) {
    if (is_numeric($v)) { t('numeric', $v); return $v; }
    t('rest', $v);
    if (!is_numeric($w)) { t('notNumeric', $w); }
}
`, map[string]string{"numeric": "string", "rest": "string", "notNumeric": "string"})
}

func TestNegatedIsObjectOnDocumentedClass(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
function f() {
    /** @var Foo $o */
    $o = g();
    if (!is_object($o)) { t('notObj', $o); }
}
`, map[string]string{
		"notObj": "?unknown",
	})
}
