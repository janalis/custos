package infer_test

import "testing"

func TestEffectiveAnnotationsCrossFile(t *testing.T) {
	checkWith(t, map[string]string{"library.php": `<?php
namespace Docs;
/**
 * @return int
 * @psalm-return bool
 * @phpstan-return string
 */
function selected() {}
class Box {
 /**
  * @param int[] $items
  * @phpstan-param string[] $items
  */
 public function __construct(public array $items) {}
 /** @psalm-return float */
 public function prefixed() {}
}
/**
 * @template T
 * @param T $value
 * @return T
 * @phpstan-return bool
 */
function overridden($value) {}
/**
 * @return ($value is int ? string : float)
 * @phpstan-return bool
 */
function overriddenConditional($value) {}
`}, `<?php
t('selected', \Docs\selected());
$box = new \Docs\Box([]);
t('promoted', $box->items[0]);
t('method', $box->prefixed());
t('noTemplateFallback', \Docs\overridden(1));
t('noConditionalFallback', \Docs\overriddenConditional(1));
`, map[string]string{"selected": "string", "promoted": "string", "method": "float", "noTemplateFallback": "bool", "noConditionalFallback": "bool"})
}

func TestEffectiveAnnotationsLocalScopes(t *testing.T) {
	check(t, `<?php
/**
 * @param int[] $items
 * @psalm-param bool[] $items
 * @phpstan-param string[] $items
 * @psalm-param float $other
 */
function inspect($items, $other) {
 t('param', $items[0]); t('independent', $other);
 /** @phpstan-var float */
 t('unnamedNotAssigned', $missing);
 /** @phpstan-var array{'first name': string} $row */
 $row = load();
 t('inline', $row['first name']);
 /**
  * @var int $value
  * @phpstan-var $value string
  * @psalm-var bool $value
  */
 $value = load(); t('inlinePriority', $value);
 $closure = /** @psalm-return bool */ function() {};
 t('closure', $closure());
 $arrow = /** @phpstan-return string */ fn() => load();
 t('arrow', $arrow());
}
/**
 * @return \Generator<int, string, int, bool>
 * @phpstan-return \Generator<int, string, float, bool>
 */
function items() { t('send', yield 'value'); return true; }
`, map[string]string{"param": "string", "independent": "float", "unnamedNotAssigned": "?unknown", "inline": "string", "inlinePriority": "string", "closure": "bool", "arrow": "string", "send": "float|null"})
}

func TestEffectiveAnnotationsHooksAndNative(t *testing.T) {
	src := `<?php
class Box {
 /** @phpstan-var string[] */
 public array $items { set { t('implicitHook', $value[0]); } }
 public array $explicit {
  /** @psalm-param float[] $incoming */
  set(array $incoming) { t('explicitHook', $incoming[0]); }
 }
}
/** @phpstan-param string[] $items */
function inspect(array $items) {
 t('param', $items[0]);
 /** @psalm-var string $value */
 $value = 1; t('inline', $value);
 $closure = /** @phpstan-return string */ function() { return 1; };
 t('closure', $closure());
}
/** @psalm-return \Generator<int, string, float, bool> */
function generator() { t('send', yield 'value'); }
`
	hookTypes(t, src, false, false, map[string]string{"implicitHook": "string", "explicitHook": "float", "param": "string", "inline": "string", "closure": "string", "send": "float|null"})
	hookTypes(t, src, false, true, map[string]string{"implicitHook": "string", "explicitHook": "float", "param": "string"})
	hookTypes(t, src, true, false, map[string]string{"implicitHook": "?unknown", "explicitHook": "?unknown", "param": "?unknown", "inline": "int", "closure": "int", "send": "?unknown"})
}

// A standalone `@var Class $x` over a variable holding a class-name
// string describes the class, not the value (Yii ActiveField::widget()).
func TestInlineVarClassNameIdiom(t *testing.T) {
	checkAnywhere(t, `<?php
class Widget {}
/** @param string $class */
function f($class, array $config, $u) {
    /** @var Widget $class */
    $config['model'] = 1;
    t('class', $class);
    $n = 1;
    /** @var string $n */
    t('override', $n);
    /** @var Widget $u */
    t('unknown', $u);
    /** @var Widget $fresh */
    t('fresh', $fresh);
}
`, map[string]string{"class": "string", "override": "string", "unknown": `\Widget`, "fresh": `\Widget`})
}

// A native `iterable` is refined by its doc type, as `array` is.
func TestIterableRefinedByDoc(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
/** @template T */
class Coll implements \IteratorAggregate { public function getIterator(): \Iterator {} }
class H {
    /** @var array<string, Foo> */
    private iterable $p = [];
    /**
     * @param Foo[] $b
     * @param iterable<string, Foo> $c
     */
    public function f(iterable $b, ?iterable $c) {
        foreach ($b as $x) { t('b', $x); }
        foreach ($c as $k => $x) { t('c', $x); t('ck', $k); }
        t('prop', $this->p);
        t('ret', $this->g());
    }
    /** @return Coll<Foo> */
    public function g(): iterable {}
}
`, map[string]string{"b": `\Foo`, "c": `\Foo`, "ck": "string", "prop": `\Foo[]`, "ret": `\Coll<\Foo>`})
}

// A pseudo-type name imported or declared as a class is that class.
func TestPseudoTypeShadowedByClass(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App { class Number {} class Scalar {} }
namespace X {
use App\Number;
class numeric {}
/**
 * @param array|Number $v
 * @param array|number $w
 * @param scalar $s
 * @param numeric $n
 */
function f($v, $w, $s, $n) { t('imported', $v); t('lower', $w); t('pseudo', $s); t('declared', $n); }
}
`, map[string]string{"imported": `\App\Number|array`, "lower": `\App\Number|array`, "pseudo": "bool|float|int|string", "declared": `\X\numeric`})
}

// Promoted properties take their doc type from the constructor's @param.
func TestPromotedPropertyDoc(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
class H {
    /** @param Foo[] $items */
    public function __construct(private array $items, private int $n = 0) {}
    public function f() { t('items', $this->items); t('n', $this->n); }
}
`, map[string]string{"items": `\Foo[]`, "n": "int"})
}
