package infer_test

import "testing"

func TestElementWriteIntoNull(t *testing.T) {
	checkAnywhere(t, `<?php
function a(?array $n) { $n['k'] = 1; t('param', $n); }
function b() { $n = null; $n[] = 'x'; t('fromNull', $n); }
function c(bool $c) { $n = null; if ($c) { $n['k'] = 1; } t('maybe', $n); }
function d(?array $n, bool $c) { if ($c) { $n['k'] = 1; t('inBranch', $n); } t('after', $n); }
function e(?array $n) { $n['k'] = 1; $n = null; t('reassigned', $n); }
function f(?array $n) { $n['k'] = 1; sort($n); t('byRef', $n); }
function g(?string $s) { $s[0] = 'a'; t('string', $s); }
function h(?array $n) { $n['a']['b'] = 1; t('nested', $n); }
function i($n) { if ($n === null) { $n = null; } $n['k'] = 1; t('unknownBase', $n); }
function j(?array $n) { foreach ([1] as $x) { $n[] = $x; } t('loopWrite', $n); }
`, map[string]string{
		"param": "array", "fromNull": "string[]<string<len 1, 1>>", "maybe": "int[]|null<int<1, 1>>", "inBranch": "array",
		"after": "array|null", "reassigned": "null", "byRef": "array", "string": "string|string[]<string<len 1, 1>>",
		"nested": "array", "loopWrite": "array|null",
	})
}

// TestElementWriteVivifies covers the other writes that make null (or
// false) an array: destructuring and increments into an offset, writes
// into a property's offset; a write into false is not followed.
func TestElementWriteVivifies(t *testing.T) {
	checkAnywhere(t, `<?php
class P {
    private ?array $p = null;
    private ?\ArrayObject $o = null;
    private ?string $s = null;
    private array|false $f = false;
    public static ?array $st = null;
    function m1() { $this->p['k'] = 1; t('prop', $this->p); }
    function m2() { self::$st[] = 1; t('static', self::$st); }
    function m3() { $this->p['a']['b'] = 1; t('propNested', $this->p); }
    function m4() { $this->o['k'] = 1; t('propObject', $this->o); }
    function m5() { $this->s[0] = 'x'; t('propString', $this->s); }
    function m6() { $this->f[] = 1; t('propFalse', $this->f); }
    function m7(bool $c) { if ($c) { $this->p['k'] = 1; } t('propBranch', $this->p); }
    function m8() { $this->p['k'] = 1; $this->reset(); t('propCall', $this->p); }
    function m9() { $this->p['k'] = 1; $this->p = null; t('propReset', $this->p); }
    function m10() { $this->p['k'] = 1; if ($this->p === null) { t('propChecked', $this->p); } }
    function reset() {}
}
function a(?array $a) { [$a['x'], $a['y']] = [1, 2]; t('destructured', $a); }
function b(?array $a) { $a['k']++; t('incremented', $a); }
function c(array|false $a) { $a['k'] = 1; t('false', $a); }
function d() { $x = false; $x[] = 1; t('falseLiteral', $x); }
function e(bool $c) { $x = $c ? false : null; $x[] = 1; t('falseOrNull', $x); }
function f(\ArrayObject|false|null $o) { $o['k'] = 1; t('objectFalse', $o); }
function g(?array $a, bool $c) { if ($c) { $a['k']++; } t('incrementBranch', $a); }
function h(?array $a) { foreach ([1] as $a['k']) {} t('foreach', $a); }
`, map[string]string{
		"prop": "array", "static": "array", "propNested": "array", "propObject": `\ArrayObject|array`,
		"propString": "array|string", "propFalse": "array|false", "propBranch": "array|null", "propCall": "array|null",
		"propReset": "array|null", "propChecked": "array",
		"destructured": "array", "incremented": "array", "false": "array|false", "falseLiteral": "false",
		"falseOrNull": "false|int[]<int<1, 1>>", "objectFalse": `\ArrayObject|false|int[]<int<1, 1>>`, "incrementBranch": "array|null",
		// The loop may not run: the write may not happen.
		"foreach": "array|null",
	})
}

// TestPropertyElementWriteForms covers the property forms of the other
// element writes: increments and decrements into an offset, destructuring
// into offsets (also nested and keyed), on instance and static properties.
func TestPropertyElementWriteForms(t *testing.T) {
	checkAnywhere(t, `<?php
class Q {
    private ?array $h = null;
    public static ?array $s = null;
    function m1(string $k) { $this->h[$k]++; t('inc', $this->h); }
    function m2(string $k) { $this->h[$k]--; t('dec', $this->h); }
    function m3() { ++$this->h['a']['b']; t('preIncNested', $this->h); }
    function m4() { self::$s['k']++; t('staticInc', self::$s); }
    function m5() { [$this->h['a'], $this->h['b']] = [1, 2]; t('destructured', $this->h); }
    function m6() { ['x' => $this->h['a']] = ['x' => 1]; t('keyed', $this->h); }
    function m7() { [[$this->h['a']]] = [[1]]; t('nestedList', $this->h); }
    function m8() { list($a, self::$s['k']) = [1, 2]; t('staticList', self::$s); }
    function m9() { [$a, $b] = [1, 2]; t('otherList', $this->h); }
    function m10(bool $c) { if ($c) { $this->h['k']++; } t('incBranch', $this->h); }
    function m11() { $this->h['k']++; $this->reset(); t('incCall', $this->h); }
    function m12() { $this->h++; t('incWhole', $this->h); }
    function reset() {}
}
`, map[string]string{
		"inc": "array", "dec": "array", "preIncNested": "array", "staticInc": "array",
		"destructured": "array", "keyed": "array", "nestedList": "array", "staticList": "array",
		"otherList": "array|null", "incBranch": "array|null", "incCall": "array|null", "incWhole": "array|null",
	})
}
