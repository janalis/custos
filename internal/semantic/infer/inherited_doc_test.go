package infer_test

import "testing"

// An override declaring a wider native type than the documented return of
// the method it overrides (`: array` with `{@inheritDoc}`) keeps that doc,
// bound to the class's template arguments.
func TestInheritedDocReturn(t *testing.T) {
	check(t, `<?php
class Meta {}
/** @template T of object */
interface Factory {
    /** @phpstan-return list<T> */
    public function all(): array;
    /** @return array<string, T>|null */
    public function byName(): ?array;
    /** @return iterable<T> */
    public function each(): iterable;
    /** @return T|null */
    public function first(): ?object;
    /** @return string[] */
    public function names(): array;
}
/**
 * @template T of object
 * @template-implements Factory<T>
 */
abstract class AbstractFactory implements Factory {
    /** {@inheritDoc} */
    public function all(): array { return []; }
    public function byName(): ?array { return null; }
    public function each(): iterable { return []; }
    public function first(): ?object { return null; }
    public function names(): array { return []; }
}
/** @extends AbstractFactory<Meta> */
final class MetaFactory extends AbstractFactory {
    /** @return array<int, mixed> */
    public function own(): array { return []; }
    public function plain(): array { return []; }
}
function run(MetaFactory $f) {
    t('all', $f->all());
    t('byName', $f->byName());
    t('each', $f->each());
    t('first', $f->first());
    t('own', $f->own());
    t('names', $f->names());
    foreach ($f->all() as $m) { t('element', $m); }
}
`, map[string]string{
		"all":     `\Meta[]`,
		"byName":  `\Meta[]|null`,
		"each":    "iterable",
		"first":   "null|object",
		"own":     "mixed[]",
		"names":   "string[]",
		"element": `\Meta`,
	})
}
