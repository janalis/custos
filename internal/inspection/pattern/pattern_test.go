package pattern

import (
	"strings"
	"testing"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/fixing"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// run analyzes src (without its opening tag) with the rule d declares and
// returns its messages and the source after applying every fix.
func run(t *testing.T, d Definition, src string) ([]string, string) {
	t.Helper()
	r, err := Compile(d)
	if err != nil {
		t.Fatal(err)
	}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{ProjectPath: func(p string) string { return p }})
	if err != nil {
		t.Fatal(err)
	}
	full := []byte("<?php\n" + src)
	findings := e.Analyze(syntax.Parse("src/a.php", full, syntax.Options{}))
	var msgs []string
	var edits []diagnostic.TextEdit
	for _, f := range findings {
		msgs = append(msgs, f.Message)
		for _, fx := range f.Fixes {
			es, _ := fixing.Edits(fx)
			edits = append(edits, es...)
		}
	}
	out, _ := fixing.Apply(full, edits)
	return msgs, strings.TrimPrefix(string(out), "<?php\n")
}

func fix(s string) *string { return &s }

func TestMatch(t *testing.T) {
	cases := []struct {
		name  string
		def   Definition
		src   string
		count int
		fixed string
	}{
		{"function any args", Definition{Match: Patterns{`\dump(...)`}}, `dump(1, 2); \dump(); var_dump(1);`, 2, ""},
		{"function alias", Definition{Match: Patterns{`\App\dbg($X)`}}, `namespace B; use function App\dbg as d; d(1); dbg(2);`, 1, ""},
		{"const alias", Definition{Match: Patterns{`\App\LIMIT`}}, `namespace B; use const App\LIMIT as L; echo L, LIMIT, App\OTHER, \App\LIMIT;`, 2, ""},
		{"class name regex", Definition{Match: Patterns{`new $C()`}, Where: map[string]Where{"$C": {Regex: `^App\\Models\\`}}}, `namespace App\Http; use App\Models\User; new User(); new \App\Models\Post(); new Request(); new $v();`, 2, ""},
		{"function name regex", Definition{Match: Patterns{`$F($X)`}, Where: map[string]Where{"$F": {Regex: `^str`}}}, `namespace A; function strx($x) {} strlen($a); \strtoupper($b); strx($c); $v($d);`, 2, ""},
		{"not a name", Definition{Match: Patterns{`new $C()`}, Where: map[string]Where{"$C": {Not: Patterns{`\DateTimeImmutable`}}}}, `namespace A; use DateTimeImmutable as I; new I(); new \DateTime(); new $x();`, 2, ""},
		{"same name spelled twice", Definition{Match: Patterns{`new $C($C::DEFAULT)`}}, `namespace A; use B\D; new D(\B\D::DEFAULT); new D(E::DEFAULT);`, 1, ""},
		{"special class names", Definition{Match: Patterns{`$C::a() || $C::b()`}}, `static::a() || STATIC::b(); class K { function f() { static::a() || self::b(); self::a() || K::b(); } }`, 1, ""},
		{"function is not class", Definition{Match: Patterns{`new $C($C())`}}, `namespace A; new Foo(foo()); new Foo(Foo::make());`, 0, ""},
		{"namespaced fallback", Definition{Match: Patterns{`\strlen($X)`}}, `namespace A; strlen($a); \strlen($b);`, 2, ""},
		{"namespaced declared wins", Definition{Match: Patterns{`\strlen($X)`}}, `namespace A; function strlen($x) {} strlen($a);`, 0, ""},
		{"class through use", Definition{Match: Patterns{`new \DateTime()`}}, `namespace A; use DateTime as D; new D; new D(); new D('now'); new \DateTime;`, 3, ""},
		{"same metavariable", Definition{Match: Patterns{`$X === $X`}}, `$a === $a; $a === $b; $a->b === $a -> b;`, 2, ""},
		{"wildcard", Definition{Match: Patterns{`$_ === $_`}}, `$a === $b;`, 1, ""},
		{"literal variable", Definition{Match: Patterns{`$this->x`}}, `$this->x; $that->x; $this->y;`, 1, ""},
		{"string value", Definition{Match: Patterns{`\ini_set('display_errors', $V)`}}, `ini_set("display_errors", 1); ini_set('log', 1);`, 1, ""},
		{"method fold", Definition{Match: Patterns{`$X->getId()`}}, `$a->GETID(); $a->getName();`, 1, ""},
		{"parens", Definition{Match: Patterns{`$X == null`}}, `if (($a) == null) {}`, 1, ""},
		{"rest", Definition{Match: Patterns{`\f($A, ...$REST)`}}, `f(1); f(1, 2, 3); f();`, 2, ""},
		{"const", Definition{Match: Patterns{`$X == null`}}, `$a == NULL; $a == \null; $a == nul;`, 2, ""},
		{"static self", Definition{Match: Patterns{`\A\B::make()`}}, `namespace A; class B { function f() { self::make(); static::make(); } } B::make(); \B::make();`, 3, ""},
		{"statement", Definition{Match: Patterns{`return null;`}}, `function f() { return null; return 1; }`, 1, ""},
		{"regex", Definition{Match: Patterns{`\define($N, $_)`}, Where: map[string]Where{"$N": {Regex: `^'APP_`}}}, `define('APP_X', 1); define('X', 1);`, 1, ""},
		{"not", Definition{Match: Patterns{`\count($X)`}, Where: map[string]Where{"$X": {Not: Patterns{`$this->items`}}}}, `count($a); count($this->items);`, 1, ""},
		{"type", Definition{Match: Patterns{`$X->format($_)`}, Where: map[string]Where{"$X": {Type: `\DateTimeInterface`}}}, `function f(\DateTime $d, $o) { $d->format('Y'); $o->format('Y'); }`, 1, ""},
		{"paths", Definition{Match: Patterns{`\dump(...)`}, Paths: []string{"src/**"}, Exclude: []string{"src/b.php"}}, `dump();`, 1, ""},
		{"excluded", Definition{Match: Patterns{`\dump(...)`}, Exclude: []string{"src/*.php"}}, `dump();`, 0, ""},
		{"php gate", Definition{Match: Patterns{`\dump(...)`}, PHP: "8.5"}, `dump();`, 0, ""},

		{"fix swap", Definition{Match: Patterns{`$X == null`}, Fix: fix(`$X === null`)}, `$a == null;`, 1, `$a === null;`},
		{"fix parens", Definition{Match: Patterns{`\double($X)`}, Fix: fix(`$X * 2`)}, `$y = double($a + $b) + 1;`, 1, `$y = (($a + $b) * 2) + 1;`},
		{"fix no parens in arg", Definition{Match: Patterns{`\wrap($X)`}, Fix: fix(`\other($X)`)}, `wrap($a + $b);`, 1, `\other($a + $b);`},
		{"fix rest", Definition{Match: Patterns{`\old($A, ...$R)`}, Fix: fix(`\new_fn(...$R)`)}, `old(1, 2, 3); old(1);`, 2, `\new_fn(2, 3); \new_fn();`},
		{"fix rest first", Definition{Match: Patterns{`\old($A, ...$R)`}, Fix: fix(`\new_fn(...$R, $A)`)}, `old(1);`, 1, `\new_fn(1);`},
		{"fix delete", Definition{Match: Patterns{`\dump(...)`}, Fix: fix(``)}, "f();\ndump(1);\nif ($x) dump(2);", 2, "f();\nif ($x) dump(2);"},
		{"fix keeps comment", Definition{Match: Patterns{`\f($X, $Y)`}, Fix: fix(`\g($X)`)}, `f($a, /* keep */ $b);`, 1, `f($a, /* keep */ $b);`},
		{"fix in condition", Definition{Match: Patterns{`$X == null`}, Fix: fix(`$X === null`)}, `if ($a == null) {} while ($a == null) {} do {} while ($a == null); switch ($a == null) {} match ($a == null) {}; if (1) {} elseif ($a == null) {}`, 6, `if ($a === null) {} while ($a === null) {} do {} while ($a === null); switch ($a === null) {} match ($a === null) {}; if (1) {} elseif ($a === null) {}`},
		{"variable variable", Definition{Match: Patterns{`$$X`}}, `$$a; $a; ${'b'};`, 2, ""},
		{"named variable not dynamic", Definition{Match: Patterns{`$a->b`}}, `$$a->b; $a->b;`, 1, ""},
		{"literal kinds", Definition{Match: Patterns{`\f(1)`}}, `f('1'); f(1); f(1.0); f(2);`, 1, ""},
		{"float", Definition{Match: Patterns{`\f(1.5)`}}, `f(1.5); f(1.50);`, 1, ""},
		{"new args", Definition{Match: Patterns{`new \Foo`}}, `new Foo(1); new Foo; new Bar;`, 1, ""},
		{"new rest", Definition{Match: Patterns{`new \Foo(...$A)`}}, `new Foo; new Foo(1);`, 2, ""},
		{"new rest fix", Definition{Match: Patterns{`new \Foo(...$A)`}, Fix: fix(`\make(...$A)`)}, `new Foo;`, 1, `\make();`},
		{"new exact args", Definition{Match: Patterns{`new \Foo(1)`}}, `new Foo; new Foo(1);`, 1, ""},
		{"repeated rest", Definition{Match: Patterns{`[\f(...$R), \g(...$R)]`}}, `[f(1), g(1)]; [f(1), g(2)]; [f(1), g(1, 2)]; [f(), g()];`, 2, ""},
		{"rest vs single", Definition{Match: Patterns{`[\f($R), \g(...$R)]`}}, `[f(1), g(1)];`, 0, ""},
		{"single vs rest", Definition{Match: Patterns{`[\g(...$R), \f($R)]`}}, `[g(1), f(1)];`, 0, ""},
		{"too few args", Definition{Match: Patterns{`\f($A, $B)`}}, `f(1); f(1, 2, 3);`, 0, ""},
		{"const fallback", Definition{Match: Patterns{`\FOO`}}, `namespace A; FOO; \FOO; \A\FOO; BAR;`, 2, ""},
		{"const declared", Definition{Match: Patterns{`\FOO`}}, `namespace A; const FOO = 1; FOO;`, 0, ""},
		{"const namespaced", Definition{Match: Patterns{`\A\FOO`}}, `namespace A; const FOO = 1; FOO; \B\FOO;`, 1, ""},
		{"const case", Definition{Match: Patterns{`\FOO`}}, `foo;`, 0, ""},
		{"function namespaced", Definition{Match: Patterns{`\A\f()`}}, `namespace A; function f() {} f(); \A\f(); \f();`, 2, ""},
		{"function undeclared namespaced", Definition{Match: Patterns{`\A\f()`}}, `namespace A; f();`, 0, ""},
		{"parent", Definition{Match: Patterns{`\P::make()`}}, `class P {} class C extends P { function f() { parent::make(); } }`, 1, ""},
		{"isset count", Definition{Match: Patterns{`isset($X)`}}, `isset($a); isset($a, $b);`, 1, ""},
		{"isset items", Definition{Match: Patterns{`isset($a, $b)`}}, `isset($a, $c); isset($a, $b);`, 1, ""},
		{"operator", Definition{Match: Patterns{`$X + 1`}}, `$a + 1; $a - 1;`, 1, ""},
		{"nullsafe", Definition{Match: Patterns{`$X?->a`}}, `$b?->a; $b->a;`, 1, ""},
		{"push", Definition{Match: Patterns{`$a[] = $X`}}, `$a[] = 1; $a[0] = 1;`, 1, ""},
		{"regex rest", Definition{Match: Patterns{`\f(...$R)`}, Where: map[string]Where{"$R": {Regex: `^1`}}}, `f(1, 2); f(2);`, 1, ""},
		{"type on rest", Definition{Match: Patterns{`\f(...$R)`}, Where: map[string]Where{"$R": {Type: `int`}}}, `f(1);`, 0, ""},
		{"regex on rest only", Definition{Match: Patterns{`\f(...$R)`}, Where: map[string]Where{"$R": {Not: Patterns{"1"}}}}, `f(1);`, 0, ""},
		{"type statement", Definition{Match: Patterns{`echo $X;`}, Where: map[string]Where{"$X": {Type: `int`}}}, `echo 1;`, 1, ""},
		{"type rejects", Definition{Match: Patterns{`echo $X;`}, Where: map[string]Where{"$X": {Type: `int`}}}, `echo 'a';`, 0, ""},
		{"second pattern", Definition{Match: Patterns{`\a()`, `\b()`, `\c()`}}, `b(); c(); d();`, 2, ""},
		{"fix rest elsewhere", Definition{Match: Patterns{`\f(...$R)`}, Fix: fix(`$R`)}, `f(1);`, 1, `f(1);`},
		{"fix single spread", Definition{Match: Patterns{`\f($X)`}, Fix: fix(`\g(...$X)`)}, `f($a);`, 1, `\g(...$a);`},
		{"fix rest middle", Definition{Match: Patterns{`\f(...$R)`}, Fix: fix(`\g(...$R, 1)`)}, `f();`, 1, `\g(1);`},
		{"fix invalid", Definition{Match: Patterns{`\f($X)`}, Fix: fix(`$X = 1`)}, `f(g());`, 1, `f(g());`},
		{"fix template parens", Definition{Match: Patterns{`\f($X)`}, Fix: fix(`$X ?: 1`)}, `$y = 2 + f($a);`, 1, `$y = 2 + ($a ?: 1);`},
		{"fix dim", Definition{Match: Patterns{`\at($A, $I)`}, Fix: fix(`$A[$I]`)}, `at($a, $i + 1);`, 1, `$a[$i + 1];`},
		{"fix delete statement", Definition{Match: Patterns{`return null;`}, Fix: fix(``)}, "function f() {\n    return null;\n}", 1, "function f() {\n}"},
		{"fix delete keeps comment", Definition{Match: Patterns{`\dump(...)`}, Fix: fix(``)}, "dump(/* c */);", 1, "dump(/* c */);"},
		{"fix delete shared line", Definition{Match: Patterns{`\dump(...)`}, Fix: fix(``)}, "f(); dump(1);", 1, "f();"},
		{"fix delete trailing blanks", Definition{Match: Patterns{`\dump(...)`}, Fix: fix(``)}, "dump(1); \r\nf();", 1, "f();"},
		{"fix delete in expression", Definition{Match: Patterns{`\dump(...)`}, Fix: fix(``)}, "$a = dump();", 1, "$a = dump();"},
		{"fix rest after comma", Definition{Match: Patterns{`\f(...$R)`}, Fix: fix(`\g(1, ...$R)`)}, `f();`, 1, `\g(1);`},
		{"glob deep miss", Definition{Match: Patterns{`\dump()`}, Paths: []string{"**/x.php"}}, `dump();`, 0, ""},
		{"glob too long", Definition{Match: Patterns{`\dump()`}, Paths: []string{"src/a.php/b"}}, `dump();`, 0, ""},
		{"exclude miss", Definition{Match: Patterns{`\dump()`}, Exclude: []string{"lib/**"}}, `dump();`, 1, ""},
		{"target parens", Definition{Match: Patterns{`\f(1)`}}, `f((1));`, 1, ""},
		{"mixed kinds", Definition{Match: Patterns{`\a()`, `new \B`}}, `a(); new B; b();`, 2, ""},
		{"fix assigned", Definition{Match: Patterns{`!!!$X`}, Fix: fix(`!$X`)}, `$a = !!!$b;`, 1, `$a = !$b;`},
		{"fix assigned low", Definition{Match: Patterns{`\both($X, $Y)`}, Fix: fix(`$X and $Y`)}, `$a = both($b, $c); $a = \both($b, 1) ? 1 : 2;`, 2, `$a = ($b and $c); $a = ($b and 1) ? 1 : 2;`},
		{"fix assigned slot", Definition{Match: Patterns{`\set($X)`}, Fix: fix(`$v = $X`)}, `set($a or $b); set($a + 1);`, 2, `$v = ($a or $b); $v = $a + 1;`},
		{"fix trailing semicolon", Definition{Match: Patterns{`\foo($X)`}, Fix: fix(`\bar($X);`)}, `baz(foo($a)); foo($b);`, 2, `baz(\bar($a)); \bar($b);`},
		{"fix empty rest keeps other commas", Definition{Match: Patterns{`\old($A, ...$R)`}, Fix: fix(`\f(...$R) . \g(1, 2)`)}, `$x = old(1);`, 1, `$x = \f() . \g(1, 2);`},
		{"fix empty rest after comma elsewhere", Definition{Match: Patterns{`\old($A, ...$R)`}, Fix: fix(`\g($A, 1) . \f(...$R)`)}, `old(5);`, 1, `\g(5, 1) . \f();`},
		{"fix new class", Definition{Match: Patterns{`\make($X)`}, Fix: fix(`new $X`)}, `make(factory()); make($cls); make(Foo::class);`, 3, `new (factory()); new $cls; new (Foo::class);`},
		{"glob dot", Definition{Match: Patterns{`\dump()`}, Paths: []string{"./src/**"}}, `dump();`, 1, ""},
		{"glob dir", Definition{Match: Patterns{`\dump()`}, Exclude: []string{"src/"}}, `dump();`, 0, ""},
		{"glob root", Definition{Match: Patterns{`\dump()`}, Exclude: []string{"./"}}, `dump();`, 1, ""},
		{"array spelling", Definition{Match: Patterns{`\in_array($N, [$A])`}}, `in_array($n, array($a)); in_array($n, [$a]);`, 2, ""},
		{"exit spelling", Definition{Match: Patterns{`exit($X)`}}, `die(1); exit(2);`, 2, ""},
		{"include keyword kept", Definition{Match: Patterns{`include $X`}}, `include 'a'; require 'b';`, 1, ""},
		{"alt syntax", Definition{Match: Patterns{`if ($X) { \f(); }`}}, `if ($a): f(); endif; if ($b) { f(); }`, 2, ""},
		{"modifier order", Definition{Match: Patterns{`class A { public static function f() {} }`}}, `class A { static public function f() {} }`, 1, ""},
		{"modifier count", Definition{Match: Patterns{`class A { public static function f() {} }`}}, `class A { public function f() {} }`, 0, ""},
		{"modifier kind", Definition{Match: Patterns{`class A { public static function f() {} }`}}, `class A { public final function f() {} }`, 0, ""},
		{"sequence", Definition{Match: Patterns{`$F = \fopen($P, $M); $_; \fclose($F);`}}, "$f = fopen('a', 'r');\nread($f);\nfclose($f);\n$g = fopen('b', 'r');\nfclose($h);", 1, ""},
		{"sequence in block", Definition{Match: Patterns{`\lock(); \unlock();`}}, "function f() { lock(); unlock(); lock(); x(); unlock(); }\nlock(); unlock();", 2, ""},
		{"sequence in case", Definition{Match: Patterns{`\a(); \b();`}}, "switch ($x) { case 1: a(); b(); }", 1, ""},
		{"sequence leading wildcard", Definition{Match: Patterns{`$_; \a(); \b();`}}, "x(); a(); b();", 1, ""},
		{"sequence fix", Definition{Match: Patterns{`$X = \count($A); return $X;`}, Fix: fix(`return \count($A);`)}, "function f($a) {\n    $n = count($a);\n    return $n;\n}", 1, "function f($a) {\n    return \\count($a);\n}"},
		{"sequence delete", Definition{Match: Patterns{`\begin(); \end();`}, Fix: fix(``)}, "x();\nbegin();\nend();\ny();", 1, "x();\ny();"},
		{"statement metavariable", Definition{Match: Patterns{`if ($C) { $S; }`}, Fix: fix(`if ($C) $S;`)}, "if ($a) { f(); }\nif ($b) { return 1; }\nif ($c) { f(); g(); }", 2, "if ($a) f();\nif ($b) return 1;\nif ($c) { f(); g(); }"},
		{"statement into expression", Definition{Match: Patterns{`if ($C) { $S; }`}, Fix: fix(`\run($S);`)}, "if ($a) { f(); }", 1, "if ($a) { f(); }"},
		{"block wildcard", Definition{Match: Patterns{`foreach ($A as $V) { $_; $V->save(); $_; }`}}, "foreach ($xs as $x) { $x->a = 1; $x->save(); }\nforeach ($xs as $x) { $y->save(); }", 1, ""},
		{"statements template in body", Definition{Match: Patterns{`return \both();`}, Fix: fix(`\a(); return \b();`)}, "function f($x) {\n    return both();\n    if ($x) return both();\n}", 2, "function f($x) {\n    \\a(); return \\b();\n    if ($x) return both();\n}"},
		{"statement from expression", Definition{Match: Patterns{`if ($C) { $S; }`}, Fix: fix(`\when($C, $S)`)}, "if ($a) { f(); }", 1, "if ($a) { f(); }"},
		{"expression replaces statement", Definition{Match: Patterns{`return \old();`}, Fix: fix(`\done()`)}, "function f() { return old(); }", 1, "function f() { \\done(); }"},
		{"sequence of methods", Definition{Match: Patterns{`$X->open(); $X->close();`}}, "$f->open(); $f->close();\n$f->open(); $g->close();", 1, ""},
		{"sequence php gate", Definition{Match: Patterns{`$X->open(); $X->close();`}, PHP: "8.5"}, "$f->open(); $f->close();", 0, ""},
		{"closure uses", Definition{Match: Patterns{`\run(function () { $_; })`}}, "run(function () use ($a) { f(); });\nrun(function () { });", 2, ""},
		{"closure uses written", Definition{Match: Patterns{`\run(function () use ($b) { $_; })`}}, "run(function () use ($a) { f(); });\nrun(function () use ($b) { });", 1, ""},
		{"hoist statement", Definition{Match: Patterns{`if (true) { $S; }`}, Fix: fix(`$S;`)}, "if (true) { f(); }\nif (true) { if ($a) { g(); } }", 2, "f();\nif ($a) { g(); }"},
		{"closure replaces statement", Definition{Match: Patterns{`return \old();`}, Fix: fix(`$f = function () {}`)}, "function g() { return old(); }", 1, "function g() { $f = function () {}; }"},
		{"globals", Definition{Match: Patterns{`$GLOBALS['x']`}}, "$GLOBALS['x']; $arr['x'];", 1, ""},
		{"not on statement", Definition{Match: Patterns{`if ($C) { $S; }`}, Where: map[string]Where{"$S": {Not: Patterns{`\g()`}}}}, "if ($q) { g(); }\nif ($q) { h(); }", 1, ""},
		{"inside", Definition{Match: Patterns{`$X->save()`}, Inside: Patterns{`foreach ($A as $X) { $_; }`}}, "foreach ($xs as $x) { $x->save(); $y->save(); }\n$x->save();", 1, ""},
		{"inside any", Definition{Match: Patterns{`\sleep($_)`}, Inside: Patterns{`while ($_) { $_; }`, `for ($_; $_; $_) { $_; }`}}, "while (1) { sleep(1); }\nfor ($i = 0; $i < 2; $i++) { sleep(1); }\nsleep(1);", 2, ""},
		{"not inside", Definition{Match: Patterns{`$X->query($Q)`}, NotInside: Patterns{`$X->transaction(function () { $_; })`}}, "$db->transaction(function () { $db->query(1); });\n$other->transaction(function () { $db->query(2); });\n$db->query(3);", 2, ""},
		{"fix new", Definition{Match: Patterns{`new \DateTimeImmutable()`}, Fix: fix(`$this->clock->now()`)}, `$d = new \DateTimeImmutable();`, 1, `$d = $this->clock->now();`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.def
			d.ID, d.Message = "MyRule", "Found"
			msgs, fixed := run(t, d, tc.src)
			if len(msgs) != tc.count {
				t.Fatalf("got %d findings %v, want %d", len(msgs), msgs, tc.count)
			}
			if tc.fixed != "" && fixed != tc.fixed {
				t.Fatalf("fixed:\n%s\nwant:\n%s", fixed, tc.fixed)
			}
		})
	}
}

func TestMessageRunes(t *testing.T) {
	msgs, _ := run(t, Definition{ID: "R", Match: Patterns{`\f($X)`}, Message: "$X"}, "f('a"+strings.Repeat("é", 40)+"');")
	if len(msgs) != 1 || msgs[0] != "'a"+strings.Repeat("é", 27)+"..." {
		t.Fatalf("got %q", msgs)
	}
}

func TestMessage(t *testing.T) {
	msgs, _ := run(t, Definition{ID: "R", Match: Patterns{`$X == $Y`}, Message: "Compare $X\nwith $Y"}, "$a\n  ->b == '"+strings.Repeat("x", 80)+"';")
	want := "Compare $a ->b\nwith '" + strings.Repeat("x", 56) + "..."
	if len(msgs) != 1 || msgs[0] != want {
		t.Fatalf("got %q, want %q", msgs, want)
	}
}
