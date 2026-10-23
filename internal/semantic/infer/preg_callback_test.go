package infer_test

import "testing"

// preg_replace_callback() passes its callback the matches as preg_match()
// stores them, by its flags.
func TestPregReplaceCallbackMatches(t *testing.T) {
	check(t, `<?php
function run(string $s, $flags) {
    preg_replace_callback('/(a)/', function ($m) { t('plain', $m); return ''; }, $s);
    preg_replace_callback('/(a)/', static function ($m) { t('closure', $m); return ''; }, $s);
    preg_replace_callback('/(a)/', function ($m) { t('null', $m); return ''; }, $s, -1, $n, PREG_UNMATCHED_AS_NULL);
    preg_replace_callback('/(a)/', function ($m) { t('offset', $m); return ''; }, $s, flags: PREG_OFFSET_CAPTURE);
    preg_replace_callback('/(a)/', function ($m) { t('zero', $m); return ''; }, $s, -1, $n, 0);
    preg_replace_callback('/(a)/', function ($m) { t('unknown', $m); return ''; }, $s, -1, $n, $flags);
    preg_replace_callback('/(a)/', function ($m, $x) { t('second', $x); return ''; }, $s);
}
`, map[string]string{
		"plain":   "string[]",
		"closure": "string[]",
		"null":    "null[]|string[]",
		"offset":  "array[]",
		"zero":    "string[]",
		"unknown": "?unknown",
		"second":  "?unknown",
	})
}
