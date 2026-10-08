package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// str_replace()/preg_replace() & co. follow the subject's members: a
// string for scalars and objects (converted), an array for arrays, null
// for the preg_ functions.
func TestReplaceSubjects(t *testing.T) {
	checkAnywhere(t, `<?php
class Name { public function __toString(): string { return ''; } }
function f(string $s, int $i, array $a, string|array $sa, Name $n, mixed $m) {
    t('preg', preg_replace('/x/', 'y', $s));
    t('pregInt', preg_replace('/x/', 'y', $i));
    t('pregArr', preg_replace('/x/', 'y', $a));
    t('pregBoth', preg_replace('/x/', 'y', $sa));
    t('pregObj', preg_replace('/x/', 'y', $n));
    t('pregMixed', preg_replace('/x/', 'y', $m));
    t('cb', preg_replace_callback('/x/', function ($m) { return 'y'; }, $s));
    t('cbArray', preg_replace_callback_array(['/x/' => function ($m) { return 'y'; }], $a));
    t('filter', preg_filter('/x/', 'y', $s));
    t('str', str_replace('a', 'b', $s));
    t('strArr', str_replace('a', 'b', $a));
    t('substr', substr_replace($s, 'b', 0, 1));
}
`, map[string]string{
		"preg": "null|string", "pregInt": "null|string", "pregArr": "array|null", "pregBoth": "array|null|string",
		"pregObj": "null|string", "pregMixed": "null|string|string[]", "cb": "null|string", "cbArray": "array|null",
		"filter": "null|string", "str": "string", "strArr": "array", "substr": "string",
	})
}

// max()/min() return one of their arguments (or an element of the single
// array argument, false too before PHP 8.0 unless known non-empty).
func TestMaxMin(t *testing.T) {
	src := `<?php
/**
 * @param int[] $xs
 * @param non-empty-array<int> $ne
 */
function f(int $i, int $j, float $f, array $xs, array $ne, $u, array $plain) {
    t('ints', max($i, $j));
    t('mixedNum', min($i, $f));
    t('list', max($xs));
    t('nonEmpty', max($ne));
    t('unknown', max($i, $u));
    t('plain', max($plain));
    t('none', max());
    t('spread', max(...$xs));
    t('abs', abs($i));
    t('round', round($i));
    t('floor', floor($f));
    t('ceil', ceil($i));
    t('intdiv', intdiv($i, $j));
}
`
	checkVer(t, phpver.PHP84, false, src, map[string]string{
		"ints": "int", "mixedNum": "float|int", "list": "int", "nonEmpty": "int", "unknown": "mixed",
		"plain": "mixed", "none": "mixed", "spread": "mixed", "abs": "int", "round": "float", "floor": "float",
		"ceil": "float", "intdiv": "int",
	})
	checkVer(t, phpver.PHP74, false, src, map[string]string{"list": "false|int", "nonEmpty": "int"})
}

// PHPDoc intersections stay intersections: a member of either side is
// found, and the type prints with `&`.
func TestIntersectionMembers(t *testing.T) {
	checkAnywhere(t, `<?php
interface A { public function a(): int; }
interface B { public function b(): string; }
class P { public ?int $p = null; }
class C { public function c(): bool { return true; } }
/** @template T */
interface Sel { /** @return T */ public function sel(); }
class Foo {}
class R {
    /** @return A&B */
    public function m() {}
    public function n(): (A&B)|null { return null; }
    /** @return P&A */
    public function o() {}
    /** @return \ArrayIterator<int, Foo>&Sel<Foo> */
    public function matching() {}
    /** @return (A&B)|C */
    public function alt() {}
}
function dnf((A&B)|(C&P) $x, (A&B)|null $y) { t('dnf', $x); t('dnfNull', $y); }
function f(R $r, A&B $ab) {
    t('ab', $r->m());
    t('a', $r->m()->a());
    t('b', $r->m()->b());
    t('missing', $r->m()->zz());
    t('nullable', $r->n());
    t('nullsafe', $r->n()?->b());
    t('prop', $r->o()->p);
    t('param', $ab->b());
    t('sel', $r->matching()->sel());
    foreach ($r->matching() as $v) { t('iter', $v); }
    t('alt', $r->alt());
    if ($ab instanceof A) { t('narrowed', $ab); }
    t('union', rand() ? $ab : new C());
}
`, map[string]string{
		"ab": `\A&\B`, "a": "int", "b": "string", "missing": "?unknown", "nullable": `(\A&\B)|null`,
		"nullsafe": "null|string", "prop": "int|null", "param": "string", "sel": `\Foo`, "iter": `\Foo`,
		"alt": `\A|\B|\C`, "narrowed": `\A`, "union": `\A|\B|\C`, "dnf": `\A|\B|\C|\P`, "dnfNull": `(\A&\B)|null`,
	})
}

