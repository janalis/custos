package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// reachingProbe reads, n times, variables that each have about perVar
// conditional definitions (read after every one of them).
func reachingProbe(n, perVar int) string {
	var b strings.Builder
	b.WriteString("<?php\nfunction f($c) {\n")
	vars := max(1, n/perVar)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, " if ($c === %d) { $r%d = %d; } $y = $r%d;\n", i, i%vars, i, i%vars)
	}
	b.WriteString("}\n")
	return b.String()
}

// Reaching definitions stay linear in the reads for a fixed number of
// definitions per variable: 4x the reads costs about 4x (the kill filter
// was quadratic in the reaching definitions: 400 definitions read 20k
// times took 10 s).
func TestReachingScales(t *testing.T) {
	if raceEnabled {
		t.Skip("timing")
	}
	small, large := reachingProbe(2000, 200), reachingProbe(8000, 200)
	best := func(src string) float64 {
		d := typeAll(t, src)
		if d2 := typeAll(t, src); d2 < d {
			d = d2
		}
		return float64(d)
	}
	if r := best(large) / best(small); r > 8 {
		t.Errorf("4x the reads took %.1fx the time", r)
	}
}

// Constructors that only store their parameters are indexed as such.
func TestStoresParamsIndex(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php
class A { public function __construct($db, protected int $n) { $this->db = $db; } }
class B { public function __construct($db) { $this->db = strtolower($db); } }
class C { public function __construct($db) { $this->db = $db; foo(); } }
class D { public function __construct() {} public function m() { return random_bytes(8); } }
class E { public function __construct($x) { $this->a[] = $x; } }
class F { public function __construct($x) { $y->a = $x; } }
class G { public function __construct($x) { $this->a = $other; } }
class H { public function __construct($x) { $this->a = &$x; } }
abstract class I { abstract public function __construct($x); }
class J { public function __construct($x) { $this?->a = $x; } }
function iv() { return openssl_random_pseudo_bytes(16); }
function weak() { return 'x'; }
function viaMethod($o) { return $o->random_bytes(8); }
function viaStatic() { return R::mcrypt_create_iv(8); }
function qualified() { return \random_bytes(8); }
`), syntax.Options{Version: phpversion.PHP84})
	fs := index.Extract(f)
	want := map[string]bool{"A": true, "B": false, "C": false, "D": true, "E": false, "F": false, "G": false, "H": false, "I": false, "J": false}
	for _, c := range fs.Classes {
		if got := c.Methods["__construct"].StoresParams; got != want[c.FQN] {
			t.Errorf("%s: StoresParams %v", c.FQN, got)
		}
	}
	if s := fs.Classes[0].Methods["__construct"].Stores; len(s) != 2 || s[0] != "n" || s[1] != "db" {
		t.Errorf("A stores %v", s)
	}
	if !fs.Classes[3].Methods["m"].CSPRNG {
		t.Error("D::m CSPRNG")
	}
	csprng := map[string]bool{"iv": true, "weak": false, "viaMethod": true, "viaStatic": true, "qualified": true}
	for _, fn := range fs.Functions {
		if fn.CSPRNG != csprng[fn.FQN] {
			t.Errorf("%s: CSPRNG %v", fn.FQN, fn.CSPRNG)
		}
	}
}

// thecodingmachine/safe wrappers are typed like the builtin, without the
// false (and preg_ null) they throw instead of returning.
func TestSafeFunctions(t *testing.T) {
	checkAnywhere(t, `<?php
namespace Safe {
/** @return string|array|null */
function preg_replace($p, $r, $s, int $l = -1, &$c = null) {}
/** @return string */
function file_get_contents(string $f) {}
function json_encode($v, int $f = 0) {}
/** @return int */
function strlen2($s) {}
/** @return \DateTime */
function substr($s, $o) {}
function no_such_builtin($x) {}
function str_replace($a, $b, $c) {}
}
namespace Safe\Sub { function strlen($s) {} }
namespace App {
use function Safe\preg_replace;
use function Safe\json_encode;
function f(string $s, array $a, $u = null) {
    t('pr', preg_replace('/x/', 'y', $s));
    t('prA', preg_replace('/x/', 'y', $a));
    t('fgc', \Safe\file_get_contents('x'));
    t('je', json_encode($a));
    t('better', \Safe\substr($s, 1));
    t('notBuiltin', \Safe\no_such_builtin(1));
    t('nested', \Safe\Sub\strlen('x'));
    t('unknownSubject', \Safe\str_replace('a', 'b', $u));
}
}
`, map[string]string{"pr": "string", "prA": "array", "fgc": "string", "je": "string", "better": `\DateTime`, "notBuiltin": "null", "nested": "null", "unknownSubject": "null"})
}

// include/require may reassign any local of the scope.
func TestIncludeClobbers(t *testing.T) {
	checkAnywhere(t, `<?php
function f() {
    $total = '0';
    include 'x.php';
    t('after', $total);
    $n = 1;
    t('later', $n);
}
`, map[string]string{"after": "?unknown", "later": "int"})
}

// The casting typer sees assignments later in a loop body on the next
// iteration.
func TestTRulesLoopBackEdge(t *testing.T) {
	src := `<?php
function f(array $rows, string $p) {
    $n = 0;
    foreach ($rows as $r) {
        t('n', $n);
        $n = (string) $r;
    }
    t('pathinfo', pathinfo($p, PATHINFO_FILENAME));
    t('safe', \Safe\preg_replace('/a/', 'b', $p));
}
namespace Safe;
/** @return string|array|null */
function preg_replace($p, $r, $s) {}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				want := map[string]string{"'n'": "int|string", "'pathinfo'": "string", "'safe'": "string"}[lit.Raw]
				if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != want {
					t.Errorf("%s: got %s want %s", lit.Raw, got, want)
				}
			}
		}
		return true
	})
}

