package infer_test

import "testing"

func TestCrossFileDeclarationBuiltinShadowInference(t *testing.T) {
	checkWith(t, map[string]string{
		"declarations.php": `<?php
namespace App;
class Original { public function value(): string { return ''; } }
define('SHADOW_FLAG', 7);
class_alias(Original::class, 'ShadowAlias');
\define('GLOBAL_FLAG', 7);
\class_alias(Original::class, 'GlobalAlias');
function readFlag() { return \SHADOW_FLAG; }
function readAlias() { return (new \ShadowAlias())->value(); }
`,
		"shadows.php": `<?php
namespace App;
function define($name, $value) { return false; }
function class_alias($original, $alias) { return false; }
`,
	}, `<?php
t('shadowConstant', SHADOW_FLAG);
t('shadowAlias', (new ShadowAlias())->value());
t('shadowConstantReturn', \App\readFlag());
t('shadowAliasReturn', \App\readAlias());
t('globalConstant', GLOBAL_FLAG);
t('globalAlias', (new GlobalAlias())->value());
`, map[string]string{
		"shadowConstant": "?unknown", "shadowAlias": "?unknown",
		"shadowConstantReturn": "?unknown", "shadowAliasReturn": "?unknown",
		"globalConstant": "int", "globalAlias": "string",
	})
}