// A guard on `$this->prop === false` (also null/true) narrows the property
// after an early return (Matomo Cookie::set()).
func TestPropertyGuardAfterReturn(t *testing.T) {
	checkAnywhere(t, `<?php
class Cookie {
    /** @var bool|string */
    protected $keyStore = false;
    protected $value = [];
    /** @param bool|string $keyStore */
    public function __construct($keyStore = false) { $this->keyStore = $keyStore; }
    public function set($name, $value)
    {
        if (is_null($value)) {
            if ($this->keyStore === false) {
                unset($this->value[$name]);
                return;
            }
            unset($this->value[t('inner', $this->keyStore)][$name]);
            return;
        }
        if ($this->keyStore === false) {
            $this->value[$name] = $value;
            return;
        }
        $this->value[t('after', $this->keyStore)][$name] = $value;
    }
}
final class Typed {
    private ?string $s = null;
    private bool|int $b = 0;
    public function f() {
        if ($this->s === null) { return; }
        t('notNull', $this->s);
        if ($this->b === true) { return; }
        t('notTrue', $this->b);
    }
}
`, map[string]string{"inner": "string|true", "after": "string|true", "notNull": "string", "notTrue": "false|int"})
}

// Writing one literal key does not change the others: `$row[1] =
// explode(…)` leaves `$row[0]` a string (phpMyAdmin Privileges).
func TestElementWriteOtherKey(t *testing.T) {
	checkAnywhere(t, `<?php
interface Res {
    /** @return list<string|null> */
    public function fetchRow(): array;
    /** @return array<string, string|null> */
    public function fetchAssoc(): array;
}
function f(Res $res, int $k) {
    while ($row = $res->fetchRow()) {
        $row[1] = explode(',', $row[1]);
        t('zero', $row[0]);
        t('one', $row[1]);
    }
    while ($a = $res->fetchAssoc()) {
        $a['privs'] = [1];
        t('user', $a['User']);
        t('privs', $a['privs']);
    }
    $l = $res->fetchRow();
    $l[] = 5;
    t('appendInt', $l[0]);
    t('appendStr', $l['x']);
    $c = $res->fetchRow();
    $c[$k] = 1.5;
    t('computed', $c[0]);
    $n = $res->fetchRow();
    $n[0][1] = 2;
    t('nested', $n[0]);
}
`, map[string]string{
		"zero": "null|string", "one": "null|string|string[]", "user": "null|string", "privs": "int[]|null|string{0: int}",
		"appendInt": "int|null|string", "appendStr": "null|string", "computed": "float|null|string", "nested": "null|string",
	})
}

