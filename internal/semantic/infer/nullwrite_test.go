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
function i(?array $a, array $xs) { foreach ($xs as $a['k']) {} t('foreachMayNotRun', $a); }
`, map[string]string{
		"prop": "array", "static": "array", "propNested": "array", "propObject": `\ArrayObject|array`,
		"propString": "array|string", "propFalse": "array|false", "propBranch": "array|null", "propCall": "array|null",
		"propReset": "array|null", "propChecked": "array",
		"destructured": "array", "incremented": "array", "false": "array|false", "falseLiteral": "false",
		"falseOrNull": "false|int[]<int<1, 1>>", "objectFalse": `\ArrayObject|false|int[]<int<1, 1>>`, "incrementBranch": "array|null",
		// The loop runs: the write happens.
		"foreach": "array",
		// The loop may not run: the write may not happen.
		"foreachMayNotRun": "array|null",
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
    function m13() { $this->h['a'] = 1; $this->reset(); $this->h['b']++; t('lastWrite', $this->h); }
    function m14() { $this->h['a'] = 1; $this->h['b']++; $this->reset(); t('lastBroken', $this->h); }
    function reset() {}
}
`, map[string]string{
		"inc": "array", "dec": "array", "preIncNested": "array", "staticInc": "array",
		"destructured": "array", "keyed": "array", "nestedList": "array", "staticList": "array",
		"otherList": "array|null", "incBranch": "array|null", "incCall": "array|null", "incWhole": "array|null",
		"lastWrite": "array", "lastBroken": "array|null",
	})
}

// TestProvenFalseOffsetWrite covers a write into an offset of false, which
// PHP turns into an array (deprecated from PHP 8.1): only the proven Env
// follows it, the native one keeps false.
func TestProvenFalseOffsetWrite(t *testing.T) {
	const head = `<?php function found(): array|false { return false; } function f(array|false $a, bool $c) { `
	cases := []struct {
		name, body, native, proven string
	}{
		{"literal", `$x = false; $x[] = 1; t('v', $x); }`, `false`, `int[]`},
		{"declared", `$a['k'] = 1; t('v', $a); }`, `array|false`, `array`},
		{"call result", `$r = found(); $r['k'] = 1; t('v', $r); }`, `array|false`, `array`},
		{"in a branch", `if ($c) { $a['k'] = 1; } t('v', $a); }`, `array|false`, `array|false`},
		{"true stays", `$x = $c; $x[] = 1; t('v', $x); }`, `bool`, `bool`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			native, proven := provenProbe(t, head+c.body)
			if native != c.native || proven != c.proven {
				t.Errorf("native %s, proven %s; want %s, %s", native, proven, c.native, c.proven)
			}
		})
	}
}

// TestProvenFalsePropertyOffsetWrite covers a write into an offset of a
// property that may hold false: as for a local, only the proven Env makes
// false an array, the native one keeps false.
func TestProvenFalsePropertyOffsetWrite(t *testing.T) {
	cases := []struct {
		name, body, native, proven string
	}{
		{"instance", `$this->f[] = 1; t('v', $this->f);`, `array|false`, `array`},
		{"static", `self::$s['k'] = 1; t('v', self::$s);`, `array|false`, `array`},
		{"narrowed", `if (self::$s === false) { self::$s[] = 1; t('v', self::$s); }`, `false`, `array`},
		{"false or null", `$this->n[] = 1; t('v', $this->n);`, `array|false`, `array`},
		{"in a branch", `if ($c) { $this->f[] = 1; } t('v', $this->f);`, `array|false`, `array|false`},
		{"call after", `$this->f[] = 1; $this->m(false); t('v', $this->f);`, `array|false`, `array|false`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := `<?php class P { private array|false $f = false; private array|false|null $n = null; private static array|false $s = false; function m(bool $c) { ` + c.body + ` } }`
			native, proven := provenProbe(t, src)
			if native != c.native || proven != c.proven {
				t.Errorf("native %s, proven %s; want %s, %s", native, proven, c.native, c.proven)
			}
		})
	}
}

