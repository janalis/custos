package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// checkVer is check at PHP version v (ShapeString results), with the
// project index built from the file only (no stubs when noStubs).
func checkVer(t *testing.T, v phpver.Version, noStubs bool, src string, want map[string]string) {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: v})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	var base *index.Index
	if !noStubs {
		base = stubs.Index()
	}
	ix := index.New(base)
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, v)
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok { // anywhere, e.g. in an arrow function
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" && len(c.Args.Args) == 2 {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).ShapeString()
			}
		}
		return true
	})
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %s want %s", k, got[k], w)
		}
	}
}

func TestCovExpressionTypes(t *testing.T) {
	check(t, `<?php
const F = 1.5; const B = true; const N = null; const S = 'x'; const H = 0x1E;
class A { public static $p = 1; public static function m(): int { return 1; } }
function f(int $i, float $fl, string $s, $u, array $a) {
    t('fcc', A::m(...));
    t('line', __LINE__);
    t('print', print 'x');
    t('incdec', $i++);
    t('match', match ($i) { 1 => 'a', default => 2 });
    t('dynprop', A::$$s);
    t('mpi', M_PI);
    t('constf', F); t('constb', B); t('constn', N); t('consts', S); t('consth', H);
    t('arrcast', (array) $u); t('boolcast', (bool) $u); t('objcast', (object) $u);
    t('tilde', ~$i); t('tildes', ~$s); t('tildeu', ~$u);
    t('at', @$i); t('neg', -$fl); t('negs', -$s);
    t('addf', $i + $fl); t('arrplus', $a + $a); t('divs', $s / 2);
    t('mod', $i % 2); t('and', $i & $u); t('ands', $s & $s); t('andu', $u | $u); t('xor', $s ^ $i);
    t('coaleq', $u ??= 1); t('modeq', $i %= 2); t('diveq', $fl /= 2); t('divequ', $u /= 2);
    t('andeq', $s &= 'x'); t('oreq', $i |= 4);
    t('newdyn', new $u);
}
`, map[string]string{
		"fcc": `\Closure`, "line": "int", "print": "int", "incdec": "int", "match": "int|string",
		"dynprop": "?unknown", "mpi": "float",
		"constf": "float", "constb": "bool", "constn": "null", "consts": "string", "consth": "int",
		"arrcast": "array", "boolcast": "bool", "objcast": "object",
		"tilde": "int", "tildes": "string", "tildeu": "?unknown",
		"at": "float|int", "neg": "float", "negs": "?unknown",
		"addf": "float", "arrplus": "array", "divs": "?unknown",
		"mod": "int", "and": "int", "ands": "string", "andu": "?unknown", "xor": "int",
		"coaleq": "?unknown", "modeq": "int", "diveq": "float|int", "divequ": "?unknown",
		"andeq": "string", "oreq": "int",
		"newdyn": "?unknown",
	})
}

func TestCovVersionSpecificOperators(t *testing.T) {
	checkVer(t, phpver.PHP85, false, `<?php
function f(string $s, $u) {
    t('pipe', $s |> strlen(...));
    t('void', (void) $u);
}
`, map[string]string{"pipe": "int", "void": "void"})
	checkVer(t, phpver.PHP74, false, `<?php
function f($u) { t('unset', (unset) $u); }
`, map[string]string{"unset": "null"})
}

