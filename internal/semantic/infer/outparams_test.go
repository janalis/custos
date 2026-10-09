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