// TestProvenUntypedPropertyOffsetWrite covers a write into an offset of a
// property no native type restricts (PHPDoc only, none, or mixed): the
// proven Env makes it an array only when the conditions guarding the write
// prove it held null, false or an array, and nothing may change it from the
// written property to the read. The native Env does not follow it.
func TestProvenUntypedPropertyOffsetWrite(t *testing.T) {
	cases := []struct {
		name, body, native, proven string
	}{
		{"untested", `$this->f[] = 1; t('v', $this->f);`, `?unknown`, `?unknown`},
		{"tested false", `if ($this->f === false) { $this->f[] = 1; t('v', $this->f); }`, `?unknown`, `array`},
		{"static", `if (self::$s === false) { self::$s['k'] = 1; t('v', self::$s); }`, `?unknown`, `array`},
		{"after the if", `if (self::$s === false) { self::$s['k'] = 1; } t('v', self::$s);`, `?unknown`, `?unknown`},
		{"no PHPDoc", `if ($this->u === null) { $this->u[] = 1; t('v', $this->u); }`, `?unknown`, `array`},
		{"an array", `if (is_array($this->u)) { $this->u[] = 1; t('v', $this->u); }`, `?unknown`, `array`},
		{"mixed", `if ($this->mx === false) { $this->mx[] = 1; t('v', $this->mx); }`, `mixed`, `array`},
		{"null or false", `if ($this->n === null || $this->n === false) { $this->n[] = 1; t('v', $this->n); }`, `?unknown`, `array`},
		{"destructured", `if ($this->f === false) { [$this->f['a'], $b] = [1, 2]; t('v', $this->f); }`, `?unknown`, `array`},
		{"incremented", `if ($this->f === false) { $this->f['a']++; t('v', $this->f); }`, `?unknown`, `array`},
		{"falsy", `if (!$this->f) { $this->f[] = 1; t('v', $this->f); }`, `?unknown`, `?unknown`},
		{"call in the value", `if ($this->f === false) { $this->f[] = $this->m(true); t('v', $this->f); }`, `?unknown`, `?unknown`},
		{"call in the key", `if ($this->f === false) { $this->f[$this->m(true)] = 1; t('v', $this->f); }`, `?unknown`, `?unknown`},
		{"call after", `if ($this->f === false) { $this->f[] = 1; $this->m(true); t('v', $this->f); }`, `?unknown`, `?unknown`},
		{"another property", `if ($this->f === false) { $this->u[] = 1; t('v', $this->f); }`, `?unknown`, `?unknown`},
		{"destructured second", `if ($this->f === false) { [$this->u['a'], $this->f['b']] = [1, 2]; t('v', $this->f); }`, `?unknown`, `array`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := `<?php class P { /** @var array|false */ private $f = false; /** @var array|false */ private static $s = false; private $u; private mixed $mx = false; /** @var array|null */ private $n; function m(bool $c) { ` + c.body + ` } }`
			native, proven := provenProbe(t, src)
			if native != c.native || proven != c.proven {
				t.Errorf("native %s, proven %s; want %s, %s", native, proven, c.native, c.proven)
			}
		})
	}
}

