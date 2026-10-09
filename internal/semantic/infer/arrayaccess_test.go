package infer

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
)

func arrayAccessEnv(t testing.TB, src string, others ...string) (*Env, map[string]*syntax.ArrayDimFetch) {
	t.Helper()
	f := syntax.Parse("reads.php", []byte(src), syntax.Options{Version: phpversion.PHP85})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	for _, other := range others {
		lib := syntax.Parse("lib.php", []byte(other), syntax.Options{Version: phpversion.PHP85})
		if len(lib.Errors) > 0 {
			t.Fatalf("library parse: %v", lib.Errors)
		}
		ix.Add(index.Extract(lib))
	}
	ix.Add(index.Extract(f))
	e := NewEnv(f, names.New(f), ix, phpversion.PHP85)
	reads := map[string]*syntax.ArrayDimFetch{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		call, ok := n.(*syntax.FuncCall)
		if !ok || len(call.Args.Args) != 2 {
			return true
		}
		arg := call.Args.Args[0].(*syntax.Arg).Value
		if dim, ok := call.Args.Args[1].(*syntax.Arg).Value.(*syntax.ArrayDimFetch); ok {
			label := strings.Trim(string(f.Src[arg.Span().Start:arg.Span().End]), "'")
			reads[label] = dim
		}
		return true
	})
	return e, reads
}

func assertArrayAccessReads(t *testing.T, e *Env, reads map[string]*syntax.ArrayDimFetch, want map[string]string) {
	t.Helper()
	for label, expected := range want {
		n := reads[label]
		if n == nil {
			t.Fatalf("missing read %s", label)
		}
		if got := e.TypeOf(n).String(); got != expected {
			t.Errorf("%s: got %s, want %s", label, got, expected)
		}
	}
}

func TestArrayAccessReads(t *testing.T) {
	e, reads := arrayAccessEnv(t, `<?php
interface IntOffsets extends ArrayAccess { public function offsetGet($offset): int; }
class IntBox implements IntOffsets { public function offsetGet($offset): int { return 1; } }
class ChildBox extends IntBox {}
class StringBox implements ArrayAccess { public function offsetGet($offset): string { return ''; } }
class Pretender { public function offsetGet($offset): int { return 1; } }
function reads(IntBox $a, ChildBox $child, IntOffsets $iface, StringBox $b,
    IntBox|StringBox $union, ?IntBox $nullable, IntBox|false $failed,
    IntBox|Pretender $bad, IntBox|int $scalar, IntBox|array $arrayUnion,
    Pretender $fake, ArrayAccess $unknown, array $array, string $string,
    IntBox&Countable $intersection, Pretender&Countable $unsupported) {
    t('own', $a[0]); t('inherited', $child[0]); t('interface', $iface[0]);
    t('union', $union[0]); t('nullable', $nullable[0]); t('failed', $failed[0]);
    t('bad', $bad[0]); t('scalar', $scalar[0]); t('arrayUnion', $arrayUnion[0]);
    t('fake', $fake[0]); t('unknown', $unknown[0]); t('array', $array[0]);
    t('string', $string[0]); t('intersection', $intersection[0]);
    t('unsupported', $unsupported[0]);
}
`)
	assertArrayAccessReads(t, e, reads, map[string]string{
		"own": "int", "inherited": "int", "interface": "int", "union": "int|string",
		"nullable": "int", "failed": "int", "bad": "?unknown", "scalar": "?unknown",
		"arrayUnion": "?unknown", "fake": "?unknown", "unknown": "mixed",
		"array": "?unknown", "string": "string", "intersection": "int", "unsupported": "?unknown",
	})
	if got := e.arrayAccessDimType(&syntax.ArrayDimFetch{}, types.Of(`\IntBox`)); !got.IsUnknown() {
		t.Errorf("append read: %s", got)
	}
	if got := e.arrayAccessDimType(reads["own"], types.Unknown); !got.IsUnknown() {
		t.Errorf("unknown receiver: %s", got)
	}
}

func TestArrayAccessGenericReadsAcrossFiles(t *testing.T) {
	e, reads := arrayAccessEnv(t, `<?php
/**
 * @param Box<string> $box
 * @param Child<int> $child
 * @param Offsets<float> $iface
 */
function reads(Box $box, Child $child, Offsets $iface, Fixed $fixed) {
    t('box', $box[0]); t('child', $child[0]); t('iface', $iface[0]); t('fixed', $fixed[0]);
}
`, `<?php
/** @template T */
class Box implements ArrayAccess {
    /** @return T */
    public function offsetGet($offset): mixed {}
}
/**
 * @template V
 * @extends Box<V>
 */
class Child extends Box {}
/** @template T */
interface Offsets extends ArrayAccess {
    /** @return T */
    public function offsetGet($offset): mixed;
}
/** @extends Box<bool> */
class Fixed extends Box {}
`)
	assertArrayAccessReads(t, e, reads, map[string]string{
		"box": "string", "child": "int", "iface": "float", "fixed": "bool",
	})
	assertArrayAccessReads(t, e.Native(), reads, map[string]string{
		"box": "mixed", "child": "mixed", "iface": "mixed", "fixed": "mixed",
	})
}

func TestArrayAccessConditionalOffset(t *testing.T) {
	e, reads := arrayAccessEnv(t, `<?php
class ConditionalBox implements ArrayAccess {
    /** @return ($offset is int ? string : bool) */
    public function offsetGet($offset): mixed {}
}
class TemplateBox implements ArrayAccess {
    /**
     * @template T
     * @param T $offset
     * @return T
     */
    public function offsetGet($offset): mixed {}
}
function reads(ConditionalBox $box, TemplateBox $template) {
    t('int', $box[1]); t('string', $box['key']);
    t('templateInt', $template[1]); t('templateString', $template['key']);
}
`)
	assertArrayAccessReads(t, e, reads, map[string]string{
		"int": "string", "string": "bool", "templateInt": "int", "templateString": "string",
	})
}

func TestArrayAccessMissingAndRecursiveContracts(t *testing.T) {
	e, reads := arrayAccessEnv(t, `<?php
interface ArrayAccess {}
class Missing implements ArrayAccess {}
final class Recursive implements ArrayAccess {
    public function offsetGet($offset) { return $this[$offset]; }
}
function reads(Missing $missing, Recursive $recursive) {
    t('missing', $missing[0]); t('recursive', $recursive[0]);
}
`)
	assertArrayAccessReads(t, e, reads, map[string]string{
		"missing": "?unknown", "recursive": "?unknown",
	})
}

func BenchmarkArrayAccessDim(b *testing.B) {
	e, reads := arrayAccessEnv(b, `<?php
/** @template T */
class Box implements ArrayAccess {
    /** @return T */
    public function offsetGet($offset): mixed {}
}
/** @param Box<string> $box */
function reads(Box $box) { t('value', $box[0]); }
`)
	n := reads["value"]
	receiver := e.TypeOf(n.Var)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.arrayAccessDimType(n, receiver)
	}
}
