package infer_test

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestFirstClassCallableReturns(t *testing.T) {
	check(t, `<?php
namespace App;
use function strlen as countBytes;
function strlen($s): bool { return true; }
function body($x) { return 3; }
function nothing($x): void {}
/** @template T @param T $x @return T */
function identity($x) { return $x; }
/** @return ($x is int ? string : float) */
function conditional($x) {}
class Open { public function body() { return 1; } }
final class Closed {
 public function body() { return 'a'; }
 public static function create(): static { return new static(); }
 public function __invoke(): float { return 1.0; }
 /** @template T @param T $x @return T */
 public function identity($x) { return $x; }
}
/** @template T */
final class Box { /** @return T */ public function get(): mixed {} }
class ParentClass { public static function body() { return 1; } }
final class ChildClass extends ParentClass {
 public static function run() {
  $parent = parent::body(...); t('parent', $parent());
  $self = self::body(...); t('self', $self());
 }
}
/** @param Box<Closed> $box */
function generic($box) { $get = $box->get(...); t('generic', $get()); }
function run($unknown, Open $open) {
 $builtin = \strlen(...); t('builtin', $builtin('a'));
 $local = strlen(...); t('local', $local('a'));
 $import = countBytes(...); t('import', $import('a'));
 $body = body(...); t('body', $body(1));
 $nothing = nothing(...); t('void', $nothing(1));
 $identity = identity(...); t('template', $identity(1));
 $conditional = conditional(...); t('conditional', $conditional(1));
 $missing = missing(...); t('missing', $missing());
 $closed = new Closed();
 $method = $closed->body(...); t('method', $method());
 $virtual = $open->body(...); t('virtual', $virtual());
 $static = Closed::create(...); t('static', $static());
 $mt = $closed->identity(...); t('methodTemplate', $mt(1));
 $invoke = $closed(...); t('invoke', $invoke());
 $nested = $builtin(...); t('nested', $nested('a'));
 $unknownCallable = $unknown(...); t('unknown', $unknownCallable());
 $time = \gettimeofday(...); t('time', $time(true));
 $micro = \microtime(...); t('micro', $micro(true));
 t('map', array_map(\strlen(...), ['a']));
 t('identityClosure', $builtin);
}
`, map[string]string{
		"builtin": "int", "local": "bool", "import": "int", "body": "int", "void": "null",
		"template": "?unknown", "conditional": "float|string", "missing": "?unknown", "method": "string",
		"virtual": "?unknown", "static": `\App\Closed`, "methodTemplate": "?unknown", "invoke": "float",
		"nested": "int", "unknown": "?unknown", "time": "float|int[]", "micro": "float|string",
		"map": "int[]", "identityClosure": `\Closure`, "parent": "int", "self": "int", "generic": `\App\Closed`,
	})
}

func TestFirstClassCrossFileReturns(t *testing.T) {
	checkWith(t, map[string]string{"lib.php": `<?php
namespace Lib;
function value() { return 1; }
final class Source {
 public function value() { return 'a'; }
 public static function create(): static { return new static(); }
}
`}, `<?php
$f = \Lib\value(...); t('function', $f());
$m = (new \Lib\Source())->value(...); t('method', $m());
$s = \Lib\Source::create(...); t('static', $s());
`, map[string]string{"function": "int", "method": "string", "static": `\Lib\Source`})
}

func TestFirstClassDuplicateDeclarations(t *testing.T) {
	check(t, `<?php
if (true) { function duplicate(): int { return 1; } }
else { function duplicate(): string { return ''; } }
$f = duplicate(...); t('union', $f());
`, map[string]string{"union": "int|string"})
	var src strings.Builder
	src.WriteString("<?php\n")
	for range 17 {
		src.WriteString("if (true) { function duplicate(): int { return 1; } }\n")
	}
	src.WriteString("$f = duplicate(...); t('limit', $f());")
	check(t, src.String(), map[string]string{"limit": "?unknown"})
}

func TestFirstClassSourceTreeAndNativeMode(t *testing.T) {
	src := []byte(`<?php
/** @return string */ function documented($x) { return 1; }
final class C {
 /** @return string */ public function documented() { return 1; }
 /** @return string */ public static function staticDocumented() { return 1; }
}
$f = documented(...); $m = (new C())->documented(...); $s = C::staticDocumented(...);
`)
	f := syntax.Parse("t.php", src, syntax.Options{Version: phpversion.PHP84})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	parents := map[syntax.Node]syntax.Node{}
	spans := map[syntax.Node]syntax.Span{}
	args := map[syntax.Node]*syntax.ArgList{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		parents[n], spans[n] = n.Parent(), n.Span()
		switch n := n.(type) {
		case *syntax.FuncCall:
			args[n] = n.Args
		case *syntax.MethodCall:
			args[n] = n.Args
		case *syntax.StaticCall:
			args[n] = n.Args
		}
		return true
	})
	for _, native := range []bool{false, true} {
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
		if native {
			env = env.Native()
		}
		for n, original := range args {
			for range 2 {
				expr := n.(syntax.Expr)
				ret := env.TypeOf(expr).DocString()
				if native && strings.Contains(ret, "string") || !native && !strings.Contains(ret, "string") {
					t.Errorf("native=%v: %s", native, ret)
				}
			}
			switch n := n.(type) {
			case *syntax.FuncCall:
				if n.Args != original {
					t.Fatal("function arguments changed")
				}
			case *syntax.MethodCall:
				if n.Args != original {
					t.Fatal("method arguments changed")
				}
			case *syntax.StaticCall:
				if n.Args != original {
					t.Fatal("static arguments changed")
				}
			}
		}
	}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if n.Parent() != parents[n] || n.Span() != spans[n] {
			t.Fatal("source tree changed")
		}
		return true
	})
}

func BenchmarkFirstClassCallableReturns(b *testing.B) {
	src := []byte(`<?php final class C { public function value(): string { return ''; } public static function make(): static { return new static(); } } $c = new C(); $f = strlen(...); $m = $c->value(...); $s = C::make(...);`)
	f := syntax.Parse("t.php", src, syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	ns := names.New(f)
	var calls []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.FuncCall:
			calls = append(calls, n)
		case *syntax.MethodCall:
			calls = append(calls, n)
		case *syntax.StaticCall:
			calls = append(calls, n)
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, ns, ix, phpversion.PHP84)
		for _, call := range calls {
			env.TypeOf(call)
		}
	}
}
