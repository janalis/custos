package infer_test

import "testing"

// De Morgan: the false branch of `A && B` is `!A || !B`, the true branch
// of `A || B` is `A || (!A && B)`; in guards, else branches, ternaries,
// loops and nested combinations.
func TestNarrowDeMorgan(t *testing.T) {
	checkAnywhere(t, `<?php
/**
 * @param int|string|array|bool|float $k
 * @param int|string|null $x
 * @param \Foo|\Bar|null $o
 */
function f($k, $x, $o, $c) {
    if (!is_scalar($k) && !$k instanceof \Stringable) { throw new \Exception(); }
    t('guard', $k);
    if (is_int($x) && $c) { t('andTrue', $x); } else { t('andElse', $x); }
    if ($x === null || is_string($x)) { t('orTrue', $x); } else { t('orElse', $x); }
    t('ternary', ($x !== null && !is_int($x)) ? 1 : $x);
    while (!(is_string($x) || $x === null)) { t('while', $x); }
    for ($i = 0; $x !== null && $i < 3; $i++) { t('for', $x); }
    if (!($o instanceof \Foo && $c) && !($o === null)) { t('nested', $o); }
    if ($o === null || $o instanceof \Foo) { return; }
    t('orGuard', $o);
}
function g(?string $s, $c) {
    if (!$c && $s === null) { return; }
    t('unrelated', $s);
    if (!is_string($s) || $s === '') { return; }
    t('orGuard2', $s);
}
function h(int|string $v) {
    if (is_int($v) && is_string($v)) { t('impossible', $v); } else { t('impossibleElse', $v); }
}
`, map[string]string{
		"guard":          "bool|float|int|string", // arrays are no Stringable
		"andTrue":        "int",
		"andElse":        "int|null|string",
		"orTrue":         "null|string",
		"orElse":         "int",
		"ternary":        "int|null",
		"while":          "int",
		"for":            "int|string",
		"nested":         `\Bar|\Foo`,
		"orGuard":        `\Bar`,
		"unrelated":      "null|string",
		"orGuard2":       "string",
		"impossibleElse": "int|string",
	})
}

// The Symfony ParameterBag::resolveValue case: the guard rejects anything
// that is neither scalar nor Stringable.
func TestNarrowDeMorganMixed(t *testing.T) {
	checkAnywhere(t, `<?php
function f(mixed $k) {
    if (!is_scalar($k) && !$k instanceof \Stringable) {
        throw new \RuntimeException('bad key');
    }
    t('key', $k);
}
function g(array|int|string $k) {
    if (!\is_scalar($k) && !$k instanceof \Stringable) {
        throw new \RuntimeException('bad key');
    }
    t('key2', $k);
}
`, map[string]string{
		"key":  `\Stringable|bool|false|float|int|string|true`, // mixed passing is_scalar()
		"key2": "int|string",                                   // an array is never Stringable
	})
}

func TestNarrowElseIfChains(t *testing.T) {
	checkAnywhere(t, `<?php
function f(int|string|null|array $x) {
    if ($x === null) {
    } elseif (is_int($x)) {
        t('second', $x);
    } elseif (is_string($x)) {
        t('third', $x);
    } else {
        t('else', $x);
    }
    if (is_int($x)) {
    } else if (is_string($x)) {
        t('elseIfSpaced', $x);
    } else {
        t('elseSpaced', $x);
    }
}
`, map[string]string{
		"second":       "int",
		"third":        "string",
		"else":         "array",
		"elseIfSpaced": "string",
		"elseSpaced":   "array|null",
	})
}

