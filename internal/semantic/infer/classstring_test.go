package infer_test

import "testing"

// A class-string<C> value names C or a class extending it: `new`, static
// calls and a successful instanceof through it give a C.
func TestClassStringReceivers(t *testing.T) {
	checkWith(t, nil, `<?php
class Field {
    public static function new(string $p): static { return new static(); }
    public function dto(): int { return 1; }
}
/** @template T */
class Box { /** @param T $v */ public function __construct($v) {} }
/**
 * @param class-string<Field> $cls
 * @param class-string<Box> $box
 * @param class-string $any
 * @param class-string<Field>|class-string<Box> $either
 */
function run(string $cls, string $box, string $any, string $either, Field $f, object|int $x) {
    t('new', new $cls());
    t('static', $cls::new('a'));
    t('chain', $cls::new('a')->dto());
    t('newGeneric', new $box(1));
    t('any', new $any());
    t('either', new $either());
    t('object', new $f());
    if ($x instanceof $cls) { t('instanceof', $x); } else { t('notInstanceof', $x); }
    t('name', new Field());
}
`, map[string]string{
		"new": `\Field`, "static": `\Field`, "chain": "int", "newGeneric": `\Box`,
		"any": "?unknown", "either": "?unknown", "object": `\Field`,
		"instanceof": `\Field`, "notInstanceof": "int|object", "name": `\Field`,
	})
}
