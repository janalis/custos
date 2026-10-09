package infer_test

import "testing"

func TestClosureReturnTypes(t *testing.T) {
	check(t, `<?php
namespace App;
class Foo { public function name(): string { return ''; } public function __invoke(int $x): float { return 1.0; } }
final class Box {
    /** @var \Closure(int): Foo */
    private \Closure $factory;
    /** @param callable(int): string $cb */
    public function run(callable $cb, ?\Closure $maybe, Foo $foo, array $xs) {
        $f = function () { return new Foo(); };
        $g = fn(int $x) => $x * 2;
        $h = function (): ?Foo { return null; };
        $v = function (): void {};
        $gen = function () { yield 1; };
        $s = static fn() => 'a';
        t('f', $f()); t('g', $g(1)); t('h', $h()); t('v', $v()); t('gen', $gen());
        t('iife', (fn() => 1.5)()); t('cb', $cb(1)); t('maybe', $maybe());
        t('prop', ($this->factory)(1)); t('invoke', $foo(1));
        t('cuf', call_user_func($f)); t('cufa', call_user_func_array($g, [1]));
        t('map', array_map($f, $xs)); t('mapstr', array_map('trim', $xs)); t('mapint', array_map('intval', $xs));
        t('mapcb', array_map($cb, $xs)); t('mapnull', array_map(null, $xs, $xs));
        t('mapunknown', array_map(fn($x) => $x, $xs));
        t('nested', (fn() => fn() => 1)()());
        t('closure', $s);
    }
}
`, map[string]string{
		"f": `\App\Foo`, "g": "int", "h": `\App\Foo|null`, "v": "null", "gen": `\Generator`,
		"iife": "float", "cb": "string", "maybe": "?unknown", "prop": `\App\Foo`, "invoke": "float",
		"cuf": `\App\Foo`, "cufa": "int", "map": `\App\Foo[]`, "mapstr": "string[]", "mapint": "int[]",
		"mapcb": "string[]", "mapnull": "array[]", "mapunknown": "array", "nested": "int", "closure": `\Closure`,
	})
}

func TestCallableTemplateBinding(t *testing.T) {
	check(t, `<?php
namespace App;
class Foo {}
/**
 * @template T
 * @param callable(): T $f
 * @return T
 */
function lazy(callable $f) { return $f(); }
/**
 * @template U
 * @param \Closure(int): U $f
 * @return list<U>
 */
function mapInts(\Closure $f): array { return []; }
function run(array $xs) {
    t('lazy', lazy(fn() => new Foo()));
    t('lazystr', lazy('phpversion'));
    t('map', mapInts(fn(int $i) => $i * 1.5));
    t('reduce', array_reduce($xs, fn($c, $x) => 1, 0));
    t('unknown', lazy(fn() => $GLOBALS['x']));
}
`, map[string]string{
		"lazy": `\App\Foo`, "lazystr": "false|string", "map": "float[]", "reduce": "int", "unknown": "mixed",
	})
}

// Sorting callbacks change neither the element type nor the variable's
// type (only the shape: keys are renumbered).
func TestSortKeepsElementType(t *testing.T) {
	check(t, `<?php
namespace App;
class Foo {}
function run() {
    $xs = [new Foo(), new Foo()];
    usort($xs, fn(Foo $a, Foo $b) => 0);
    t('usort', $xs);
    $m = ['a' => 1, 'b' => 2];
    uasort($m, fn($a, $b) => $a <=> $b);
    t('uasort', $m);
    $f = array_filter(['a' => new Foo(), 'b' => null]);
    t('filterShape', $f);
}
`, map[string]string{"usort": `\App\Foo[]`, "uasort": "int[]", "filterShape": `\App\Foo[]`})
}
