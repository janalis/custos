package infer_test

import "testing"

// TestAssertionEdges covers calls whose assertions cannot apply (first-class
// callables, dynamic names, unresolvable classes, unbound templates), class
// templates in assertions, self::/parent:: receivers and the bool/array
// set arithmetic of positive and negated assertions.
func TestAssertionEdges(t *testing.T) {
	lib := `<?php
namespace Lib;
class Assert {
    /** @psalm-assert string $value */
    public static function string($value) {}
    /** @psalm-assert mixed $value */
    public static function anything($value) {}
    /** @psalm-assert bool $value */
    public static function bool($value) {}
    /** @psalm-assert iterable<string> $value */
    public static function allString($value) {}
    /** @psalm-assert !bool $value */
    public static function notBool($value) {}
    /** @psalm-assert !false $value */
    public static function notFalse($value) {}
    /** @psalm-assert !true $value */
    public static function notTrue($value) {}
    /** @psalm-assert !int $value */
    public static function notInt($value) {}
    /**
     * @template T of object
     * @param class-string<T> $class
     * @psalm-assert T $value
     */
    public static function isInstanceOf($value, $class) {}
}
/** @template T */
class Checker {
    /** @phpstan-assert T $x */
    public function check($x): void {}
}
class Base {
    public ?string $name = null;
    /** @phpstan-assert-if-true !null $this->name */
    public function hasName(): bool { return $this->name !== null; }
}
class Child extends Base {}
/** @phpstan-assert-if-true int $x */
function isInt(mixed $x): bool { return \is_int($x); }
`
	checkWith(t, map[string]string{"lib.php": lib}, `<?php
use Lib\{Assert, Base, Child, Checker};
class T extends Base {
    /** @param Checker<int> $ck */
    public function run(int|string $a, int|string $b, int|string $c, int|string $d, ?Base $e, string $cls,
            $name, Checker $ck, int|string $f, int|true $g, array|int $h, bool|int $i, bool|int $j,
            bool|string $k, int $l, int|string $m, mixed $n) {
        Assert::string(...);
        Assert::string($a);
        t('static', $a);
        $fn = \Lib\isInt(...);
        $ck->check(...);
        $ck->$name($b);
        t('dynamic', $b);
        $cls::string($c);
        t('dynClass', $c);
        Assert::isInstanceOf($e, $cls);
        t('unbound', $e);
        $ck->check($f);
        t('classTpl', $f);
        Assert::anything($d);
        t('mixed', $d);
        Assert::bool($g);
        t('bool', $g);
        Assert::allString($h);
        t('iterable', $h);
        Assert::notBool($i);
        t('notBool', $i);
        Assert::notFalse($j);
        t('notFalse', $j);
        Assert::notTrue($k);
        t('notTrue', $k);
        Assert::notInt($l);
        t('notInt', $l);
        if (parent::hasName()) { t('parentProp', $this->name); }
        if (static::hasName()) { t('staticProp', $this->name); }
        if (Base::hasName()) { t('namedProp', $this->name); }
        if (\Lib\isInt(...) && $m) { t('fccFunc', $m); }
        if (Assert::string(...) && $m) { t('fccStatic', $m); }
        $either = $l ? $ck : new Base();
        $either->check($m);
        t('mixedRecv', $m);
        $name->check($m);
        $ck->undeclared($m);
        t('noMethod', $m);
        $name::string($m);
        t('untypedClass', $m);
        if (\Lib\isInt($n)) { t('mixedArg', $n); }
    }
}
`, map[string]string{
		"static": "string", "dynamic": "int|string", "dynClass": "int|string",
		"unbound": `\Lib\Base|null`, "classTpl": "int", "mixed": "int|string",
		"bool": "true", "iterable": "array", "notBool": "int", "notFalse": "int|true",
		"notTrue": "false|string", "notInt": "int", "parentProp": "string", "staticProp": "string", "namedProp": "null|string",
		"mixedRecv": "int|string", "noMethod": "int|string", "untypedClass": "int|string", "fccFunc": "int|string", "fccStatic": "int|string",
		"mixedArg": "int",
	})
}
