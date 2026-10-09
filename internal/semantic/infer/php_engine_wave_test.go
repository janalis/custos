package infer_test

import "testing"

func TestPHPDocBareClassStringUnionInference(t *testing.T) {
	check(t, `<?php
final class Product {}
/**
 * @template T
 * @param class-string<T> $name
 * @return T
 */
function create($name) { return new $name(); }
/** @param class-string<Product>|string $name */
function broad($name) { t('broad', create($name)); }
/** @param string|class-string<Product> $name */
function reversed($name) { t('reversed', create($name)); }
/** @param class-string<Product>|null $name */
function constrained($name) {
 if ($name !== null) { t('constrained', create($name)); }
}
`, map[string]string{"broad": "mixed", "reversed": "mixed", "constrained": `\Product`})
}

func TestRuntimeDeclarationBuiltinInference(t *testing.T) {
	check(t, `<?php
namespace App;
use function define as register;
use function class_alias as alias_type;
final class Product { public function count(): int { return 1; } }
register(value: 12, constant_name: 'CUSTOS_WAVE_COUNT');
alias_type(alias: 'App\ProductAlias', class: Product::class);
define('CUSTOS_WAVE_SHADOW', 1);
class_alias(Product::class, 'App\ShadowAlias');
function define($name, $value) {}
function class_alias($class, $alias) {}
function run() {
 t('constant', \CUSTOS_WAVE_COUNT);
 t('alias', (new ProductAlias())->count());
 t('shadowConstant', \CUSTOS_WAVE_SHADOW);
 t('shadowAlias', (new ShadowAlias())->count());
}
`, map[string]string{"constant": "int", "alias": "int", "shadowConstant": "?unknown", "shadowAlias": "?unknown"})
}
