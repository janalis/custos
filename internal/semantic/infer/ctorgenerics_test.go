package infer_test

import "testing"

// `new` binds the class templates its constructor documents from the
// arguments.
func TestConstructorBindsClassTemplates(t *testing.T) {
	checkWith(t, nil, `<?php
class Foo {}
/** @template T */
class Box {
    /** @param T $value */
    public function __construct(public $value) {}
    /** @return T */
    public function get() { return $this->value; }
}
/**
 * @template TKey of array-key
 * @template T
 * @implements IteratorAggregate<TKey, T>
 */
class Coll implements IteratorAggregate {
    /** @param array<TKey, T> $items */
    public function __construct(private array $items = []) {}
    public function getIterator(): Iterator { return new ArrayIterator($this->items); }
}
/** @template T of Foo */
class Bounded {
    /** @param T|null $v @param int $n */
    public function __construct($v = null, int $n = 0) {}
}
/**
 * @template K of array-key
 * @template V
 */
class Pair {
    /** @param V $v */
    public function __construct($v) {}
}
/** @template T */
class Sub extends Box {}
/** @template T */
class Shadow {
    /**
     * @template T
     * @param T $v
     */
    public function __construct($v) {}
}
/** @template T */
class Plain { public function __construct($v) {} }
function run(Foo $f, $untyped) {
    t('box', new Box($f));
    t('get', (new Box($f))->get());
    t('named', new Box(value: $f));
    t('coll', new Coll(['a' => $f]));
    foreach (new Coll([$f]) as $k => $v) { t('iterValue', $v); t('iterKey', $k); }
    t('empty', new Coll());
    t('bounded', new Bounded(null, 1));
    t('boundedArg', new Bounded($f));
    t('sub', new Sub($f));
    t('pair', new Pair($f));
    t('shadow', new Shadow($f));
    t('plain', new Plain($f));
    t('unknownArg', new Box($untyped));
}
`, map[string]string{
		"box": `\Box<\Foo>`, "get": `\Foo`, "named": `\Box<\Foo>`,
		"coll": `\Coll<string, \Foo>`, "iterValue": `\Foo`, "iterKey": "int",
		"empty": `\Coll`, "bounded": `\Bounded`, "boundedArg": `\Bounded<\Foo>`,
		"sub": `\Sub`, "pair": `\Pair<int|string, \Foo>`, "shadow": `\Shadow`, "plain": `\Plain`, "unknownArg": `\Box`,
	})
}
