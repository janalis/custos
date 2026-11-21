package syntax

import (
	"testing"

	phpversion "github.com/janalis/custos/internal/php/version"
)

var allVersions = []phpversion.Version{
	phpversion.PHP53, phpversion.PHP54, phpversion.PHP55, phpversion.PHP56,
	phpversion.PHP70, phpversion.PHP71, phpversion.PHP72, phpversion.PHP73, phpversion.PHP74,
	phpversion.PHP80, phpversion.PHP81, phpversion.PHP82, phpversion.PHP83, phpversion.PHP84, phpversion.PHP85,
}

func previousVersion(v phpversion.Version) phpversion.Version {
	for i, o := range allVersions {
		if o == v && i > 0 {
			return allVersions[i-1]
		}
	}
	return 0
}

// requiredVersionCases pair a snippet with the lowest PHP version it needs.
// Unless semantic is set (older versions parse it with another meaning),
// php -l of that version must accept it and php -l of the previous version
// must reject it, whenever those binaries are installed.
var requiredVersionCases = []struct {
	src      string
	want     phpversion.Version
	semantic bool
}{
	// baseline
	{`echo 1; $a = array(1, 2); f($a); $o->m(); A::m(); $f = function ($x) use ($y) { return $x; };`, phpversion.PHP53, false},
	{`list($a, $b) = $c; list($a, list($b)) = $c; $a = &$b; foreach ($x as $k => $v) {}`, phpversion.PHP53, false},
	{`$a = -1; $b = +1.5; $c = 0x1F; $d = 017; $s = 'a\u{1F}'; $t = "\\u{1F}"; $n = <<<'N'
\u{1F}
N;
`, phpversion.PHP53, false},
	{"$h = <<<EOT\nx\nEOT;\n$i = <<<EOT\nx\nEOT\n;", phpversion.PHP53, false},
	{`$c = clone $o; $d = clone($o); $o = new A; $p = new A(); $x = $a['k'](); $y = $$n(); $z = $o->p;`, phpversion.PHP53, false},
	{`try { f(); } catch (A $e) {} function f(array $a = array(1), A $b = null, $c = A::B, $d = -1) {} static $s = 1;`, phpversion.PHP53, false},
	{`class C { const X = 1; var $p = array(); public static $q = PHP_EOL; function m(self $s) {} } interface I { const Y = 2; }`, phpversion.PHP53, false},
	{`use A\B, C\D as E; empty($a); empty($a['k']); empty($o->p); empty(A::$p); throw new E();`, phpversion.PHP53, false},
	{`echo __LINE__, __DIR__; isset($a, $b); unset($a);`, phpversion.PHP53, false},
	{`list(, $b) = $c; list($a, list(, $b)) = $c;`, phpversion.PHP53, false},
	{`[, $b] = $c;`, phpversion.PHP71, false},
	{`$s = "a$b"; $t = "{$o->p}"; $f = (int) $x; $g = !$y; $h = A::$p; $i = $a ? : $b;`, phpversion.PHP53, false},
	// 5.4
	{`$a = [1, 2];`, phpversion.PHP54, false},
	{`trait T {}`, phpversion.PHP54, false},
	{`class C { use T; }`, phpversion.PHP54, false},
	{`$f = static function () {};`, phpversion.PHP54, false},
	{`function f(callable $c) {}`, phpversion.PHP54, true},
	{`f()[0];`, phpversion.PHP54, false},
	{`$o->m()[0];`, phpversion.PHP54, false},
	{`A::m()[0];`, phpversion.PHP54, false},
	{`(new A)->m();`, phpversion.PHP54, false},
	{`(new A)->p;`, phpversion.PHP54, false},
	{`(new A)['k'];`, phpversion.PHP54, false},
	{`[$o, 'm']();`, phpversion.PHP54, false},
	{`$b = 0b101;`, phpversion.PHP54, false},
	{`echo __TRAIT__;`, phpversion.PHP54, false},
	// 5.5
	{`function g() { yield 1; }`, phpversion.PHP55, false},
	{`try {} finally {}`, phpversion.PHP55, false},
	{`echo A::class;`, phpversion.PHP55, false},
	{`echo static::class;`, phpversion.PHP55, false},
	{`foreach ($x as list($a, $b)) {}`, phpversion.PHP55, false},
	{`empty(f());`, phpversion.PHP55, false},
	{`echo array(1)[0];`, phpversion.PHP55, false},
	{`echo 'abc'[0];`, phpversion.PHP55, false},
	// 5.6
	{`function f(...$a) {}`, phpversion.PHP56, false},
	{`f(...$a);`, phpversion.PHP56, false},
	{`$a = 2 ** 3;`, phpversion.PHP56, false},
	{`$a **= 3;`, phpversion.PHP56, false},
	{`use function A\b;`, phpversion.PHP56, false},
	{`use const A\B;`, phpversion.PHP56, false},
	{`const A = 1 + 2;`, phpversion.PHP56, false},
	{`const A = array(1);`, phpversion.PHP56, false},
	{`class C { const A = array(1); }`, phpversion.PHP56, false},
	{`class C { public $p = 1 << 2; }`, phpversion.PHP56, false},
	{`function f($a = 1 ? 2 : 3) {}`, phpversion.PHP56, false},
	{`function f() { static $s = !true; }`, phpversion.PHP56, false},
	{`const A = (1);`, phpversion.PHP56, false},
	{`class C { const A = B[0]; }`, phpversion.PHP56, false},
	// 7.0
	{`$a = $b ?? $c;`, phpversion.PHP70, false},
	{`$a = $b <=> $c;`, phpversion.PHP70, false},
	{`function f(int $a) {}`, phpversion.PHP70, true},
	{`function f(float $a, string $b, bool $c) {}`, phpversion.PHP70, true},
	{`function f(): array {}`, phpversion.PHP70, false},
	{`class C { function m(): self {} }`, phpversion.PHP70, false},
	{`$f = function (): A {};`, phpversion.PHP70, false},
	{`$o = new class {};`, phpversion.PHP70, false},
	{`function g() { yield from h(); }`, phpversion.PHP70, false},
	{`use A\{B, C};`, phpversion.PHP70, false},
	{`use A\{B, function c};`, phpversion.PHP70, false},
	{`echo "\u{1F600}";`, phpversion.PHP70, true},
	{"echo <<<EOT\n\\u{1F600} $x\nEOT;\n", phpversion.PHP70, true},
	{`(function () {})();`, phpversion.PHP70, false},
	{`f()();`, phpversion.PHP70, false},
	{`$o->m()();`, phpversion.PHP70, false},
	{`'strlen'('x');`, phpversion.PHP70, false},
	// 7.1
	{`function f(?int $a) {}`, phpversion.PHP71, false},
	{`function f(): void {}`, phpversion.PHP71, true},
	{`function f(iterable $a) {}`, phpversion.PHP71, true},
	{`[$a, $b] = $c;`, phpversion.PHP71, false},
	{`[[$a], [$b]] = $c;`, phpversion.PHP71, false},
	{`foreach ($x as [$a, $b]) {}`, phpversion.PHP71, false},
	{`list('k' => $a) = $c;`, phpversion.PHP71, false},
	{`class C { private const X = 1; }`, phpversion.PHP71, false},
	{`try {} catch (A | B $e) {}`, phpversion.PHP71, false},
	// 7.2
	{`function f(object $o) {}`, phpversion.PHP72, true},
	{`use A\{B, C,};`, phpversion.PHP72, false},
	// 7.3
	{`f(1, 2,);`, phpversion.PHP73, false},
	{`f(1, 2 /* c */ ,);`, phpversion.PHP73, false},
	{`unset($a, $b,);`, phpversion.PHP73, false},
	{`isset($a, $b,);`, phpversion.PHP73, false},
	{`list($a, &$b) = $c;`, phpversion.PHP73, false},
	{"$x = <<<EOT\n  a\n  EOT;\n", phpversion.PHP73, false},
	{"$x = [<<<EOT\na\nEOT, 1];\n", phpversion.PHP73, false},
	{"$x = [<<<'N'\na\nN, 1];\n", phpversion.PHP73, false},
	{"$x = <<<EOT\n\t$a\n\tEOT;\n", phpversion.PHP73, false},
	// 7.4
	{`$f = fn($x) => $x;`, phpversion.PHP74, false},
	{`$a ??= 1;`, phpversion.PHP74, false},
	{`class C { public int $p; }`, phpversion.PHP74, false},
	{`class C { public A $p; }`, phpversion.PHP74, false},
	{`$a = [...$b];`, phpversion.PHP74, false},
	{`$a = 1_000;`, phpversion.PHP74, false},
	// 8.0
	{`function f(int|string $a) {}`, phpversion.PHP80, false},
	{`function f(int|false $a) {}`, phpversion.PHP80, false},
	{`function f(?A $a = null): int|null {}`, phpversion.PHP80, false},
	{`function f(mixed $a) {}`, phpversion.PHP80, true},
	{`class C { function m(): static {} }`, phpversion.PHP80, false},
	{`$a?->b;`, phpversion.PHP80, false},
	{`$a?->b();`, phpversion.PHP80, false},
	{`echo match ($a) { 1 => 2, default => 3 };`, phpversion.PHP80, false},
	{`f(name: 1);`, phpversion.PHP80, false},
	{`#[A] function f() {}`, phpversion.PHP80, true},
	{`$f = fn() => throw new E();`, phpversion.PHP80, false},
	{`class C { function __construct(public $a) {} }`, phpversion.PHP80, false},
	{`echo $o::class;`, phpversion.PHP80, false},
	{`try {} catch (E) {}`, phpversion.PHP80, false},
	{`function f($a, $b,) {}`, phpversion.PHP80, false},
	{`$f = function () use ($a, $b,) {};`, phpversion.PHP80, false},
	{`class C { function m($a,) {} }`, phpversion.PHP80, false},
	// 8.1
	{`enum E { case A; }`, phpversion.PHP81, false},
	{`class C { public readonly int $p; }`, phpversion.PHP81, false},
	{`class C { function __construct(public readonly int $p) {} }`, phpversion.PHP81, false},
	{`function f(): never {}`, phpversion.PHP81, true},
	{`$f = strlen(...);`, phpversion.PHP81, false},
	{`function f($a = new A()) {}`, phpversion.PHP81, false},
	{`function f(A&B $a) {}`, phpversion.PHP81, false},
	{`$a = 0o17;`, phpversion.PHP81, false},
	{`class C { final const X = 1; }`, phpversion.PHP81, false},
	// 8.2
	{`readonly class C {}`, phpversion.PHP82, false},
	{`function f((A&B)|null $a) {}`, phpversion.PHP82, false},
	{`function f(): null {}`, phpversion.PHP82, false},
	{`function f(false $a) {}`, phpversion.PHP82, false},
	{`function f(int|true $a) {}`, phpversion.PHP82, false},
	{`trait T { const X = 1; }`, phpversion.PHP82, false},
	// 8.3
	{`class C { const int X = 1; }`, phpversion.PHP83, false},
	{`echo C::{$x};`, phpversion.PHP83, false},
	// 8.4
	{`class C { public int $p { get => 1; } }`, phpversion.PHP84, false},
	{`class C { public private(set) int $p; }`, phpversion.PHP84, false},
	{`class C { function __construct(public protected(set) int $p) {} }`, phpversion.PHP84, false},
	{`new A()->m();`, phpversion.PHP84, false},
	{`new A()();`, phpversion.PHP84, false},
	{`new A()->p;`, phpversion.PHP84, false},
	{`new A()[0];`, phpversion.PHP84, false},
	{`new A()::m();`, phpversion.PHP84, false},
	{`new A()::$p;`, phpversion.PHP84, false},
	{`new A()::X;`, phpversion.PHP84, false},
	{`class C { public $p { get => __PROPERTY__; } }`, phpversion.PHP84, false},
	// 8.5
	{`$a = $b |> f(...);`, phpversion.PHP85, false},
	{`(void) f();`, phpversion.PHP85, false},
	{`$b = clone($a, ['p' => 1]);`, phpversion.PHP85, false},
}

