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