func TestNarrowMatch(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
enum Kind { case A; case B; }
/** @param int|string|Foo|null $x */
function f($x, Kind|int $k) {
    $a = match (true) {
        is_string($x) => t('str', $x),
        $x instanceof Foo, $x === null => t('fooOrNull', $x),
        default => t('default', $x),
    };
    $b = match ($x) {
        null => t('null', $x),
        default => t('nonNull', $x),
    };
    $c = match (true) {
        is_int($x) => 1,
        t('armCond', $x) => 2,
        default => 3,
    };
    $d = match ($k) {
        Kind::A => t('enum', $k),
        1, 2 => t('ints', $k),
        default => t('kdefault', $k),
    };
    $e = match (true) {
        (bool) $x => 1,
        default => t('nonBool', $x),
    };
    $f = match (gettype($x)) {
        'string' => t('gettype', $x),
        'NULL' => t('gettypeNull', $x),
        default => t('gettypeDefault', $x),
    };
}
`, map[string]string{
		"str":            "string",
		"fooOrNull":      `\Foo|null`,
		"default":        "int",
		"null":           "null",
		"nonNull":        `\Foo|int|string`,
		"armCond":        `\Foo|null|string`,
		"enum":           `\Kind`,
		"ints":           "int",
		"kdefault":       `\Kind|int`,
		"nonBool":        `\Foo|int|null|string`, // (bool) $x is a cast, not a condition on $x
		"gettype":        "string",
		"gettypeNull":    "null",
		"gettypeDefault": `\Foo|int`,
	})
}

func TestNarrowSwitch(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
/** @param int|string|Foo|null $x */
function f($x) {
    switch (true) {
        case is_string($x):
            t('str', $x);
            break;
        case is_int($x):
        case $x === null:
            t('group', $x);
            break;
        default:
            t('default', $x);
    }
    switch (true) {
        case is_string($x):
            t('fall1', $x);
        case is_int($x):
            t('fallen', $x);
            break;
    }
    switch ($x) {
        case null:
            t('loose', $x);
            break;
        default:
            t('notNull', $x);
    }
    switch (true) {
        case is_int($x):
            break;
        default:
        case is_string($x):
            t('withDefault', $x);
    }
    switch (true) {
        case is_int($x):
            $y = 1;
        default:
            t('fallDefault', $x);
    }
}
`, map[string]string{
		"str":         "string",
		"group":       "int|null",
		"default":     `\Foo`,
		"fall1":       "string",
		"fallen":      `\Foo|int|null|string`,
		"loose":       `\Foo|int|null|string`,
		"notNull":     `\Foo|int|string`,
		"withDefault": `\Foo|int|null|string`,
		"fallDefault": `\Foo|int|null|string`,
	})
}