func TestRequiredVersion(t *testing.T) {
	for _, tc := range requiredVersionCases {
		src := "<?php " + tc.src
		f := parse(t, src, phpversion.Max)
		if len(f.Errors) != 0 {
			t.Fatalf("%s: %v", tc.src, f.Errors)
		}
		got, at := RequiredVersion(f)
		if got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.src, got, tc.want)
		}
		if (at == nil) != (tc.want == phpversion.Min) {
			t.Errorf("%s: node %T", tc.src, at)
		}
		if at != nil && minVersion(nil, at) != got {
			t.Errorf("%s: MinVersion(%T) disagrees", tc.src, at)
		}
		if tc.semantic {
			continue
		}
		if valid, ok := phpLint(t, tc.want, src); ok && !valid {
			t.Errorf("%s: php %s rejects it", tc.src, tc.want)
		}
		if prev := previousVersion(tc.want); prev != 0 {
			if valid, ok := phpLint(t, prev, src); ok && valid {
				t.Errorf("%s: php %s already accepts it", tc.src, prev)
			}
		}
	}
}

func TestRequiredVersionOnBrokenCode(t *testing.T) {
	for src, want := range map[string]phpversion.Version{
		`<?php function f(): {}`: phpversion.PHP70,
		`<?php f(1`:              phpversion.Min,
	} {
		f := parse(t, src, phpversion.Max)
		if len(f.Errors) == 0 {
			t.Fatalf("%s: expected parse errors", src)
		}
		if got, _ := RequiredVersion(f); got != want {
			t.Errorf("%s: got %s, want %s", src, got, want)
		}
	}
}

func TestMinVersionNil(t *testing.T) {
	if MinVersion(nil) != phpversion.Min {
		t.Fatal("nil")
	}
	// Error recovery always yields a type node; other expressions need nothing.
	if typeVersion(&BadExpr{}) != phpversion.Min {
		t.Fatal("non-type")
	}
}

func BenchmarkRequiredVersion(b *testing.B) {
	var src []byte
	for range 200 {
		src = append(src, `class C { public function m(array $a, $b = 1) { foreach ($a as $k => $v) { if ($v) { return f($k, $b)[0]; } } return [$a, "x$b"]; } }`...)
		src = append(src, '\n')
	}
	f := Parse("x.php", append([]byte("<?php\n"), src...), Options{Version: phpversion.Max})
	b.ReportAllocs()
	for b.Loop() {
		RequiredVersion(f)
	}
}
