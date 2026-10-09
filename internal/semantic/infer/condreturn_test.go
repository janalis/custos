package infer_test

import "testing"

func TestConditionalReturnTypes(t *testing.T) {
	check(t, `<?php
namespace App;
class Foo {}
class Bar extends Foo {}
final class K { const A = 1; const B = 2; const C = 1; const S = 'x'; }
/** @return ($x is string ? int : float) */
function f($x) {}
/** @phpstan-return ($flag is true ? Foo : Bar) */
function g(bool $flag = false) {}
/** @return ($x is not null ? int : null) */
function nn(?int $x) {}
/** @return ($x is 'a' ? int : ($x is 'b' ? float : string)) */
function lit($x) {}
/** @return ($m is K::A ? int : ($m is K::B ? float : string)) */
function kk(int $m = \App\K::B) {}
/**
 * @template T
 * @param T $v
 * @return (T is array ? list<string> : string)
 */
function tpl($v) {}
/** @return ($x is non-empty-string ? int : float) */
function refined($x) {}
/** @return ($x is Foo ? int : float) */
function cls($x) {}
class Svc {
    /** @psalm-return ($n is 1 ? int : string) */
    public function m(int $n): int|string { return 1; }
    public function run(string $s, int $i, $u, bool $b, Bar $bar) {
        t('fs', f($s)); t('fi', f($i)); t('fu', f($u)); t('flit', f('x'));
        t('gt', g(true)); t('gf', g(false)); t('gd', g()); t('gb', g($b)); t('gnamed', g(flag: true));
        t('nn', nn(1)); t('nnull', nn(null));
        t('la', lit('a')); t('lb', lit('b')); t('lc', lit('c')); t('lu', lit($s)); t('li', lit($i));
        t('ka', kk(K::A)); t('kc', kk(K::C)); t('kb', kk()); t('k1', kk(1)); t('k3', kk(3)); t('ks', kk($i));
        t('tarr', tpl([1])); t('tstr', tpl($s));
        t('rs', refined($s)); t('ri', refined($i));
        t('cbar', cls($bar)); t('cs', cls($s));
        t('m1', $this->m(1)); t('m2', $this->m(2)); t('mi', $this->m($i));
        $args = [true]; t('unpack', g(...$args));
    }
}
`, map[string]string{
		"fs": "int", "fi": "float", "fu": "float|int", "flit": "int",
		"gt": `\App\Foo`, "gf": `\App\Bar`, "gd": `\App\Bar`, "gb": `\App\Bar|\App\Foo`, "gnamed": `\App\Foo`,
		"nn": "int", "nnull": "null",
		"la": "int", "lb": "float", "lc": "string", "lu": "float|int|string", "li": "string",
		"ka": "int", "kc": "int", "kb": "float", "k1": "int", "k3": "string", "ks": "float|int|string",
		"tarr": "string[]", "tstr": "string",
		"rs": "float|int", "ri": "float",
		"cbar": "int", "cs": "float",
		"m1": "int", "m2": "string", "mi": "int|string",
		"unpack": `\App\Bar|\App\Foo`,
	})
}
