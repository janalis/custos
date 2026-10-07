package infer_test

import "testing"

// TestMethodTemplateEdges covers the binding paths that must fall back to
// the call's untemplated type (unbound, ambiguous or unsupported patterns)
// and the class-template substitution inside a method-level return type.
func TestMethodTemplateEdges(t *testing.T) {
	lib := `<?php
namespace Lib;
class Entity {}
class User extends Entity {}
/** @template T */
class Box {}
/** @template T of Entity */
class Coll {
    /**
     * @template K
     * @param K $k
     * @return K|T
     */
    public function keyed($k) { return $k; }
}
/** @template T */
class Free {
    /**
     * @template K
     * @param K $k
     * @return K|T
     */
    public function either($k) { return $k; }
}
class M {
    /**
     * @template T
     * @param T ...$xs
     * @return T
     */
    public function variadic(...$xs) { return $xs[0]; }
    /**
     * @template T
     * @param T $value
     * @return T
     */
    public function identity($value) { return $value; }
    /**
     * @template T
     * @template U
     * @param T $a
     * @return T|U
     */
    public function half($a) { return $a; }
    /**
     * @template T
     * @template U
     * @param T|U $x
     * @return T
     */
    public function ambiguous($x) { return $x; }
    /**
     * @template T
     * @param T|T[] $x
     * @return T
     */
    public function oneOrMany($x) { return $x; }
    /**
     * @template T
     * @param array<T>|null $xs
     * @return T
     */
    public function maybeList(?array $xs) { return null; }
    /**
     * @template T
     * @param array<T> $xs
     * @return T
     */
    public function firstOf($xs) { return null; }
    /**
     * @template T
     * @param iterable<T> $xs
     * @return T
     */
    public function firstIt(iterable $xs) { return null; }
    /**
     * @template K
     * @template V
     * @param iterable<K, V> $xs
     * @return array<K, V>
     */
    public function toArray(iterable $xs): array { return []; }
    /**
     * @template T
     * @param Box<T> $b
     * @return T
     */
    public function unbox(Box $b) { return null; }
    /**
     * @template T
     * @param T $x
     * @return T
     */
    public function asBool($x): bool { return true; }
    /**
     * @template T
     * @param T $x
     * @return T
     */
    public function asInt($x): int { return 1; }
    /**
     * @template T
     * @param mixed $x
     * @phpstan-assert T $x
     */
    public function assertOnly($x): void {}
}
`
	checkWith(t, map[string]string{"lib.php": lib}, `<?php
use Lib\{M, User, Box, Coll, Free};
/**
 * @param Box<User> $box
 * @param Coll<User> $cu
 * @param Free<User> $fu
 */
function run(M $m, Box $box, Box $plainBox, Coll $c, Coll $cu, Free $f, Free $fu, array $xs, $untyped) {
    t('variadic', $m->variadic(1, 2));
    t('unpack', $m->identity(...$xs));
    t('named', $m->identity(other: 1));
    t('missing', $m->identity());
    t('untyped', $m->identity($untyped));
    t('half', $m->half(1));
    t('ambiguous', $m->ambiguous(1));
    t('oneOrMany', $m->oneOrMany(1));
    t('maybeList', $m->maybeList([1]));
    t('notArray', $m->firstOf('s'));
    t('shapeIt', $m->firstIt(['a' => 1]));
    t('mixedKey', $m->toArray($xs));
    t('unbox', $m->unbox($box));
    t('plainBox', $m->unbox($plainBox));
    t('bool', $m->asBool(true));
    t('intMismatch', $m->asInt('s'));
    t('assertOnly', $m->assertOnly(1));
    t('classBound', $c->keyed(1));
    t('classArg', $cu->keyed(1));
    t('classFree', $f->either(1));
    t('classFreeArg', $fu->either(1));
}
`, map[string]string{
		"variadic":     `mixed`,
		"unpack":       `mixed`,
		"named":        `mixed`,
		"missing":      `mixed`,
		"untyped":      `mixed`,
		"half":         `mixed`,
		"ambiguous":    `mixed`,
		"oneOrMany":    `mixed`,
		"maybeList":    `mixed`,
		"notArray":     `mixed`,
		"shapeIt":      `int`,
		"mixedKey":     `mixed[]`,
		"unbox":        `\Lib\User`,
		"plainBox":     `mixed`,
		"bool":         `true`,
		"intMismatch":  `int`,
		"assertOnly":   `void`,
		"classBound":   `\Lib\Entity|int`,
		"classArg":     `\Lib\User|int`,
		"classFree":    `int|mixed`,
		"classFreeArg": `\Lib\User|int`,
	})
}
