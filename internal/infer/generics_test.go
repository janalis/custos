package infer_test

import "testing"

func TestGenericTemplates(t *testing.T) {
	lib := `<?php
namespace Lib;
/**
 * @template TKey of array-key
 * @template-covariant T
 * @template-extends \IteratorAggregate<TKey, T>
 */
interface ReadableCollection extends \IteratorAggregate {
    /**
     * @return mixed
     * @phpstan-return T|false
     */
    public function first();
    /** @return T|null */
    public function get($key);
    /** @return T[] */
    public function toArray();
    /** @return list<T> */
    public function getValues();
    /** @return TKey|null */
    public function key();
    /**
     * @template U
     * @return ReadableCollection<TKey, U>
     */
    public function map(\Closure $f);
}
/**
 * @phpstan-template TKey of array-key
 * @phpstan-template T
 * @template-extends ReadableCollection<TKey, T>
 */
interface Collection extends ReadableCollection {}
/**
 * @template TKey of array-key
 * @template T
 * @implements Collection<TKey, T>
 */
class ArrayCollection implements Collection {}
/** @template T of Entity */
class Repository {
    /** @return ?T */
    public function find(int $id) { return null; }
    /** @return T[] */
    public function findAll(): array { return []; }
}
class Entity {}
class User extends Entity {}
/** @extends Repository<User> */
class UserRepository extends Repository {
    public function first() { return parent::find(1); }
}
/** @template-implements \IteratorAggregate<int, User> */
class Users implements \IteratorAggregate {}
/**
 * @template T
 */
class Box {
    /** @return T */
    public function get() {}
    /** @return static */
    public function self() { return $this; }
}
`
	checkWith(t, map[string]string{"lib.php": lib}, `<?php
use Lib\{Collection, ArrayCollection, UserRepository, Repository, Users, Box, User};
class Holder {
    /** @var Collection<int, User> */
    private Collection $users;
    /** @var Collection<User>|User[] */
    private $mixedUsers;
    public function run(UserRepository $repo, Repository $plain, Users $list, \ArrayIterator $raw) {
        t('first', $this->users->first());
        t('get', $this->users->get(1));
        t('toArray', $this->users->toArray());
        t('values', $this->users->getValues());
        t('key', $this->users->key());
        t('map', $this->users->map(fn ($u) => $u));
        foreach ($this->users as $k => $u) { t('fkey', $k); t('fval', $u); }
        foreach ($this->mixedUsers as $u) { t('mixedVal', $u); }
        t('find', $repo->find(1));
        t('findAll', $repo->findAll());
        t('parent', $repo->first());
        t('plainFind', $plain->find(1));
        foreach ($list as $k => $u) { t('aggKey', $k); t('aggVal', $u); }
        foreach ($raw as $v) { t('rawVal', $v); }
        /** @var \ArrayIterator<string, User> $it */
        $it = f();
        foreach ($it as $k => $v) { t('itKey', $k); t('itVal', $v); }
        /** @var \Generator<int, User> $gen */
        $gen = f();
        foreach ($gen as $v) { t('genVal', $v); }
        /** @var iterable<string, User> $iter */
        $iter = f();
        foreach ($iter as $k => $v) { t('iterKey', $k); t('iterVal', $v); }
        /** @var Box<User> $box */
        $box = f();
        t('box', $box->get());
        t('boxSelf', $box->self()->get());
        if ($box instanceof Box) { t('narrowed', $box->get()); }
        /** @var Collection<int, User>|null $maybe */
        $maybe = f();
        t('nullsafe', $maybe?->first());
        $doc = new \DOMDocument();
        foreach ($doc->getElementsByTagName('a') as $el) { t('dom', $el); }
    }
}
`, map[string]string{
		"first":     `\Lib\User|false`,
		"get":       `\Lib\User|null`,
		"toArray":   `\Lib\User[]`,
		"values":    `\Lib\User[]`,
		"key":       "int|null",
		"map":       `\Lib\ReadableCollection<int, mixed>`,
		"fkey":      "int",
		"fval":      `\Lib\User`,
		"mixedVal":  `\Lib\User`,
		"find":      `\Lib\User|null`,
		"findAll":   `\Lib\User[]`,
		"parent":    "?unknown",   // UserRepository::first is virtual and untyped
		"plainFind": "mixed|null", // unbound: as before (the bound is not used)
		"aggKey":    "int",
		"aggVal":    `\Lib\User`,
		"rawVal":    "?unknown",
		"itKey":     "string",
		"itVal":     `\Lib\User`,
		"genVal":    `\Lib\User`,
		"iterKey":   "string",
		"iterVal":   `\Lib\User`,
		"box":       `\Lib\User`,
		"boxSelf":   "mixed", // `static` loses the arguments
		"narrowed":  `\Lib\User`,
		"nullsafe":  `\Lib\User|false|null`,
		"dom":       `\DOMElement`,
	})
}