func TestCovDeclarationsAndCalls(t *testing.T) {
	check(t, `<?php
namespace App;
enum E { case One; }
interface I {}
/**
 * @template T
 * @phpstan-type Id int
 */
class Box implements I {
    const K = 1;
    /** @return static[] */
    public function many(): array { return []; }
    /** @return static|null */
    public function maybe() { return null; }
    /** @return Box|string */
    public function mixedDoc(): Box { return $this; }
    /**
     * @template K
     * @phpstan-type Name string
     */
    public function run($u, string $s, array $a, int $i) {
        /** @var T $tv */
        $tv = $u;
        /** @var K $kv */
        $kv = $u;
        /** @var Id $id */
        $id = $u;
        /** @var Name $nm */
        $nm = $u;
        /** @var Box $bx */
        $bx = $u;
        /** @var $first int */
        $first = $u;
        t('tv', $tv); t('kv', $kv); t('id', $id); t('nm', $nm); t('bx', $bx); t('first', $first);
        t('many', $this->many()); t('maybe', $this->maybe()); t('mixeddoc', $this->mixedDoc());
        t('dynstatic', Box::$s()); t('dynconst', Box::{$s}); t('staticclass', static::class);
        t('unkconst', $u::K); t('case', E::One);
        t('prc', preg_replace_callback('/x/', fn($m) => 'y', $s));
        t('sr', substr_replace($s, 'x', 0)); t('sra', str_replace('a', 'b', $a));
        t('mbs', mb_convert_encoding($s, 'UTF-8')); t('mba', mb_convert_encoding($a, 'UTF-8'));
        t('slice', array_slice([1, 2], 1)); t('filter', array_filter($a));
        t('varvar', $$s);
    }
}
`, map[string]string{
		"tv": "mixed", "kv": "mixed", "id": "int", "nm": "string", "bx": `\App\Box`, "first": "int",
		"many": `\App\Box[]`, "maybe": `\App\Box|null`, "mixeddoc": `\App\Box`,
		"dynstatic": "?unknown", "dynconst": "?unknown", "staticclass": "string",
		"unkconst": "?unknown", "case": `\App\E`,
		"prc": "null|string", "sr": "string", "sra": "array",
		"mbs": "false|string", "mba": "array|false",
		"slice": "int[]", "filter": "array",
		"varvar": "?unknown",
	})
}

func TestCovVariables(t *testing.T) {
	checkVer(t, phpver.PHP84, false, `<?php
class MyEx extends Exception {}
abstract class Abs { abstract public function f(int $p); }
function g(array $arr, $u) {
    $x = 1;
    $af = fn() => t('arrow', $x);
    $c = function () use ($undefined) { t('use', $undefined); };
    try { g([], 1); } catch (MyEx | TypeError $e) { t('catch', $e); }
    [&$r] = $arr; t('byref', $r);
    [$u => $k] = [1]; t('dynkey', $k);
    if ($u) { $a = 1; $b = 1; } else { $a = 'x'; }
    t('branch', $a);
    $w = ['k' => 1];
    $w['k'] .= 'x';
    t('compound', $w['k']);
    $lp = [1];
    foreach ($u as $x) { t('loopwhole', $lp); $lp[] = $x; }
    $s = ['a' => 1];
    $sf = fn() => t('arrowshape', $s);
}
`, map[string]string{
		"arrow": "int", "use": "?unknown", "catch": `\MyEx|\TypeError`,
		"byref": "?unknown", "dynkey": "?unknown", "branch": "int|string",
		"compound": "int|string", "loopwhole": "non-empty int[]", "arrowshape": "int[]{a: int}",
	})
}

