package infer_test

import "testing"

func TestElementWriteIntoNull(t *testing.T) {
	checkAnywhere(t, `<?php
function a(?array $n) { $n['k'] = 1; t('param', $n); }
function b() { $n = null; $n[] = 'x'; t('fromNull', $n); }
function c(bool $c) { $n = null; if ($c) { $n['k'] = 1; } t('maybe', $n); }
function d(?array $n, bool $c) { if ($c) { $n['k'] = 1; t('inBranch', $n); } t('after', $n); }
function e(?array $n) { $n['k'] = 1; $n = null; t('reassigned', $n); }
function f(?array $n) { $n['k'] = 1; sort($n); t('byRef', $n); }
function g(?string $s) { $s[0] = 'a'; t('string', $s); }
function h(?array $n) { $n['a']['b'] = 1; t('nested', $n); }
function i($n) { if ($n === null) { $n = null; } $n['k'] = 1; t('unknownBase', $n); }
function j(?array $n) { foreach ([1] as $x) { $n[] = $x; } t('loopWrite', $n); }
`, map[string]string{
		"param": "array", "fromNull": "string[]", "maybe": "int[]|null", "inBranch": "array",
		"after": "array|null", "reassigned": "null", "byRef": "array|null", "string": "null|string",
		"nested": "array", "loopWrite": "array|null",
	})
}
