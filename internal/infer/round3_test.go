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

// The back edge of a do-while (or for step) re-enters the body only when
// the condition held: `while ($e = $e->getPrevious())` loops with a truthy
// $e.
func TestDoWhileBackEdge(t *testing.T) {
	checkAnywhere(t, `<?php
function f(\Exception $e, ?array $rows, $c) {
    $all = [];
    do {
        $all[] = $e;
        t('body', $e);
    } while ($e = $e->getPrevious());
    t('all', $all);
    t('after', $e);
    for ($r = $rows; $r !== null; $r = $c ? null : $rows) {
        t('for', $r);
    }
    do { t('plain', $c); $c = rand() ? null : 1; } while ($c);
    for ($n = $rows; $n !== null; $n = null) { t('nullStep', $n); }
}
`, map[string]string{
		"body": `\Exception|\Throwable`, "all": `\Exception[]|\Throwable[]`, "after": `\Exception|\Throwable|null`, // the pre-loop value is not known to be overwritten
		"for": "array", "plain": "?unknown", "nullStep": "array",
	})
}

// Static properties narrow like `$this->prop`, and lose the fact on any
// call or write in between.
func TestStaticPropertyNarrowing(t *testing.T) {
	checkAnywhere(t, `<?php
class W {}
class S {
    /** @var W[] */
    private static array $stack = [];
    private static ?W $cur = null;
    public static function f() {
        if (!empty(self::$stack)) {
            $w = array_pop(self::$stack);
            t('pop', $w);
        }
        if (!empty(static::$stack)) { t('static', static::$stack); }
        if (self::$cur === null) { return; }
        t('cur', self::$cur);
        self::reset();
        t('afterCall', self::$cur);
        if (S::$cur !== null) { t('named', S::$cur); }
        self::$cur = new W();
        t('written', self::$cur);
    }
    public static function reset(): void {}
}
`, map[string]string{
		"pop": `\W`, "static": `non-empty \W[]`, "cur": `\W`, "afterCall": `\W`, // as for $this->p, an early-exit guard survives calls "named": `\W`, "written": `\W`,
	})
}

// A standalone `@var Class $x` over a variable holding a class-name
// string describes the class, not the value (Yii ActiveField::widget()).
func TestInlineVarClassNameIdiom(t *testing.T) {
	checkAnywhere(t, `<?php
class Widget {}
/** @param string $class */
function f($class, array $config, $u) {
    /** @var Widget $class */
    $config['model'] = 1;
    t('class', $class);
    $n = 1;
    /** @var string $n */
    t('override', $n);
    /** @var Widget $u */
    t('unknown', $u);
    /** @var Widget $fresh */
    t('fresh', $fresh);
}
`, map[string]string{"class": "string", "override": "string", "unknown": `\Widget`, "fresh": `\Widget`})
}

// A then-branch reassignment does not reach the else branch, so the
// condition still narrows there; definitions on paths that always leave
// (return, throw, exit, break out of the region) reach nothing after.
func TestExclusiveBranchesAndExits(t *testing.T) {
	checkAnywhere(t, `<?php
function f(array|int $c, $cc, int $i) {
    if (is_int($c)) { $c = []; } else { t('else', $c); }
    $v = $cc ? ($i = 'a') : t('ternary', $i);
    $m = match (true) { $cc === 1 => $i = 2.5, default => t('arm', $i) };
    while ($cc) {
        if (is_int($c)) { $c = 'x'; } else { t('loopElse', $c); }
    }
    $x = 5;
    if ($cc) { $x = 'a'; return 1; }
    t('afterReturn', $x);
    $y = 1;
    if ($cc) { $y = 'b'; throw new \Exception(); }
    t('afterThrow', $y);
    $z = 1;
    if ($cc) { $z = 'c'; } else { $z = 2.5; exit; }
    t('afterIfElse', $z);
    try {
        $t = 1;
        if ($cc) { $t = 'd'; throw new \Exception(); }
    } catch (\Exception $e) { t('catch', $t); }
    foreach ([1] as $k) {
        $w = 1;
        if ($cc) { $w = 'e'; break; }
        t('inLoop', $w);
    }
    t('afterBreak', $w);
    foreach ([1] as $k) {
        foreach ([2] as $l) {
            if ($cc) { $q = 'f'; break; }
        }
        t('outerLoop', $q);
        $q = 1;
    }
    foreach ([1] as $k) {
        $p = 1;
        if ($cc) { $p = 'g'; continue; }
        t('continued', $p);
    }
    $o = 1;
    foreach ([1] as $k) {
        t('nextIteration', $o);
        if ($cc) { $o = 'j'; continue; }
        if ($k) { $o2 = 2.5; continue; }
        t('viaHead', $o);
    }
    switch ($cc) {
        case 1:
            switch ($i) {
                case 2:
                    $n = 'k';
                    break;
                default:
                    t('innerSwitch', $n ?? 0);
            }
            t('afterInner', $n ?? 0);
    }
    for ($r = 0; ; $r = 'l') { t('noCond', $r); break; }
    switch ($cc) {
        case 1:
            $s = 'h';
            break;
        default:
            t('switchDefault', $s ?? null);
    }
    $u = 1;
    if ($cc) { $u = 'i'; return; t('dead', $u); }
}
`, map[string]string{
		"else": "array", "ternary": "int", "arm": "int|string", "loopElse": "array|int|string",
		"afterReturn": "int", "afterThrow": "int", "afterIfElse": "string", "catch": "int|null|string",
		"inLoop": "int", "afterBreak": "int|null|string", "outerLoop": "int|null|string", "continued": "int",
		"nextIteration": "int|string", "viaHead": "int|string", "innerSwitch": "?unknown", "afterInner": "int|string", "noCond": "int|string",
		"switchDefault": "?unknown", "dead": "string",
	})
}

