package infer_test

import "testing"

// An untyped callback parameter receives the class template argument the
// receiver binds, through the callable signature of the method it is
// passed to.
func TestCallbackParamFromReceiverTemplate(t *testing.T) {
	checkShape(t, `<?php
class Rec {}
/**
 * @template TKey of array-key
 * @template TValue
 */
class Coll {
	/**
	 * @param callable(TValue, TKey): bool $cb
	 * @return static
	 */
	public function reject($cb) { return $this; }
	/**
	 * @template TMap
	 * @param callable(TValue, TKey): TMap $cb
	 * @return static<TKey, TMap>
	 */
	public function map(callable $cb) { return $this; }
	/**
	 * @param (callable(TValue, TKey): bool)|bool|TValue $cb
	 * @return static
	 */
	public function union($cb = true) { return $this; }
}
/**
 * @template TKey of array-key
 * @template TValue
 */
trait Enumerates {
	/**
	 * @param callable(TValue, TKey): bool $cb
	 * @return static
	 */
	public function traitMethod($cb) { return $this; }
}
/**
 * @template TKey of array-key
 * @template TValue
 */
class Coll2 { /** @use Enumerates<TKey, TValue> */ use Enumerates; }
/** @param Coll<int, Rec> $c */
function f($c) {
	$c->reject(function ($r, $k) { t('r', $r); t('k', $k); return true; });
	$c->map(fn ($m) => t('m', $m));
	$c->union(fn ($u) => t('u', $u));
}
/** @param Coll2<int, Rec> $c */
function g($c) {
	$c->traitMethod(fn ($tr) => t('tr', $tr));
}
`, map[string]string{"r": `\Rec`, "k": "int", "m": `\Rec`, "u": `\Rec`, "tr": `\Rec`})
}
