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

// typeAt infers the type of the expression statement following `// @type`.
func check(t *testing.T, src string, want map[string]string) {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if es, ok := n.(*syntax.ExprStmt); ok {
			if c, ok := es.Expr.(*syntax.FuncCall); ok {
				if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" && len(c.Args.Args) == 2 {
					label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
					got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
				}
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

func TestInfer(t *testing.T) {
	check(t, `<?php
namespace App;
class Item { public function name(): string { return ''; } public static function make(): static { return new static(); } }
class Repo {
    /** @var Item[] */
    private array $items = [];
    public ?Item $last = null;
    /** @return Item[] */
    public function all(): array { return $this->items; }
    public function run(int $n, ?string $s, array $list, \DateTime ...$dates) {
        $a = 1; $b = 1.5; $c = $a + 2; $d = $a . 'x'; $e = $n > 2;
        t('a', $a); t('c', $c); t('d', $d); t('e', $e); t('n', $n); t('s', $s);
        t('this', $this); t('items', $this->items); t('last', $this->last);
        foreach ($this->all() as $item) { t('item', $item); t('name', $item->name()); }
        t('strlen', strlen($d)); t('str_replace', str_replace('a', 'b', $d));
        t('new', new Item()); t('static', Item::make()); t('class', Item::class);
        t('coalesce', $s ?? 'x'); t('dates', $dates); t('ternary', $e ? 1 : 'a');
        $x = $this->last?->name(); t('nullsafe', $x);
        /** @var Item $hinted */
        $hinted = make_it(); t('hinted', $hinted);
        t('dt', (new \DateTime())->format('Y'));
        t('cast', (int) $s); t('arr', [1, 2]); t('mixedarr', [1, 'a']);
    }
}
`, map[string]string{
		"a": "int", "c": "int", "d": "string", "e": "bool", "n": "int", "s": "null|string",
		"this": `\App\Repo`, "items": `\App\Item[]`, "last": `\App\Item|null`,
		"item": `\App\Item`, "name": "string", "strlen": "int", "str_replace": "string",
		"new": `\App\Item`, "static": `\App\Item`, "class": "string", "coalesce": "string",
		"dates": `\DateTime[]`, "ternary": "int|string", "nullsafe": "null|string",
		"hinted": `\App\Item`, "dt": "string", "cast": "int", "arr": "int[]", "mixedarr": "array",
	})
}

func TestBodyReturnType(t *testing.T) {
	src := `<?php
function a(?int $i) { return $i; }
function b() { if (true) { return 1; } return 'x'; }
function c(): int { return 1; }
function d($u) { return $u; }
function g() { yield 1; }
a(); b(); c(); d(1); g();
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	want := map[string]string{"a": "int|null", "b": "int|string", "c": "?unknown", "d": "?unknown", "g": "?unknown"}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			name := c.Name.(*syntax.Name).Value
			if got := env.BodyReturnType(env.ResolveFunction(c)).String(); got != want[name] {
				t.Errorf("%s: got %s want %s", name, got, want[name])
			}
		}
		return true
	})
}

func TestNullableArrayDim(t *testing.T) {
	check(t, `<?php
class Box {}
/** @param Box[]|null $xs */
function f($xs) {
    t('elem', $xs[0]);
}
`, map[string]string{"elem": `\Box`})
}

func TestInlineVarOverridesEarlierDefinitions(t *testing.T) {
	check(t, `<?php
function f(int $a, array $b) {
    /** @var int[] $b */
    t('b', $b);
    $c = 1;
    /** @var string $c */
    t('c', $c);
}
`, map[string]string{"b": "int[]", "c": "string"})
}

func TestSelfReassignment(t *testing.T) {
	check(t, `<?php
class Inv {}
function f(?Inv $first, Inv $second = null) {
    $first = $first ?: null;
    t('first', $first);
    $second = $second ?? null;
    t('second', $second);
    $n = 1;
    $n = $n + 1;
    t('n', $n);
}
`, map[string]string{"first": `\Inv|null`, "second": `\Inv|null`, "n": "int"})
}

func TestNarrowing(t *testing.T) {
	check(t, `<?php
class Foo { public function run(): int { return 1; } }
function g(array|bool $a, ?Foo $f, int|string $u) {
    t('tern', is_array($a) ? $a : null);
    if (is_array($a) && count($a) > 0) { t('ifand', $a); }
    if ($f !== null) { t('ifnn', $f); } else { t('elsenull', $f); }
    if (!is_string($u)) { t('notstr', $u); }
    t('fcc', strlen(...));
    t('mfcc', (new Foo())->run(...));
    if (null === $f) { return; }
    t('guard', $f);
}
`, map[string]string{"tern": "array|null", "ifand": "array", "ifnn": `\Foo`, "elsenull": "null",
		"notstr": "int", "fcc": `\Closure`, "mfcc": `\Closure`, "guard": `\Foo`})
}

func TestAliasesAndFalse(t *testing.T) {
	check(t, `<?php
/**
 * @phpstan-type Row array{path: string, line: int}
 * @psalm-import-type Other from Somewhere
 */
final class Panel {
    /** @param Row $r @param Other $o */
    public function show($r, $o, string $url) {
        t('row', $r);
        $parts = parse_url($url);
        t('parts', false === $parts ? null : $parts);
    }
}
`, map[string]string{"row": "array", "parts": "array|null"})
}

func TestNarrowRealCases(t *testing.T) {
	check(t, `<?php
function h(string $url, $x) {
    $parts = \parse_url($url);
    t('tern_else', false === $parts ? '' : $parts);
    /** @var array|bool $found */
    $found = $x;
    t('and_tern', \is_array($found) && [] !== $found ? $found : null);
}
`, map[string]string{"tern_else": "array|string", "and_tern": "array|null"})
}

func TestElementWritesWidenElementType(t *testing.T) {
	check(t, `<?php
function f(array $names) {
    $schema = ['ref' => 'x'];
    $schema['req'] ??= [];
    t('coalesce', $schema['req']);
    $plain = ['a' => 'x'];
    $plain['b'] = 'y';
    t('same', $plain['b']);
    $list = ['a' => 'x'];
    foreach ($names as $n) {
        t('loop', $list['b']);
        $list['b'] = [];
    }
}
`, map[string]string{"coalesce": "array|string", "same": "string", "loop": "array|string"})
}

func TestArrayPopReturnsNullWhenEmpty(t *testing.T) {
	check(t, `<?php
/** @param string[] $names */
function f(array $names) {
    t('pop', array_pop($names));
    t('shift', array_shift($names));
    t('end', end($names));
}
`, map[string]string{"pop": "null|string", "shift": "null|string", "end": "false|string"})
}

func TestMixedArrayDimIsUnknown(t *testing.T) {
	check(t, `<?php
class Box {}
/** @param array|Box[] $xs */
function f($xs, array $raw, bool $c) {
    t('elem', $xs[0]);
    $d = $c ? $raw : ['k' => $raw];
    t('mixed', $d['k']);
}
`, map[string]string{"elem": "?unknown", "mixed": "?unknown"})
}

func TestAbsAndNamedSubject(t *testing.T) {
	check(t, `<?php
function f(int $a, int $b, float $x, string $s) {
    t('absint', abs($a - $b));
    t('absfloat', abs($x));
    t('named', str_replace(['a'], replace: 'b', subject: $s));
}
`, map[string]string{"absint": "int", "absfloat": "float", "named": "string"})
}

func TestLoopCarriedDefinitions(t *testing.T) {
	check(t, `<?php
function f(array $lines) {
    $previous = null;
    foreach ($lines as $line) {
        t('carried', $previous);
        $previous = strlen($line);
    }
    t('after', $previous);
}
`, map[string]string{"carried": "int|null", "after": "int|null"})
}

func TestReplaceOnScalarSubject(t *testing.T) {
	check(t, `<?php
function f(array $map) {
    foreach ($map as $key => $v) {
        t('preg', preg_replace('/x/', 'y', $key));
        t('str', str_replace('x', 'y', $key));
    }
}
`, map[string]string{"preg": "null|string", "str": "string"})
}

func TestGuardedOverwrite(t *testing.T) {
	check(t, `<?php
function f(string $line, int $len, ?string $name) {
    $end = strpos($line, 'x');
    if (false === $end) {
        $end = $len;
    }
    t('end', $end);
    if (null === $name) {
        $name = 'anon';
    }
    t('name', $name);
    $pos = strpos($line, 'y');
    if (false === $pos) {
        $pos = false;
    }
    t('kept', $pos);
    if (false === $cut = strrpos($line, 'z')) {
        $cut = $len;
    }
    t('inline', $cut);
}
`, map[string]string{"end": "int", "name": "string", "kept": "false|int", "inline": "int"})
}

func TestParseURL(t *testing.T) {
	check(t, `<?php
function f(string $u) {
    t('all', parse_url($u));
    t('port', parse_url($u, PHP_URL_PORT));
    t('host', parse_url($u, PHP_URL_HOST));
    if (false === $parts = parse_url($u)) {
        return;
    }
    t('parts', $parts);
}
`, map[string]string{"all": "array|false", "port": "false|int|null", "host": "false|null|string", "parts": "array"})
}

func TestUnconditionalAssignmentHidesEarlierDefinitions(t *testing.T) {
	check(t, `<?php
function f(array $lines, bool $c) {
    $rows = count($lines);
    t('first', $rows);
    $rows = [];
    foreach ($lines as $line) {
        t('nested', $rows);
        $u = 'x';
        t('reset', $u);
        $u = 5;
    }
    $v = 'a';
    if ($c) {
        $v = 1;
    }
    t('branch', $v);
}
`, map[string]string{"first": "int", "nested": "array", "reset": "string", "branch": "int|string"})
}

func TestWhileConditionAndHrtime(t *testing.T) {
	check(t, `<?php
function f($fh) {
    while (false !== $row = fgetcsv($fh)) {
        t('row', $row);
    }
    if (false === $t = hrtime()) {
        return;
    }
    t('hr', $t);
}
`, map[string]string{"row": "array|null", "hr": "int[]"})
}

func TestHrtimeVariants(t *testing.T) {
	check(t, `<?php
function f() {
    t('none', hrtime());
    t('false', hrtime(false));
    t('qualified', \hrtime(\FALSE));
    t('true', hrtime(true));
    t('named', hrtime(as_number: true));
}
`, map[string]string{"none": "false|int[]", "false": "false|int[]", "qualified": "false|int[]", "true": "false|float|int", "named": "false|float|int"})
}

func TestForeachOverMixedIterable(t *testing.T) {
	check(t, `<?php
class Box {}
/** @param Box[]|null|false $boxes */
function f(string|iterable $needle, $boxes, bool $c, $x) {
    if (\is_string($needle)) {
        $needle = [$needle];
    }
    foreach ($needle as $n) {
        t('mixed', $n);
    }
    foreach ($boxes as $b) {
        t('typed', $b);
    }
}
`, map[string]string{"mixed": "?unknown", "typed": `\Box`})
}

func TestBranchJoin(t *testing.T) {
	check(t, `<?php
function j(int|string $p, bool $c) {
    $v = $p;
    if ($c) { $v = 'a'; } elseif ($p) { $v = 'b'; } else { $v = 'c'; }
    t('joined', $v);
    $w = $p;
    if ($c) { $w = 'x'; }
    t('partial', $w);
    $z = $p;
    if ($c) { $z = 1; } else { throw new \Exception(); }
    t('leave', $z);
}
`, map[string]string{"joined": "string", "partial": "int|string", "leave": "int"})
}

func TestByRefClosureImport(t *testing.T) {
	check(t, `<?php
class Item {}
function f(callable $run, Item $i) {
    $captured = null;
    $byValue = null;
    $cb = function () use (&$captured, $byValue) { $captured = [1, 2]; };
    t('before', $captured);
    $run($cb);
    t('captured', $captured);
    t('byValue', $byValue);
    $captured = $i;
    t('reassigned', $captured);
}
`, map[string]string{"captured": "?unknown", "byValue": "null", "reassigned": `\Item`})
}

func TestThisPropertyNarrowing(t *testing.T) {
	check(t, `<?php
class Ed {
    public ?string $abbr = null;
    public ?Ed $parent = null;
    public function f() {
        if ($this->abbr) { t('truthy', $this->abbr); }
        t('plain', $this->abbr);
        if (null === $this->parent) { return; }
        t('guarded', $this->parent);
        $this->parent = null;
        t('reassigned', $this->parent);
    }
}
`, map[string]string{"truthy": "string", "plain": "null|string", "guarded": `\Ed`, "reassigned": `\Ed|null`})
}

func TestInlineVarAttachedStatementOnly(t *testing.T) {
	check(t, `<?php
function f(string $from) {
    /** @var string $from */
    $ids = explode(',', $from);
    $from = [];
    t('later', $from);
    /** @var int $n */
    $n = g();
    t('attached', $n);
}
`, map[string]string{"later": "array", "attached": "int"})
}
