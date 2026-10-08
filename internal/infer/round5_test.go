package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
	"custos/internal/testbudget"
)

// extract(), one-argument parse_str() and `$$name =` may set any local:
// later reads of variables defined before them are unknown.
func TestDynamicWritesClobber(t *testing.T) {
	src := `<?php
function g($c, array $v, $q, $name) {
    $n = (int) $c;
    extract($v);
    t('extract', $n);
    $fresh = 1;
    t('after', $fresh);
}
function h($q) { $p = 1; parse_str($q); t('parse', $p); $o = 1; parse_str($q, $out); t('parse2', $o); }
function k($name) { $m = 1; $$name = 'x'; t('varvar', $m); }
function l(array $rows) {
    $x = 1;
    foreach ($rows as $r) { t('loop', $x); extract($r); }
}
`
	checkAnywhere(t, src, map[string]string{
		"extract": "?unknown", "after": "int", "parse": "?unknown", "parse2": "int", "varvar": "?unknown", "loop": "?unknown",
	})
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				if lit.Raw == "'extract'" {
					if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value); !got.IsUnknown() {
						t.Errorf("T-rules extract: %s", got)
					}
				}
			}
		}
		return true
	})
}

// A variable that only plain assignments define, read on a path that
// assigns it nowhere, may be null.
func TestPossiblyUndefined(t *testing.T) {
	checkAnywhere(t, `<?php
function f($t, $c, array $xs) {
    if ($t) { $r = 'X'; }
    t('maybe', $r);
    if ($t) { $a = 1; } else { $a = 2; }
    t('ifElse', $a);
    switch ($c) { case 1: $s = 'a'; break; default: $s = 'b'; }
    t('switchDefault', $s);
    switch ($c) { case 1: $s2 = 'a'; break; case 2: $s2 = 'b'; break; }
    t('switchNoDefault', $s2);
    try { $u = 1; } catch (\Exception $e) { $u = 2; }
    t('tryCatch', $u);
    do { $d = 1; } while ($c);
    t('doWhile', $d);
    if (($w = strlen('x')) > 0) {}
    t('condAssign', $w);
    if ($t) { $z = 1; } else { return; }
    t('elseReturns', $z);
    foreach ($xs as $x) { $l = 1; }
    t('afterLoop', $l);
    $c2 = $c && ($v = 1);
    t('rhsAssign', $v);
    if (isset($r)) { t('isset', $r); }
    while (($line = fgets(STDIN)) !== false) { t('whileCond', $line); }
    for ($i = 0; $i < 3; $i++) { t('forInit', $i); }
    echo $e2 = 'x';
    t('echo', $e2);
    $tern = $t ? ($q = 1) : 2;
    t('ternary', $q);
    if ($t) { $k = 1; } elseif ($c) { $k = 2; } else { $k = 3; }
    t('chain', $k);
    if ($t) { $k2 = 1; } elseif ($c) { } else { $k2 = 3; }
    t('chainGap', $k2);
    try { $f1 = 1; } finally { $f2 = 2; }
    t('finally', $f2);
    t('tryBody', $f1);
    try { $g1 = 1; } catch (\Exception $e) { }
    t('catchGap', $g1);
    { $blk = 1; }
    t('block', $blk);
    while ($c) { $wb = 1; }
    t('whileBody', $wb);
    foreach ($xs as $fx) { }
    $fe = null;
    foreach (($fe2 = $xs) as $y) {}
    t('foreachExpr', $fe2);
    switch ($sw = $c) { default: }
    t('switchCond', $sw);
    $m = match ($mm = $c) { default => 1 };
    t('matchCond', $mm);
    if ($t) { $br = 1; } else { $br = 2; }
    switch ($c) { case 1: $sc = 1; case 2: $sc = 2; break; default: $sc = 3; }
    t('fallthrough', $sc);
    switch ($c) { case 1: $se = 1; break; default: }
    t('emptyDefault', $se);
    foreach ($xs as $x) { if ($x) { $cb = 1; continue; } $cb = 2; }
    t('continueList', $cb);
}
`, map[string]string{
		"maybe": "null|string", "ifElse": "int", "switchDefault": "string", "switchNoDefault": "null|string",
		"tryCatch": "int", "doWhile": "int", "condAssign": "int", "elseReturns": "int", "afterLoop": "int|null",
		"rhsAssign": "int|null", "isset": "string", "whileCond": "string", "forInit": "int", "echo": "string",
		"ternary": "int|null", "chain": "int", "chainGap": "int|null", "finally": "int", "tryBody": "int",
		"catchGap": "int|null", "block": "int", "whileBody": "int|null", "foreachExpr": "array", "switchCond": "?unknown",
		"matchCond": "?unknown", "fallthrough": "int", "emptyDefault": "int|null", "continueList": "int|null",
	})
}

