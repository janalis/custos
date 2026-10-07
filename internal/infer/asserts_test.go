package infer_test

import "testing"

func TestAssertions(t *testing.T) {
	lib := `<?php
namespace Lib;
class Assert {
    /** @psalm-assert string $value */
    public static function string($value, $message = '') {}
    /** @psalm-assert !null $value */
    public static function notNull($value) {}
    /** @psalm-assert iterable<string> $value */
    public static function allString($value) {}
    /**
     * @psalm-template ExpectedType of object
     * @psalm-param class-string<ExpectedType> $class
     * @psalm-assert ExpectedType $value
     */
    public static function isInstanceOf($value, $class) {}
    /**
     * @template ExpectedType of object
     * @param class-string<ExpectedType> $expected
     * @phpstan-assert =ExpectedType $actual
     */
    final public static function assertInstanceOf(string $expected, mixed $actual): void {}
}
class Base {
    /** @phpstan-assert-if-true Child $this */
    public function isChild(): bool { return false; }
    public ?string $name = null;
    /** @phpstan-assert-if-true !null $this->name */
    public function hasName(): bool { return $this->name !== null; }
    /** @phpstan-assert-if-false string $this->name */
    public function nameless(): bool { return $this->name === null; }
}
class Child extends Base { public function only(): int { return 1; } }
/** @phpstan-assert-if-true int $x */
function isInt(mixed $x): bool { return \is_int($x); }
/** @phpstan-assert-if-false null $x */
function isSet2($x): bool { return $x !== null; }
`
	checkWith(t, map[string]string{"lib.php": lib}, `<?php
use Lib\{Assert, Base, Child};
class T extends \Lib\Base {
    public function run(?string $s, int|string $u, Base $b, ?Base $nb, array $xs, int|string|null $v, Base $c) {
        Assert::string($u);
        t('string', $u);
        Assert::notNull($s);
        t('notNull', $s);
        Assert::isInstanceOf($nb, Child::class);
        t('instance', $nb);
        Assert::assertInstanceOf(Child::class, $b);
        t('phpunit', $b);
        t('notYet', $v);
        if ($c->isChild()) { t('ifTrue', $c); } else { t('ifFalse', $c); }
        if (\Lib\isInt($v)) { t('fnTrue', $v); } else { t('fnFalse', $v); }
        if (!\Lib\isInt($v)) { return; }
        t('guard', $v);
        if ($this->hasName()) { t('prop', $this->name); }
        if (!$this->nameless()) { t('propFalse', $this->name); }
        $w = $s;
        Assert::string($w);
        $w = null;
        t('reassigned', $w);
    }
    public function other(?string $s, $unknown) {
        $x = $s;
        if (\Lib\isSet2($x)) { t('isSetTrue', $x); } else { t('isSetFalse', $x); }
        Assert::string($unknown);
        t('unknown', $unknown);
    }
}
`, map[string]string{
		"string":     "string",
		"notNull":    "string",
		"instance":   `\Lib\Child`,
		"phpunit":    `\Lib\Child`,
		"notYet":     "int|null|string",
		"ifTrue":     `\Lib\Child`,
		"ifFalse":    `\Lib\Base`,
		"fnTrue":     "int",
		"fnFalse":    "int|null|string",
		"guard":      "int",
		"prop":       "string",
		"propFalse":  "string",
		"reassigned": "null",
		"isSetTrue":  "null|string",
		"isSetFalse": "null",
		"unknown":    "?unknown",
	})
}
