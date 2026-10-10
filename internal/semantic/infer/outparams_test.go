package infer_test

import "testing"

func TestOutParameters(t *testing.T) {
	check(t, `<?php
namespace App;
class Foo {}
/**
 * @param-out Foo $out
 * @param-out int $n
 */
function fill(&$out, &$n, &$plain, $byval) {}
class Svc {
    /** @phpstan-param-out list<string> $lines */
    public function read(?array &$lines): void {}
    /** @psalm-param-out Foo $x */
    public static function make(&$x): void {}
    public function run(string $s) {
        $m = null;
        preg_match('/a/', $s, $m);
        t('m', $m);
        preg_match('/a/', $s, $mf, PREG_OFFSET_CAPTURE); t('mf', $mf);
        preg_match_all('/a/', $s, $all); t('all', $all);
        if (preg_match('/(a)/', $s, $c)) { t('c', $c); }
        str_replace('a', 'b', $s, $count); t('count', $count);
        parse_str($s, $q); t('q', $q);
        exec('ls', $out, $code); t('out', $out); t('code', $code);
        $ints = [1]; exec('ls', $ints); t('ints', $ints);
        fill($o, $n, $p, $b); t('o', $o); t('n', $n);
        $p = 'x'; fill($o2, $n2, $p, $b); t('p', $p);
        $this->read($lines); t('lines', $lines);
        self::make($x); t('x', $x);
        $y = 1; $s !== '' && preg_match('/a/', $s, $y); t('y', $y);
    }
}
`, map[string]string{
		"m": "string[]", "mf": "array", "all": "string[][]", "c": "string[]", "count": "int",
		"q": "array", "out": "string[]", "code": "int", "ints": "int[]|string[]",
		"o": `\App\Foo`, "n": "int", "p": "string", "lines": "string[]", "x": `\App\Foo`,
		"y": "int|string[]",
	})
}

func TestPregMatchOutputTypes(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $s, $flags) {
    preg_match('/(a)(b)?/', $s, $m);
    t('m', $m);
    preg_match('/(a)(b)?/', $s, $m2, PREG_UNMATCHED_AS_NULL);
    t('m2', $m2);
    preg_match_all('/(a)(b)?/', $s, $m3, PREG_UNMATCHED_AS_NULL);
    t('m3', $m3);
    preg_match('/(a)/', $s, $m4, 0);
    t('m4', $m4);
    preg_match('/(a)/', $s, $m5, PREG_OFFSET_CAPTURE);
    t('m5', $m5);
    preg_match('/(a)/', $s, $m6, $flags);
    t('m6', $m6);
    preg_match('/(a)/', $s, $m7, ...[0]);
    t('m7', $m7);
}
`, map[string]string{
		"m":  "string[]",
		"m2": "null[]|string[]",
		"m3": "null[][]|string[][]",
		"m4": "string[]",
		"m5": "array",
		"m6": "array",
		"m7": "array",
	})
}
