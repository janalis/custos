package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/syntax"
)

func TestCovCallables(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App;
class Foo {}
function plain() { return 1; }
/** @template T @param T $x @return T */
function ident($x) { return $x; }
function make() {}
class Inv {
    /** @return ($x is string ? int : float) */
    public function __invoke($x) {}
}
class NoInvoke {}
function run(array $xs, Inv $inv, NoInvoke $ni) {
    t('noInvoke', $ni());
    $doc = /** @return Foo */ function () { return make(); };
    t('doc', $doc());
    $orVoid = /** @return int|void */ function () { return 1; };
    t('orVoid', $orVoid());
    t('plainName', array_map('App\plain', $xs));
    t('staticName', array_map('Foo::bar', $xs));
    t('missingName', array_map('nope', $xs));
    t('tplName', array_map('App\ident', $xs));
    t('interp', array_map("p$x", $xs));
    t('invokeCond', $inv('a'));
    t('array', call_user_func([$inv, '__invoke'], 1));
    $objs = [new Foo(), null];
    t('filterNull', array_filter($objs));
    t('filterCb', array_filter($objs, fn($o) => $o !== null));
    $nulls = [null, null];
    t('allNull', array_filter($nulls));
    t('filterPlain', array_filter($xs));
    $mixed = [1, $GLOBALS];
    t('filterMixed', array_filter([$xs, $GLOBALS['a']]));
}
`, map[string]string{
		"doc": `\App\Foo`, "orVoid": "int|null", "plainName": "int[]", "staticName": "array",
		"missingName": "array", "tplName": "array", "interp": "array",
		"noInvoke": "?unknown", "invokeCond": "float|int", "array": "mixed",
		"filterNull": `\App\Foo[]`, "filterCb": `\App\Foo[]|null[]`, "allNull": "null[]", "filterPlain": "array",
		"filterMixed": "array",
	})
}

func TestCovConditionalReturns(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App;
enum E { case A; }
final class K {
    const A = 1;
    const REF = self::A;
    const HEX = 0x1;
    const S = 'x';
}
/** @return ($zz is int ? int : float) */
function unknownParam($x) {}
/** @return ($x is -1 ? int : float) */
function neg($x) {}
/** @return ($x is K::NOPE ? int : float) */
function missingConst($x) {}
/** @return ($x is K::REF ? int : float) */
function refConst($x) {}
/** @return ($x is K::HEX ? int : float) */
function hexConst($x) {}
/** @return ($x is 0x1 ? int : float) */
function hexLit($x) {}
/** @return ($x is mixed ? int : float) */
function anything($x) {}
/** @return ($x is string ? int : float) */
function str($x) {}
/** @return ($x is 'x' ? int : float) */
function lit($x) {}
/** @return ($x is E::A ? int : float) */
function enumCase($x) {}
/** @return ($x is int ? int : float) */
function variadic(...$x) {}
function run($r, callable $c, iterable $it, object $o, array $a, E $e, mixed $m) {
    t('float', str(1.5)); t('list', str([1])); t('mixedArg', str($m));
    $res = fopen('php://memory', 'r');
    t('unknownParam', unknownParam(1));
    t('neg', neg(-1)); t('neg2', neg(-2));
    t('missingConst', missingConst(1));
    t('refConst', refConst(1));
    t('hexConst', hexConst(1));
    t('hexLit', hexLit(1));
    t('anything', anything(1));
    t('resource', str($res)); t('callable', str($c)); t('iterable', str($it));
    t('object', str($o)); t('array', str($a)); t('never', str(exit()));
    t('litConst', lit(K::S)); t('litInt', lit(1)); t('litArray', lit($a));
    t('enumCase', enumCase(E::A)); t('enumOther', enumCase($e));
    t('variadic', variadic(1));
    t('fcc', str(...));
}
`, map[string]string{
		"unknownParam": "float|int", "neg": "int", "neg2": "float", "missingConst": "float|int",
		"refConst": "float|int", "hexConst": "float|int", "hexLit": "float|int", "anything": "int",
		"resource": "float", "callable": "float|int", "iterable": "float", "object": "float", "array": "float",
		"never": "float",
		"float": "float", "list": "float", "mixedArg": "float|int", "litConst": "int", "litInt": "float", "litArray": "float",
		"enumCase": "int", "enumOther": "float|int", "variadic": "float|int", "fcc": `\Closure(): (float|int)`,
	})
	// A conditional that does not parse (a stale or hand-written index) is
	// ignored.
	fs := &index.FileSymbols{Path: "f.php", Functions: []*index.Function{{FQN: "f", CondReturn: "garbage", DocReturn: "int"}}}
	env, f := envFor(t, `<?php t('g', f(1));`, fs)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "f" {
				if got := env.TypeOf(c).String(); got != "int" {
					t.Errorf("garbage conditional: %s", got)
				}
			}
		}
		return true
	})
}

func TestCovOutParamsAndWrites(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App;
class Foo { public function go(): void {} }
function run(string $s, array $arr, bool $c) {
    similar_text('a', 'b', $pct); t('pct', $pct);
    is_callable('f', false, $name); t('name', $name);
    exec('ls', $arr); t('plain', $arr);
    $t = $c ? preg_match('/a/', $s, $tm) : 0; t('ternary', $tm);
    $d = $c || preg_match('/a/', $s, $om); t('or', $om);
    $tm = 1;
    $q = $c ? 1 : preg_match('/a/', $s, $tm); t('ternaryElse', $tm);
    while (preg_match('/a/', $s, $wm)) { t('while', $wm); break; }
    switch (preg_match('/a/', $s, $sm)) { default: t('switch', $sm); }
    echo preg_match('/a/', $s, $em); t('echo', $em);
    if ($c) preg_match('/a/', $s, $bm); t('unbraced', $bm);
    $a = [1, 2]; next($a); t('pointer', end($a));
    $n = null; $n['k'] = 1; $g = fn() => $n; t('arrow', $g());
    $cmp = preg_match('/a/', $s, $cm) > 0; t('compare', $cm);
    $af = fn() => [preg_match('/a/', $s, $am), t('inArrow', $am)];
    $tv = preg_match('/a/', $s, $tc) ? 1 : 0; t('ternaryCond', $tc);
    $z = null; $y = $z['k'] = 1; t('chained', $z);
    mixedOut($mo); t('mixedOut', $mo);
}
/** @param-out mixed $x */
function mixedOut(&$x) {}
function ret(string $s) { return preg_match('/a/', $s, $rm); }
class P {
    private ?Foo $p = null;
    public function dead() {
        $this->p = new Foo();
        if ($this->p === null) { t('dead', $this->p); }
    }
}
preg_match('/a/', 'x', $top); t('top', $top);
`, map[string]string{
		"pct": "float", "name": "string", "plain": "array", "ternary": "string[]", "or": "string[]",
		"ternaryElse": "int|string[]", "while": "string[]", "switch": "string[]", "echo": "string[]",
		"unbraced": "string[]", "pointer": "int", "arrow": "int[]", "dead": "null", "compare": "string[]", "inArrow": "string[]", "ternaryCond": "string[]", "chained": "int[]|null", "mixedOut": "?unknown", "top": "string[]",
	})
}
