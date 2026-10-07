package infer_test

import "testing"

func TestMethodTemplates(t *testing.T) {
	lib := `<?php
namespace Lib;
/** @template T of object */
class Repo {
    /** @return T|null */
    public function find(int $id) { return null; }
}
class Entity {}
class User extends Entity {}
interface Stub {}
class Manager {
    /**
     * @template T of object
     * @param class-string<T> $className
     * @return Repo<T>
     */
    public function getRepository(string $className): Repo { return new Repo(); }
    /**
     * @template T
     * @param class-string<T> $id
     * @return T
     */
    public function get(string $id): ?object { return null; }
    /**
     * @psalm-template RealInstanceType of object
     * @psalm-param class-string<RealInstanceType> $type
     * @psalm-return RealInstanceType&Stub
     */
    public static function createStub(string $type): Stub {}
    /**
     * @template T
     * @param T $value
     * @return T
     */
    public function identity($value) { return $value; }
    /**
     * @template T
     * @param ?T $value
     * @param T $default
     * @return T
     */
    public function orDefault($value, $default) { return $value ?? $default; }
    /**
     * @template T
     * @param array<T> $xs
     * @return T
     */
    public function firstOf(array $xs) { return reset($xs); }
    /**
     * @template K
     * @template V
     * @param iterable<K, V> $xs
     * @return array<K, V>
     */
    public function toArray(iterable $xs): array { return []; }
    /**
     * @template T
     * @param class-string<T> $c
     * @return int
     */
    public function declaredWins(string $c): int { return 1; }
    /**
     * @template T
     * @param class-string<T> $c
     * @return T
     */
    public function mismatch(string $c): \Lib\Entity {}
}
/**
 * @template T
 * @param list<T> $xs
 * @return T|null
 */
function head(array $xs) { return $xs[0] ?? null; }
`
	checkWith(t, map[string]string{"lib.php": lib}, `<?php
use Lib\{Manager, User, Entity};
function run(Manager $m, string $name, array $plain) {
    $users = [new User(), new User()];
    $cls = User::class;
    t('repo', $m->getRepository(User::class));
    t('find', $m->getRepository(User::class)->find(1));
    t('get', $m->get(User::class));
    t('getVar', $m->get($cls));
    t('getNamed', $m->get(id: User::class));
    t('getDynamic', $m->get($name));
    t('stub', Manager::createStub(User::class));
    t('identity', $m->identity(1));
    t('orDefault', $m->orDefault(null, 'x'));
    t('firstOf', $m->firstOf($users));
    t('firstOfPlain', $m->firstOf($plain));
    t('shapeFirst', $m->firstOf(['a' => 1, 'b' => 'x']));
    t('toArray', $m->toArray($users));
    t('head', \Lib\head($users));
    t('declaredWins', $m->declaredWins(User::class));
    t('mismatch', $m->mismatch(\stdClass::class));
    t('subtypeOk', $m->mismatch(User::class));
    t('classConst', User::class);
}
`, map[string]string{
		"repo":         `\Lib\Repo<\Lib\User>`,
		"find":         `\Lib\User|null`,
		"get":          `\Lib\User`,
		"getVar":       `\Lib\User`,
		"getNamed":     `\Lib\User`,
		"getDynamic":   "null|object",
		"stub":         `\Lib\Stub|\Lib\User`,
		"identity":     "int",
		"orDefault":    "string",
		"firstOf":      `\Lib\User`,
		"firstOfPlain": "mixed",
		"shapeFirst":   "int|string",
		"toArray":      `\Lib\User[]`,
		"head":         `\Lib\User|null`,
		"declaredWins": "int",
		"mismatch":     `\Lib\Entity`,
		"subtypeOk":    `\Lib\User`,
		"classConst":   `class-string<\Lib\User>`,
	})
}
