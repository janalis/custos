package infer_test

import "testing"

// A closure passed where a function or method documents a callable
// signature receives the parameter types that signature lists, with the
// receiver's and the other arguments' templates bound.
func TestDocumentedCallbackParams(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App;
class Foo {}
class Bar {}
/**
 * @template TKey of array-key
 * @template T
 */
interface Collection {
    /**
     * @param \Closure(T, TKey): bool $p
     * @return static
     */
    public function filter(\Closure $p);
    /**
     * @template U
     * @param \Closure(T): U $fn
     * @return Collection<TKey, U>
     */
    public function map(\Closure $fn);
    /** @param callable(Foo $x, int ...$rest): void $cb */
    public static function each(callable $cb, int $n = 0);
}
/**
 * @template T
 * @param list<T> $items
 * @param callable(T, int): void $fn
 */
function walk(array $items, callable $fn): void {}
/** @param callable(Bar): void $fn */
function plain(callable $fn): void {}
class Other {
    /** @param callable(Bar): void $fn */
    public function run(callable $fn) {}
}
class Diff {
    /** @param callable(Foo): void $fn */
    public function run(callable $fn) {}
}
/**
 * @param Collection<int, Foo> $foos
 * @param Other|Diff $either
 * @param Other|Other $same
 */
function f(Collection $foos, $either, Other $same) {
    $foos->filter(function ($foo, $key) { t('elem', $foo); t('key', $key); return true; });
    $foos->map(fn($foo) => t('mapped', $foo));
    $foos->filter(function (Bar $typed) { t('typed', $typed); return true; });
    Collection::each(function ($x, $n) { t('first', $x); t('variadic', $n); });
    walk([new Bar()], function ($item, $i) { t('tpl', $item); t('index', $i); });
    plain(fn($b) => t('plain', $b));
    $either->run(fn($v) => t('disagree', $v));
    $same->run(fn($v) => t('agree', $v));
    plain(fn: fn($named) => t('named', $named));
    unknownfn(fn($u) => t('unknown', $u));
}
`, map[string]string{
		"elem": `\App\Foo`, "key": "int", "mapped": `\App\Foo`, "typed": `\App\Bar`, "first": `\App\Foo`,
		"variadic": "?unknown", "tpl": `\App\Bar`, "index": "int", "plain": `\App\Bar`, "disagree": "?unknown",
		"agree": `\App\Bar`, "named": `\App\Bar`, "unknown": "?unknown",
	})
}

// Callees resolved only in part, intersections and native typing.
func TestDocumentedCallbackParamEdges(t *testing.T) {
	src := `<?php
namespace App;
class Foo {}
interface Tagged {}
class Runner {
    /** @param callable(Foo): void $fn */
    public function run(callable $fn) {}
    /** @param callable(Foo): void $fn */
    public static function call(callable $fn) {}
}
/**
 * @param Runner&Tagged $both
 * @param Tagged&Runner $other
 */
function f(Runner $r, $m, $both, $other) {
    $r->$m(fn($dyn) => t('dynMethod', $dyn));
    Runner::$m(fn($dyns) => t('dynStatic', $dyns));
    Runner::missing(fn($miss) => t('missing', $miss));
    Runner::call(fn($st) => t('static', $st));
    $r->run(fn() => 1, fn($extra) => t('extra', $extra));
    $both->run(fn($in) => t('intersection', $in));
    $other->run(fn($in2) => t('intersection2', $in2));
}
`
	checkAnywhere(t, src, map[string]string{
		"dynMethod": "?unknown", "dynStatic": "?unknown", "missing": "?unknown", "static": `\App\Foo`,
		"extra": "?unknown", "intersection": `\App\Foo`, "intersection2": `\App\Foo`,
	})
	checkNative(t, src, map[string]string{"static": "?unknown"})
}