func TestCovGenericsAndBodies(t *testing.T) {
	checkVer(t, phpver.PHP84, false, `<?php
class Assert {
    /** @phpstan-assert string $v */
    public static function str($v): void {}
    /** @phpstan-assert int $v */
    public static function int($v): void {}
    /** @phpstan-assert-if-true string $v */
    public static function isStr($v): bool { return true; }
}
/**
 * @template K
 * @template V
 */
class Pair {
    /** @return K */
    public function first() { return null; }
    /** @return V */
    public function second() { return null; }
}
trait Helper { private function h() { return 1; } public function g() { t('traitself', self::h()); return 1; } }
/** @extends Pair<int, mixed> */
final class P extends Pair { use Helper; }
function rec() { return rec(); }
function d1() { return d2(); } function d2() { return d3(); } function d3() { return d4(); }
function d4() { return d5(); } function d5() { return d6(); } function d6() { return 1; }
function withClosure() { $f = function () { return 'x'; }; return 1; }
function gen() { yield 1; return 2; }
function bare($x) { if ($x) { return; } return 1; }
abstract class Abs { abstract protected function abs(); public function g() { t('abstract', $this->abs()); } }
/** @param iterable<Pair> $it */
function f($u, P $p, iterable $raw, $it, string $s) {
    Assert::str($u); t('assertunknown', $u);
    $n = null;
    if ($n !== null && Assert::isStr($n)) { t('assertnone', $n); }
    t('first', $p->first()); t('second', $p->second());
    t('bare', bare($u)); t('rec', rec()); t('deep', d1()); t('closure', withClosure()); t('gen', gen());
    foreach ($it as $v) { t('iterv', $v); }
    t('m', match ($u) {});
    [$c] = $s; t('strdestruct', $c);
}
`, map[string]string{
		"assertunknown": "?unknown", "assertnone": "?unknown", "first": "int", "second": "mixed", "traitself": "?unknown",
		"bare": "int|null", "abstract": "?unknown", "rec": "?unknown", "deep": "?unknown", "closure": "int", "gen": `\Generator<int, int, mixed, int>`,
		"iterv": `\Pair`, "m": "?unknown", "strdestruct": "?unknown",
	})
	// Without stubs no Traversable class is known: iterating an object is unknown.
	checkVer(t, phpver.PHP84, true, `<?php
class Foo {}
function f(Foo $o) { foreach ($o as $v) { t('v', $v); } }
`, map[string]string{"v": "?unknown"})
}

// envFor parses src and builds an Env whose project index holds fs (or
// the file's own symbols when fs is nil).
func envFor(t *testing.T, src string, fs *index.FileSymbols) (*infer.Env, *syntax.File) {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	if fs == nil {
		fs = index.Extract(f)
	}
	ix := index.New(stubs.Index())
	ix.Add(fs)
	return infer.NewEnv(f, names.New(f), ix, phpver.PHP84), f
}

func TestCovEnvEdges(t *testing.T) {
	env, f := envFor(t, `<?php abstract class A { abstract public function f(int $p); } echo EMPTYC;`, nil)
	if !env.TypeOf(nil).IsUnknown() {
		t.Error("TypeOf(nil)")
	}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if p, ok := n.(*syntax.Param); ok {
			if got := env.TypeOf(p.Var).String(); got != "int" {
				t.Errorf("abstract method parameter: %s", got)
			}
		}
		return true
	})
	// A constant whose initialiser text is empty (malformed source) is unknown.
	fs := &index.FileSymbols{Path: "c.php", Constants: []*index.Constant{{FQN: "EMPTYC", Value: ""}}}
	env2, f2 := envFor(t, `<?php echo EMPTYC;`, fs)
	syntax.InspectFile(f2, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ConstFetch); ok {
			if got := env2.TypeOf(c); !got.IsUnknown() {
				t.Errorf("empty constant: %s", got)
			}
		}
		return true
	})
	// The index describes an older version of this very file (an editor
	// buffer ahead of the index): the class is not found at its span and its
	// untyped property is unknown.
	old := index.Extract(syntax.Parse("t.php", []byte(`<?php final class C { private $p; public function set() { $this->p = 1; } } function one() { return 1; }`), syntax.Options{Version: phpver.PHP84}))
	env3, f3 := envFor(t, `<?php

final class C { private $p; public function set() { $this->p = 1; } public function get() { return $this->p; } }
function one() { return 1; }
function two() { return one(); }`, old)
	syntax.InspectFile(f3, func(n syntax.Node) bool {
		if r, ok := n.(*syntax.Return); ok {
			if _, lit := r.Expr.(*syntax.Literal); lit {
				return true
			}
			if got := env3.TypeOf(r.Expr); !got.IsUnknown() {
				t.Errorf("stale index property: %s", got)
			}
		}
		return true
	})
}