// TestForeachTargetWrite covers a foreach loop writing its value (or key)
// into an element (`foreach ($xs as $a['k'])`): after a loop that runs,
// the variable is an array, and after one that may not, it may also keep
// what it held.
func TestForeachTargetWrite(t *testing.T) {
	const head = `<?php function f(array $xs, bool $c) { $a = null; `
	cases := []struct {
		name, body, want string
	}{
		{"array literal", `foreach ([1, 2] as $a['k']) {} t('v', $a); }`, `array`},
		{"key target", `foreach ([1] as $a['k'] => $v) {} t('v', $a); }`, `array`},
		{"destructured", `foreach ([[1]] as [$a['k']]) {} t('v', $a); }`, `array`},
		{"list", `foreach ([[1]] as list($a['k'])) {} t('v', $a); }`, `array`},
		{"parenthesized", `foreach ([1] as ($a)['k']) {} t('v', $a); }`, `array`},
		{"nested", `foreach ([1] as $a['k']['l']) {} t('v', $a); }`, `array`},
		{"non-empty value", `$ys = [1]; $ys[] = 2; foreach ($ys as $a[]) {} t('v', $a); }`, `array`},
		{"in the body", `foreach ($xs as $a['k']) { t('v', $a); } }`, `array`},
		// Not proven.
		{"may be empty", `foreach ($xs as $a['k']) {} t('v', $a); }`, `array|null`},
		{"empty literal", `foreach ([] as $a['k']) {} t('v', $a); }`, `array|null`},
		{"unpacked only", `foreach ([...$xs] as $a['k']) {} t('v', $a); }`, `array|null`},
		{"reset in the body", `foreach ([1] as $a['k']) { $a = null; } t('v', $a); }`, `array|null`},
		{"unset in the body", `foreach ([1] as $a['k']) { unset($a); } t('v', $a); }`, `array|null`},
		{"in a branch", `if ($c) { foreach ([1] as $a['k']) {} } t('v', $a); }`, `array|null`},
		{"not in a block", `if ($c) foreach ([1] as $a['k']) {} t('v', $a); }`, `array|null`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			native, proven := provenProbe(t, head+c.body)
			if native != c.want || proven != c.want {
				t.Errorf("native %s, proven %s; want %s", native, proven, c.want)
			}
		})
	}
}

// TestPropertyWholeDestructuring covers a destructuring that writes both
// into an offset of a property and the whole property: the statement
// resets it, so neither the offset write nor an earlier guard holds past
// it.
func TestPropertyWholeDestructuring(t *testing.T) {
	checkAnywhere(t, `<?php
class Q {
    private ?array $h = null;
    public static ?array $s = null;
    function m1() { [$this->h['a'], $this->h] = [1, null]; t('whole', $this->h); }
    function m2() { $this->h['a'] = 1; [$this->h['b'], $this->h] = [1, null]; t('wholeAfterWrite', $this->h); }
    function m3() { if ($this->h === null) { return; } [$this->h] = [null]; t('guardReset', $this->h); }
    function m4() { if ($this->h === null) { return; } ['k' => [$x, $this->h]] = ['k' => [1, null]]; t('nestedReset', $this->h); }
    function m5() { [self::$s['a'], self::$s] = [1, null]; t('static', self::$s); }
    function m6() { if ($this->h === null) { return; } [$x, $y] = [1, 2]; t('otherList', $this->h); }
}
`, map[string]string{
		"whole": "array|null", "wholeAfterWrite": "array|null", "guardReset": "array|null", "nestedReset": "array|null",
		"static": "array|null", "otherList": "array",
	})
}

// TestParenthesizedOffsetWrites covers writes into an offset of a
// parenthesized variable or property (`($a)['k'] = 1`), which PHP accepts
// as plain writes into its offsets.
func TestParenthesizedOffsetWrites(t *testing.T) {
	checkAnywhere(t, `<?php
class Q {
    private ?array $h = null;
    function m1() { ($this->h)['a'] = 1; t('prop', $this->h); }
    function m2() { [($this->h)['a']] = [1]; t('propList', $this->h); }
    function m3() { ($this->h)['a']++; t('propInc', $this->h); }
}
function a(?array $n) { ($n)['k'] = 1; t('local', $n); }
function b(?array $n) { (($n)['a'])['b'] = 1; t('nested', $n); }
function c(?array $n) { [($n)['k']] = [1]; t('list', $n); }
function d(?array $n) { ($n)['k']++; t('inc', $n); }
`, map[string]string{
		"prop": "array", "propList": "array", "propInc": "array",
		"local": "array", "nested": "array", "list": "array", "inc": "array",
	})
}
