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

func TestPipeInference(t *testing.T) {
	checkVer(t, phpver.PHP85, false, `<?php
namespace App;
function strlen($x): bool { return true; }
/**
 * @template T
 * @param T $x
 * @return T
 */
function identity($x) { return $x; }
final class Mapper {
 /**
 * @template T
 * @param T $x
 * @return T
 */
 public function identity($x) { return $x; }
 public static function count($x): int { return 1; }
 public function __invoke($x): string { return ''; }
}
function nothing($x): void {}
/** @param callable(int): string $cb */
function run($cb, $unknown) {
 $m = new Mapper();
 t('builtin', 'abc' |> \strlen(...));
 t('namespace', 'abc' |> strlen(...));
 t('template', 3 |> identity(...));
 t('method', 'abc' |> $m->identity(...));
 t('static', 'abc' |> Mapper::count(...));
 t('chain', 'abc' |> \strtoupper(...) |> \strlen(...));
 t('closure', 3 |> (fn(int $x): float => $x * 1.5));
 t('signature', 3 |> $cb);
 t('invoke', 3 |> $m);
 t('literal', 'abc' |> 'strlen');
 t('unknown', 3 |> $unknown);
 t('void', 3 |> nothing(...));
 t('factory', 3 |> (fn() => fn(): int => 1)());
 t('methodFactory', 3 |> $m->identity(fn(): int => 1));
 t('staticValue', 3 |> Mapper::count(1));
 t('inputOverride', true |> \gettimeofday(...));
}
`, map[string]string{
		"builtin": "int", "namespace": "bool", "template": "int", "method": "string", "static": "int", "chain": "int", "closure": "float", "signature": "string", "invoke": "string", "literal": "int", "unknown": "?unknown", "void": "null", "factory": "int", "methodFactory": "int", "staticValue": "?unknown", "inputOverride": "float",
	})
}

func TestGeneratorInference(t *testing.T) {
	check(t, `<?php
function values() { yield 1; yield 2; return 'done'; }
function keyed() { yield 'x' => 'a'; yield 'y' => 'b'; }
function bare() { yield; }
function unknown($x) { yield $x; }
function unknownKey($k) { yield $k => 1; }
function delegated() { yield from ['x' => 1, 'y' => 2]; }
function delegatedGenerator() { yield from values(); }
function mixedDelegation($x) { yield 1; yield from $x; }
function dead() { return 1; yield 2; }
function nested() { $f = function () { yield 1; }; return 3; }
function recursive() { yield from recursive(); }
final class Source { public function items() { yield 'x' => 1; } }
function run() {
 t('identity', values()); t('dead', dead()); t('nested', nested()); t('recursive', recursive());
 foreach(values() as $k => $v) { t('key', $k); t('value', $v); }
 foreach(keyed() as $k => $v) { t('stringKey', $k); t('stringValue', $v); }
 foreach(bare() as $v) { t('bare', $v); }
 foreach(unknown(1) as $k => $v) { t('unknownKeyKnown', $k); t('unknownValue', $v); }
 foreach(unknownKey(1) as $k => $v) { t('unknownKey', $k); t('knownValue', $v); }
 foreach(delegated() as $k => $v) { t('delegatedKey', $k); t('delegatedValue', $v); }
 foreach(delegatedGenerator() as $v) { t('delegatedGenerator', $v); }
 foreach(mixedDelegation(1) as $v) { t('mixedDelegation', $v); }
 $closure = function () { yield 1; };
 $arrow = fn() => yield 'a';
 foreach($closure() as $v) { t('closure', $v); }
 foreach($arrow() as $v) { t('arrow', $v); }
 foreach((new Source())->items() as $v) { t('method', $v); }
}
`, map[string]string{
		"identity": `\Generator`, "dead": `\Generator`, "nested": "int", "recursive": `\Generator`, "key": "int", "value": "int", "stringKey": "string", "stringValue": "string", "bare": "null", "unknownKeyKnown": "int", "unknownValue": "?unknown", "unknownKey": "int|string", "knownValue": "int", "delegatedKey": "string", "delegatedValue": "int", "delegatedGenerator": "int", "mixedDelegation": "?unknown", "closure": "int", "arrow": "string", "method": "int",
	})
}

func TestGeneratorCrossFile(t *testing.T) {
	checkWith(t, map[string]string{"lib.php": `<?php function items() { yield 'name' => 1; }`}, `<?php
foreach(items() as $k => $v) { t('key', $k); t('value', $v); }
`, map[string]string{"key": "string", "value": "int"})
}

func TestGeneratorDelegationAndBoundaries(t *testing.T) {
	check(t, `<?php
/** @param int[] $xs */
function arrays($xs) { yield from $xs; }
function integerKeys() { yield from [1, 2]; }
function emptyDelegation() { yield from []; yield 1; }
function emptyGenerator() { yield from []; }
function nesting() {
 $arrow = fn() => yield 1;
 $obj = new class { public function values() { yield 1; } };
 return 'a';
}
/** @return \Generator<string, float> */
function documented() { yield 1; }
function declared(): \Generator { yield 1; }
function run() {
 foreach(arrays([]) as $v) { t('arrayValue', $v); }
 foreach(integerKeys() as $k => $v) { t('key', $k); t('value', $v); }
 foreach(documented() as $k => $v) { t('documentedKey', $k); t('documentedValue', $v); }
 t('nested', nesting()); t('declared', declared()); t('empty', emptyGenerator());
 foreach(emptyDelegation() as $v) { t('emptyThenValue', $v); }
}
`, map[string]string{"arrayValue": "int", "key": "int", "value": "int", "documentedKey": "string", "documentedValue": "float", "nested": "string", "declared": `\Generator`, "empty": `\Generator`, "emptyThenValue": "int"})
	checkVer(t, phpver.PHP55, false, `<?php function old() { yield 1; } t('old', old());`, map[string]string{"old": `\Generator<int, int, mixed, null>`})
}

func TestPipePreservesSourceTree(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php function f() { return 'abc' |> strlen(...); }`), syntax.Options{Version: phpver.PHP85})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP85)
	parents := map[syntax.Node]syntax.Node{}
	spans := map[syntax.Node]syntax.Span{}
	var pipe *syntax.Binary
	var callable *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		parents[n], spans[n] = n.Parent(), n.Span()
		if n, ok := n.(*syntax.Binary); ok {
			pipe = n
		}
		if n, ok := n.(*syntax.FuncCall); ok {
			callable = n
		}
		return true
	})
	before := env.TypeOf(callable)
	args := callable.Args
	for range 2 {
		if got := env.TypeOf(pipe).String(); got != "int" {
			t.Fatalf("pipe: %s", got)
		}
	}
	if callable.Args != args || !before.Equal(env.TypeOf(callable)) {
		t.Fatal("callable changed after pipe inference")
	}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if n.Parent() != parents[n] || n.Span() != spans[n] {
			t.Fatal("source tree changed")
		}
		return true
	})
}
