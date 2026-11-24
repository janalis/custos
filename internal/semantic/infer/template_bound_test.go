package infer_test

import "testing"

// A template bound to objects takes no scalar from an argument:
// `new \ReflectionClass($key)` with an int key matched `@param
// T|class-string<T>` and typed the instance it creates as an int.
func TestObjectBoundTemplateIgnoresScalars(t *testing.T) {
	check(t, `<?php
/** @template T of object */
final class Box {
 /** @param T|int $value */
 public function __construct(public $value) {}
 /** @return T */
 public function get() {}
}
function f(int $i, string $s, \Countable $c) {
 t('fromInt', (new \ReflectionClass($i))->newInstanceWithoutConstructor());
 t('fromObject', (new \ReflectionClass($c))->newInstanceWithoutConstructor());
 t('boxInt', (new Box($i))->get());
 t('boxObject', (new Box($c))->get());
}
`, map[string]string{
		"fromInt": "object", "fromObject": `\Countable`, "boxInt": "mixed", "boxObject": `\Countable`,
	})
}