// The SpecOnly T-rules typer drops assignments followed by an exit too.
func TestSpecOnlyExits(t *testing.T) {
	src := `<?php
function f(string $sql, array $c) {
    if ($c) { $sql = array_shift($c); return $sql; }
    $y = $sql;
    t('y', str_replace('a', 'b', $y));
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
	tr.SpecOnly = true
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != "string" {
					t.Errorf("SpecOnly: %s", got)
				}
			}
		}
		return true
	})
}

// A pseudo-type name imported or declared as a class is that class.
func TestPseudoTypeShadowedByClass(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App { class Number {} class Scalar {} }
namespace X {
use App\Number;
class numeric {}
/**
 * @param array|Number $v
 * @param array|number $w
 * @param scalar $s
 * @param numeric $n
 */
function f($v, $w, $s, $n) { t('imported', $v); t('lower', $w); t('pseudo', $s); t('declared', $n); }
}
`, map[string]string{"imported": `\App\Number|array`, "lower": `\App\Number|array`, "pseudo": "bool|float|int|string", "declared": `\X\numeric`})
}

// class_alias() names resolve to the original class.
func TestClassAlias(t *testing.T) {
	checkAnywhere(t, `<?php
namespace X;
class user { public function id(): int { return 1; } }
class_alias(user::class, \core_user::class);
class_alias('X\user', 'legacy_user');
class_alias('legacy_user', 'older_user');
class_alias('loop_a', 'loop_b');
class_alias('loop_b', 'loop_a');
class_alias($dyn, 'never');
class_alias(static::class, 'never2');
class_alias(1, 'never3');
class_alias(user::class);
class_alias(class: 'X\user', alias: 'named');
class_alias('', 'empty');
function f() {
    t('const', (new \core_user)->id());
    t('string', (new \legacy_user)->id());
    t('chain', (new \older_user)->id());
    t('cycle', (new \loop_a)->id());
}
`, map[string]string{"const": "int", "string": "int", "chain": "int", "cycle": "?unknown"})
	ix := index.New(nil)
	f := syntax.Parse("a.php", []byte("<?php class A {} class_alias('A', 'B');"), syntax.Options{Version: phpver.PHP84})
	ix.Add(index.Extract(f))
	if ix.Class("B", 0) == nil {
		t.Fatal("alias not found")
	}
	ix.Remove("a.php")
	if ix.Class("B", 0) != nil {
		t.Fatal("alias kept after Remove")
	}
}

// An elseif chain on instanceof narrows the join past it.
func TestElseIfInstanceofJoin(t *testing.T) {
	checkAnywhere(t, `<?php
class Lang {} class Code {} class Sub extends Code {}
interface Marker {}
/**
 * @param string|Code $code
 * @param string|Code $c2
 * @param string|Code $c3
 */
function f($code, $c2, $c3, $x) {
    if ($code instanceof Lang) { $z = 1; } elseif ($code instanceof Code) { $code = 'x'; }
    t('join', $code);
    if ($c2 instanceof Code) { $c2 = 'y'; } else { $w = 1; }
    t('withElse', $c2);
    if ($c3 instanceof Sub) { $q = 1; } elseif ($x) { $c3 = $x; }
    t('unknownWrite', $c3);
    if ($c3 instanceof Code) { foo(); } elseif ($x) { $c3 .= 'a'; }
    t('otherWrite', $c3);
    if ($c2 instanceof Marker) { t('iface', $c2); }
    if ($x instanceof Code) { return; } elseif ($x) { throw new \Exception(); }
    t('leaves', $x);
}
`, map[string]string{
		"join": "string", "withElse": "string", "unknownWrite": "?unknown", "otherWrite": "?unknown",
		"iface": `\Marker|string`, "leaves": "?unknown",
	})
}

// var_export()/print_r() return a string only with a true $return.
func TestPrintReturn(t *testing.T) {
	checkAnywhere(t, `<?php
function f($x, bool $b) {
    t('ve', var_export($x, true));
    t('ve0', var_export($x));
    t('veFalse', var_export($x, false));
    t('veVar', var_export($x, $b));
    t('pr', print_r($x, true));
    t('pr0', print_r($x));
    t('named', print_r($x, return: true));
}
`, map[string]string{"ve": "string", "ve0": "null", "veFalse": "null", "veVar": "null|string", "pr": "string", "pr0": "true", "named": "string"})
}