func TestNarrowNullsafeAndChains(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo { public ?Foo $next = null; public array $items = []; public function ok(): bool { return true; } }
function f(?Foo $x, ?Foo $y, ?Foo $z, ?array $a, ?Foo $w, ?Foo $v) {
    if ($x?->ok()) { t('nullsafe', $x); }
    if ($y?->next !== null) { t('neNull', $y); }
    if (isset($z->next->items)) { t('isset', $z); }
    if ($a[0] ?? false) {}
    if (!empty($a['k'])) { t('emptyDim', $a); }
    if ($w !== null && $w->ok()) { t('andRhs', $w); }
    if ($v?->next instanceof Foo) { t('instChain', $v); }
    if (!$x?->ok()) { t('negNullsafe', $x); }
}
`, map[string]string{
		"nullsafe":    `\Foo`,
		"neNull":      `\Foo`,
		"isset":       `\Foo`,
		"emptyDim":    "non-empty array",
		"andRhs":      `\Foo`,
		"instChain":   `\Foo`,
		"negNullsafe": `\Foo|null`,
	})
}

func TestNarrowTruthinessBool(t *testing.T) {
	checkAnywhere(t, `<?php
function f(?bool $b, bool|string $s, int|true $i) {
    if ($b) { t('truthy', $b); } else { t('falsy', $b); }
    if (!empty($s)) { t('notEmpty', $s); }
    if (empty($s)) { t('empty', $s); }
    if (!$i) { t('noTrue', $i); }
}
`, map[string]string{
		"truthy":   "true",
		"falsy":    "false|null",
		"notEmpty": "string|true",
		"empty":    "false|string",
		"noTrue":   "int",
	})
}

func TestNarrowThisPropInstanceof(t *testing.T) {
	checkAnywhere(t, `<?php
interface Node {}
class Leaf implements Node {}
class Tree {
    private ?Node $root = null;
    public function f() {
        if ($this->root instanceof Leaf) { t('prop', $this->root); }
        if (!$this->root instanceof Leaf) { return; }
        t('propGuard', $this->root);
    }
}
`, map[string]string{
		"prop":      `\Leaf`,
		"propGuard": `\Leaf`,
	})
}

func TestNarrowTypeGuards(t *testing.T) {
	checkAnywhere(t, `<?php
class Base {}
class Sub extends Base {}
final class Fin {}
/** @param Base|Fin|string|null $x */
function f($x, mixed $m, int|string|null $s, ?array $a) {
    if (is_a($x, Base::class)) { t('isA', $x); } else { t('notIsA', $x); }
    if (is_a($x, 'Base', true)) { t('isAString', $x); }
    if (is_subclass_of($x, Base::class)) { t('subclass', $x); } else { t('notSubclass', $x); }
    if (is_subclass_of($x, Base::class, false)) { t('subclassObj', $x); }
    if (gettype($m) === 'integer') { t('gettype', $m); }
    if (gettype($s) !== 'NULL') { t('gettypeNot', $s); }
    if ('string' == gettype($s)) { t('gettypeLoose', $s); }
    if (gettype($x) === 'resource') { t('gettypeRes', $x); }
    if (gettype($x) !== 'resource') { t('gettypeNotRes', $x); }
    if (gettype($x) === 'nonsense') { t('gettypeBad', $x); }
    if (get_debug_type($m) === 'float') { t('debugType', $m); }
    if (get_debug_type($x) === Fin::class) { t('debugClass', $x); }
    if (get_debug_type($x) !== 'Fin') { t('debugNotClass', $x); }
    if ($x::class === Sub::class) { t('classConst', $x); }
    if (get_class($x) === 'Sub') { t('getClass', $x); }
    if (get_class($x) !== Sub::class) { t('getClassNot', $x); }
    if (count($a) > 0) { t('count', $a); }
    if (in_array($s, ['a', 'b'], true)) { t('inArray', $s); }
    if (in_array($m, [1, null], true)) { t('inArrayMixed', $m); }
    if (in_array($s, ['a', 'b'])) { t('inArrayLoose', $s); }
    if (!in_array($s, ['a'], true)) { t('notInArray', $s); }
    if (in_array($s, $a, true)) { t('inArrayPlain', $s); }
    if (true === is_string($s)) { t('trueIdent', $s); }
    if (false === is_string($s)) { t('falseIdent', $s); }
    if (true !== is_string($s)) { t('trueNotIdent', $s); }
    if (false != is_int($s)) { t('falseNe', $s); }
    if (true !== $m) {}
    if ($s === 5) { t('identInt', $s); }
    if ($s === -1) { t('identNeg', $s); }
    if ($m === 1.5) { t('identFloat', $m); }
}
`, map[string]string{
		"isA":           `\Base`,
		"notIsA":        `\Fin|null|string`,
		"isAString":     `\Base|string`,
		"subclass":      `\Base|string`,
		"notSubclass":   `\Base|\Fin|null|string`,
		"subclassObj":   `\Base`,
		"gettype":       "int",
		"gettypeNot":    "int|string",
		"gettypeLoose":  "string",
		"gettypeRes":    "resource",
		"gettypeNotRes": `\Base|\Fin|null|string`,
		"gettypeBad":    `\Base|\Fin|null|string`,
		"debugType":     "float",
		"debugClass":    `\Fin`,
		"debugNotClass": `\Base|\Fin|null|string`,
		"classConst":    `\Sub`,
		"getClass":      `\Sub`,
		"getClassNot":   `\Base|\Fin|null|string`,
		"count":         "non-empty array",
		"inArray":       "string",
		"inArrayMixed":  "int|null",
		"inArrayLoose":  "int|null|string",
		"notInArray":    "int|null|string",
		"inArrayPlain":  "int|null|string",
		"trueIdent":     "string",
		"falseIdent":    "int|null",
		"trueNotIdent":  "int|null",
		"falseNe":       "int",
		"identInt":      "int",
		"identNeg":      "int",
		"identFloat":    "float",
	})
}

// A definition in a switch case that ends with break does not reach the
// later cases (except through an enclosing loop's back edge).
func TestSwitchCaseReaching(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string|int|null $d, $c) {
    switch ($c) {
        case 1:
            $d = explode(',', $d);
            break;
        case 2:
            t('afterBreak', $d);
            $d = 1.5;
        case 3:
            t('fallThrough', $d);
            break;
        case 4:
            if ($c) { $d = [1]; break; }
            $d = true;
            goto out;
        default:
            t('afterGoto', $d);
    }
    out:
    foreach ([1, 2] as $i) {
        switch ($i) {
            case 1:
                t('loop', $d);
                break;
            case 2:
                $d = false;
                break;
        }
    }
}
`, map[string]string{
		"afterBreak":  "int|null|string",
		"fallThrough": "float|int|null|string",
		"afterGoto":   "int|null|string|true", // [1] is followed by break
		"loop":        "false|float|int|int[]|null|string|string[]|true",
	})
}