// Argument-decided builtins, `??=` on literal arrays, refuted type checks,
// preg_match groups.
func TestRound6Builtins(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
function f(string $f, $id, string $s, array $rows, $flags) {
    t('pe', pathinfo($f, PATHINFO_EXTENSION));
    t('pa', pathinfo($f));
    t('pv', pathinfo($f, $flags));
    t('g0', gettimeofday());
    t('g1', gettimeofday(true));
    t('gf', gettimeofday(false));
    t('gv', gettimeofday($flags));
    $r = [];
    $r[$id] ??= [];
    t('coal', $r[$id]);
    $q = [];
    t('expr', $q[$id] ??= []);
    $w = ['a' => 1];
    t('expr2', $w[$id] ??= 'x');
    $by = [];
    foreach ($rows as $row) { $by[$row['k']] ??= []; $by[$row['k']][] = $row; t('nested', $by[$row['k']]); }
    $lit = ['a' => 1, 'b' => 'x'];
    t('computed', $lit[$id]);
    $u = [];
    t('emptyComputed', $u[$id]);
    $z = rand() ? ['a' => 1] : 'x';
    t('mixedBase', $z[$id]);
    t('literalBase', ['a' => 1, 'b' => 2][$id]);
    $x = [];
    [$x[$id]] = [1];
    t('unknownWrite', $x[$id]);
    $obj = (object) [];
    t('objDim', $obj->{'a'} ??= 1);
    t('propCoalesce', $obj->list[$id] ??= 1);
    /** @var Foo $o */
    $o = g();
    if (!is_object($o)) { t('notObj', $o); }
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
		"pe": "string", "pa": "array", "pv": "array|string", "g0": "array", "g1": "float", "gf": "array", "gv": "float|int[]",
		"coal": "array{}", "expr": "array{}", "expr2": "int|string", "nested": "array", "computed": "?unknown",
		"emptyComputed": "?unknown", "mixedBase": "?unknown", "literalBase": "int", "unknownWrite": "?unknown", "objDim": "?unknown", "propCoalesce": "?unknown", "notObj": "?unknown",
		"m": "string[]", "m2": "null[]|string[]", "m3": "null[][]|string[][]", "m4": "string[]", "m5": "array",
		"m6": "array", "m7": "array",
	})
}

// A definition always run earlier in the enclosing condition hides the
// definitions made before it (`preg_match(…, $m) && $m[1]`).
func TestDominatingConditionDefinitions(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $s, $c) {
    preg_match_all('/(a)/', $s, $m);
    if (preg_match('/(a)/', $s, $m, 0) && t('and', $m)) {}
    $x = 1;
    if (($x = 'a') && $c) { t('ifBody', $x); }
    $y = 1;
    while ($c && ($y = 'b')) { t('andCond', $y); }
    $y2 = 1;
    if ($c || ($y2 = 'b')) { t('conditional', $y2); }
    $n = 1;
    if (!($c || ($n = 'x'))) { t('notOr', $n); }
    $o = 1;
    $r = ($c || ($o = 'x')) || t('orFalse', $o);
    $z = 1;
    if ($c ?: ($z = 'c')) { t('ternary', $z); }
    $w = 1;
    $v = ($w = 'd') || t('or', $w);
    $a2 = 1;
    if (($a2 = 'e') || $c) { t('orLeft', $a2); }
    $q2 = 1;
    if (strlen($c ?? ($q2 = 'f'))) { t('coalesceArg', $q2); }
    $m2 = 1;
    if (match ($c) { 1 => ($m2 = 'g'), default => 0 }) { t('matchArm', $m2); }
}
`, map[string]string{
		"and": "string[]", "ifBody": "string", "andCond": "string", "conditional": "int|string", "notOr": "string", "orFalse": "string", "orLeft": "string", "coalesceArg": "int|string", "matchArm": "int|string", "ternary": "int|string", "or": "string",
	})
}