// Builtins that returned false (or null) on invalid arguments before PHP
// 8.0 while the stubs give only the 8.0 type.
func TestPre80FailureReturns(t *testing.T) {
	src := `<?php
function f(string $s, array $a) {
    t('hash', hash('md5', $s));
    t('hmac', hash_hmac('md5', $s, 'k'));
    t('chunk', array_chunk($a, 2));
    t('words', str_word_count($s));
    t('count', substr_count($s, 'a'));
    t('substr', substr($s, 1));
}
`
	checkVer(t, phpver.PHP72, false, src, map[string]string{
		"hash": "false|string", "hmac": "false|string", "chunk": "array|null", "words": "false|int|string[]",
		"count": "false|int", "substr": "false|string",
	})
	checkVer(t, phpver.PHP84, false, src, map[string]string{
		"hash": "string", "hmac": "string", "chunk": "array", "words": "int|string[]", "count": "int", "substr": "string",
	})
}

// A native Env ignores user PHPDoc (params, properties, returns, inline
// @var, templates, assertions, @param-out, closures) but keeps builtin
// stub docs.
func TestNativeEnv(t *testing.T) {
	src := `<?php
class Box {
    /** @var int */
    private $doc = 1;
    private int $nat = 1;
    /** @return int */
    public function legacy() { return 1; }
    public function typed(): int { return 1; }
    /** @return int */
    public function wide(): mixed { return 1; }
    /** @phpstan-assert string $v */
    public static function isStr($v): void {}
}
/**
 * @param array<int, int> $ids
 * @param int $u
 */
function f(array $ids, $u, Box $b, string $s) {
    foreach ($ids as $id) { t('elem', $id); }
    t('param', $u);
    t('docProp', $b->doc);
    t('natProp', $b->nat);
    t('legacy', $b->legacy());
    t('typed', $b->typed());
    t('wide', $b->wide());
    /** @var int $v */
    $v = g();
    t('inline', $v);
    t('builtin', explode(',', $s));
    t('str', strtoupper($s));
    $cb = /** @return int */ function () { return g(); };
    t('closure', $cb());
    $m = $u;
    Box::isStr($m);
    t('assert', $m);
    t('named', array_map('trim', ['a']));
    t('strfn', array_map('legacyFn', ['a']));
}
/** @return int */
function legacyFn($x) { return 1; }
/** @param-out int $o */
function out(&$o): void { $o = 1; }
function g() { out($r); t('out', $r); return $r; }
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	native := env.Native()
	if native.Native() != native || !native.IsNative() || env.IsNative() || env.Native() != native {
		t.Fatal("Native() must be cached and idempotent")
	}
	got := map[string][2]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				x := c.Args.Args[1].(*syntax.Arg).Value
				got[lit.Raw[1:len(lit.Raw)-1]] = [2]string{env.TypeOf(x).String(), native.TypeOf(x).String()}
			}
		}
		return true
	})
	want := map[string][2]string{
		"elem":    {"int", "?unknown"},
		"param":   {"int", "?unknown"},
		"docProp": {"int", "?unknown"},
		"natProp": {"int", "int"},
		"legacy":  {"int", "?unknown"},
		"typed":   {"int", "int"},
		"wide":    {"int", "mixed"},
		"inline":  {"int", "?unknown"},
		"builtin": {"string[]", "string[]"},
		"str":     {"string", "string"},
		"closure": {"int", "?unknown"},
		"assert":  {"string", "?unknown"},
		"named":   {"string[]", "string[]"},
		"strfn":   {"int[]", "array"},
		"out":     {"int", "?unknown"},
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %v want %v", k, got[k], w)
		}
	}
	// The T-rules typer over a native Env drops user docs too.
	tr := infer.NewTRules(native)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				x := c.Args.Args[1].(*syntax.Arg).Value
				switch lit.Raw[1 : len(lit.Raw)-1] {
				case "param", "legacy":
					if r := tr.TypeOf(x); !r.IsUnknown() {
						t.Errorf("T-rules %s: got %s want unknown", lit.Raw, r)
					}
				case "typed", "builtin":
					if r := tr.TypeOf(x); r.IsUnknown() {
						t.Errorf("T-rules %s: unknown", lit.Raw)
					}
				}
			}
		}
		return true
	})
}

// Statements once covered only by FuzzRules/FuzzFromDoc seeds.
func TestInferEdgesUnitOnly(t *testing.T) {
	checkVer(t, phpver.PHP84, false, `<?php
namespace N {
    const LOCAL = 'x';
    function f(array $a, string $s, $o) {
        t('values', array_values($a));
        t('isset', isset($a['k']));
        t('clone', clone $o);
        t('yield', yield 1);
        t('intConst', PHP_INT_MAX);
        t('dynName', \N\K::{'m'}());
        t('bare', (new K)->bare);
        t('staticProp', (new K)->m);
        t('unknownClass', $o::m());
        ['a' => $e] = strtoupper($s);
        t('destructCall', $e);
        t('offset', $s[0]);
        t('fallbackConst', \N\GLOBAL_INT);
        t('globalConst', GLOBAL_INT);
        t('missingConst', NO_SUCH_CONST);
        t('anon', new class {});
        t('dynStatic', \N\K::$m());
        t('magic', (new K)->magic);
        t('dup', ['a' => 1, 'a' => 'x']);
        ['a' => $d] = $s;
        t('destructString', $d);
        t('dq', ["a\n" => 1]["a\n"]);
        if ($a !== []) {
            $r = rand() ? sort($a) : t('ternaryOther', $a);
        }
    }
    /**
     * @property int $magic
     * @property $bare
     */
    class K { public static $m; }
}
namespace {
    const GLOBAL_INT = 5;
}
`, map[string]string{
		"values": "array", "isset": "bool", "clone": "?unknown", "offset": "string", "fallbackConst": "?unknown",
		"globalConst": "int", "missingConst": "?unknown", "anon": "object", "dynStatic": "?unknown", "magic": "int",
		"yield": "?unknown", "intConst": "int", "dynName": "?unknown", "bare": "?unknown", "staticProp": "?unknown", "unknownClass": "?unknown", "destructCall": "?unknown",
		"dup": "array{a: string}", "destructString": "?unknown", "dq": "int", "ternaryOther": "non-empty array",
	})
}

// The T-rules typer: max()/min() of known arguments, replacement
// functions on unknown subjects, assertions on an intersection receiver.
func TestTRulesFollowUp(t *testing.T) {
	checkAnywhere(t, `<?php
interface Checker { /** @phpstan-assert-if-true string $v */ public function isStr($v): bool; }
interface Named {}
class Q { /** @return Checker&Named */ public function c() {} }
function f(Q $q, int|string $v) {
    if ($q->c()->isStr($v)) { t('assertInter', $v); }
}
`, map[string]string{"assertInter": "string"})
	src := `<?php
function f(int $i, int $j, $u, float $f) {
    t('max', max($i, $j));
    t('maxMixed', max($i, $u));
    t('maxSpread', max(...[$i]));
    t('maxSpread2', max($i, ...[$j]));
    t('maxOne', max([$i]));
    t('minNum', min($i, $f));
    t('pregUnknown', preg_replace('/x/', 'y', $u));
    t('strUnknown', str_replace('a', 'b', $u));
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	for _, spec := range []bool{false, true} {
		tr := infer.NewTRules(env)
		tr.SpecOnly = spec
		want := map[string]string{
			"max": "int", "maxMixed": "mixed", "maxSpread": "mixed", "maxSpread2": "mixed", "maxOne": "mixed", "minNum": "float|int",
			"pregUnknown": "array|null|string", "strUnknown": "array|string",
		}
		if !spec { // an unknown subject leaves the result unknown
			want["pregUnknown"], want["strUnknown"] = "?unknown", "?unknown"
		}
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
					lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
					label := lit.Raw[1 : len(lit.Raw)-1]
					if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != want[label] {
						t.Errorf("spec=%v %s: got %s want %s", spec, label, got, want[label])
					}
				}
			}
			return true
		})
	}
}