// assert(cond) narrows the code after it like a guard.
func TestAssertNarrows(t *testing.T) {
	checkAnywhere(t, `<?php
namespace N;
class Code {}
function assert($x) {}
function f(?object $b, ?object $c) {
    \assert($b instanceof Code);
    t('assert', $b);
    assert($c instanceof Code);
    t('shadowed', $c);
}
function g(?object $d) { assert($d !== null); t('fallback', $d); }
`, map[string]string{"assert": `\N\Code`, "shadowed": "null|object", "fallback": "null|object"})
	checkAnywhere(t, `<?php
function g(?object $d, $x) { assert(); assert(...[$d]); assert(value: $d); $x->assert($d); assert($d !== null); t('global', $d); }
`, map[string]string{"global": "object"})
}

// AncestorsComplete reports a hierarchy cut by MaxAncestors.
func TestAncestorsComplete(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\nclass Big implements ")
	for i := 0; i < index.MaxAncestors+10; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "I%d", i)
	}
	b.WriteString(" {}\nclass Small {}\n")
	for i := 0; i < index.MaxAncestors+10; i++ {
		fmt.Fprintf(&b, "interface I%d {}\n", i)
	}
	f := syntax.Parse("t.php", []byte(b.String()), syntax.Options{Version: phpver.PHP84})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	if ix.AncestorsComplete("Big", 0) || !ix.AncestorsComplete("Small", 0) || !ix.AncestorsComplete("Missing", 0) {
		t.Error("AncestorsComplete")
	}
}

// Many possibly-undefined reads and dynamic writes stay bounded.
func TestPossiblyUndefinedBounded(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\nfunction f($c, array $v) {\n if ($c) { $r = 1; }\n")
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&b, " if ($c === %d) { $z = 1; } $y = $r;\n", i)
	}
	b.WriteString("}\nfunction g($c, array $v) {\n $r = 1;\n")
	for i := 0; i < 20000; i++ {
		b.WriteString(" extract($v); $y = $r;\n")
	}
	b.WriteString("}\n")
	if d := typeAll(t, b.String()); d > testbudget.Of(2*time.Second) && !raceEnabled {
		t.Errorf("took %v", d)
	}
}

// Conditions evaluated before a read assign the variable on every path.
func TestPossiblyUndefinedConditions(t *testing.T) {
	checkAnywhere(t, `<?php
function f($t, $c, array $xs) {
    if ($t) {} elseif ($e = $c) { t('elseifCond', $e); }
    foreach (($fx = $xs) as $y) { t('foreachExprIn', $fx); }
    switch ($s = $c) { case 1: t('switchIn', $s); }
    $m = match ($mm = $c) { 1 => t('matchIn', $mm), default => 0 };
    $b = ($bb = $c) && t('rhs', $bb);
    $q = ($qq = $c) ? t('ternaryIn', $qq) : 0;
}
`, map[string]string{"elseifCond": "?unknown", "foreachExprIn": "array", "switchIn": "?unknown", "matchIn": "?unknown", "rhs": "?unknown", "ternaryIn": "?unknown"})
	checkAnywhere(t, `<?php
function g(int $c, array $xs) {
    if ($c) {} elseif ($e = $c) { t('elseifCond', $e); }
    switch ($s = $c) { case 1: t('switchIn', $s); }
    $m = match ($mm = $c) { 1 => t('matchIn', $mm), default => 0 };
    $b = ($bb = $c) && t('rhs', $bb);
    $q = ($qq = $c) ? t('ternaryIn', $qq) : 0;
    for ($fi = 0; $fi < 3; $fi++) {}
    t('afterFor', $fi);
    if ($c) { $kc = 1; }
    switch ($c) { case 1: $kc2 = 1; break; case t('caseCond', $kc): break; }
}
/**
 * @param object|\ArrayIterator<int, string> $o
 * @param string|\ArrayIterator $x
 */
function h($o, $x) {
    if ($o instanceof \ArrayIterator) { t('args', $o); }
    if ($x instanceof \Countable) { return; } elseif ($x === '') { throw new \Exception(); }
    t('leaves', $x);
}
`, map[string]string{"elseifCond": "int", "switchIn": "int", "matchIn": "int", "rhs": "int", "ternaryIn": "int",
		"args": `\ArrayIterator<int, string>`, "leaves": "string", "afterFor": "int", "caseCond": "int|null"})
	src := `<?php
function k($x) {
    t('later', $u);
    $u = 1;
    t('pr', print_r($x, true));
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	for _, spec := range []bool{false, true} {
		tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
		tr.SpecOnly = spec
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
					lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
					got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
					if want := map[string]string{"'later'": "?unknown", "'pr'": "string"}[lit.Raw]; got != want {
						t.Errorf("spec=%v %s: %s want %s", spec, lit.Raw, got, want)
					}
				}
			}
			return true
		})
	}
}
