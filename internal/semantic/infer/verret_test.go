package infer_test

import (
	"testing"

	phpversion "custos/internal/php/version"
)

// Builtin return types follow the target PHP version (phpstorm-stubs
// version maps): substr() returns string|false before PHP 8.0.
func TestVersionedBuiltinReturns(t *testing.T) {
	src := `<?php
function f(string $s) { t('substr', substr($s, 1)); t('strpos', strpos($s, 'a')); t('date', (new DateTime())->format('Y')); }
`
	checkVer(t, phpversion.PHP74, false, src, map[string]string{"substr": "false|string", "strpos": "false|int", "date": "string"})
	checkVer(t, phpversion.PHP84, false, src, map[string]string{"substr": "string", "strpos": "false|int", "date": "string"})
}

func TestExplodeLiteralSeparator(t *testing.T) {
	src := `<?php
function f(string $s, string $sep) { t('lit', explode(',', $s)); t('var', explode($sep, $s)); t('empty', explode('', $s)); t('interp', explode("$sep", $s)); }
`
	checkVer(t, phpversion.PHP74, false, src, map[string]string{"lit": "string[]", "var": "false|string[]", "empty": "false|string[]", "interp": "false|string[]"})
}