// Stub PHPDoc never widens a builtin's version-resolved native return.
func TestBuiltinDocDoesNotWiden(t *testing.T) {
	src := `<?php
function f(string $s) {
    t('substr', substr($s, 1));
    t('split', str_split($s));
    t('explode', explode(',', $s));
    t('cb', array_map('substr', [$s], [1]));
}
`
	checkVer(t, phpver.PHP81, false, src, map[string]string{"substr": "string", "split": "string[]", "explode": "string[]", "cb": "string[]"})
	checkVer(t, phpver.PHP74, false, src, map[string]string{"substr": "false|string", "split": "false|string[]"})
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP81})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP81))
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
				want := map[string]string{"'substr'": "string", "'split'": "array|string[]", "'explode'": "array", "'cb'": "array"}[lit.Raw]
				if got != want {
					t.Errorf("T-rules %s: got %s want %s", lit.Raw, got, want)
				}
			}
		}
		return true
	})
}

// `@method` tags describe only what no real declaration provides.
func TestMagicMethodTags(t *testing.T) {
	checkAnywhere(t, `<?php
class Base { public function find(): ?object { return null; } }
/**
 * @method static int foo()
 * @method string bar()
 * @method \Base find()
 */
class M extends Base { public function foo(): string { return ''; } }
function f(M $m) {
    t('own', $m->foo());
    t('magic', $m->bar());
    t('inherited', $m->find());
}
`, map[string]string{"own": "string", "magic": "string", "inherited": "null|object"})
	f := syntax.Parse("t.php", []byte(`<?php
/** @method static self make() */
class F {}
`), syntax.Options{Version: phpver.PHP84})
	fs := index.Extract(f)
	m := fs.Classes[0].Methods["make"]
	if m == nil || !m.Magic || !m.Static {
		t.Fatalf("make: %+v", m)
	}
}

// Many exit regions and branches stay linear.
func TestExitRegionsBounded(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\nfunction f($c, $x) {\n")
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&b, " $x = %d; if ($c === %d) { $x = 'a'; return; } if (is_int($x)) { $x = 1; } else { $y = $x; }\n", i, i)
	}
	b.WriteString(" return $x;\n}\n")
	if d := typeAll(t, b.String()); d > testbudget.Of(2*time.Second) && !raceEnabled {
		t.Errorf("took %v", d)
	}
}

// A user declaration's doc may widen its declared type (memberType picks
// the declared one); a builtin's never does.
func TestUserDocVsDeclared(t *testing.T) {
	checkAnywhere(t, `<?php
/** @return string|false */
function u(): string { return ''; }
function f() { t('user', u()); t('cb', array_map('u', [1])); t('strlen', strlen('x')); t('ob', ob_get_clean()); }
`, map[string]string{"user": "string", "cb": "string[]", "strlen": "int", "ob": "false|string"})
}
