package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

const arrayKeyTemplateLib = `<?php
/** @template K
 * @param array<K,int> $xs
 * @return K
 */
function keyOf($xs) { return array_key_first($xs); }
/** @template K
 * @template V
 * @param array<K,V> $xs
 * @return array<K,V>
 */
function copyKeys($xs): array { return $xs; }
/** @template K
 * @param array<K,mixed> $xs
 * @return K
 */
function mixedKey($xs) {}
/** @template K
 * @param array<K,int> $a
 * @param array<K,int> $b
 * @return K
 */
function combinedKey($a, $b) {}
/** @template K
 * @param array<K|int,int> $xs
 * @return K
 */
function compoundKey($xs) {}
/** @template K
 * @param array<K|string,int> $xs
 * @return K
 */
function stringCompoundKey($xs) {}
/** @template K
 * @param array<K|array-key,int> $xs
 * @return K
 */
function coveredKey($xs) {}
/** @template K
 * @template J
 * @param array<K|J,int> $xs
 * @return K
 */
function ambiguousKey($xs) {}
/** @template K
 * @template V
 * @param array<K|int,V> $xs
 * @return array<K|int,V>
 */
function compoundCopy($xs): array { return $xs; }
/** @template K
 * @template V
 * @param array<K,V> $xs
 * @return V
 */
function valueOf($xs) {}
/** @template K
 * @param array<K,int|string> $xs
 * @return K
 */
function unionValueKey($xs) {}
/** @template K
 * @param non-empty-array<K,int> $xs
 * @return K
 */
function nonEmptyKey($xs) {}
/** @template K
 * @param array<K,int> $xs
 * @return array<K,mixed>
 */
function mixedCopy($xs): array { return $xs; }
/** @template K
 * @template V
 */
class Box {
 /** @return array<K,V> */
 public function values(): array { return []; }
 /** @return array<K|int,V> */
 public function compoundValues(): array { return []; }
 /** @template T
  * @param array<T|int,int> $xs
  * @return T
  */
 public function key($xs) {}
}
`

func TestArrayKeyTemplateBinding(t *testing.T) {
	checkWith(t, map[string]string{"lib.php": arrayKeyTemplateLib}, `<?php
/** @param array<string,int> $strings
 * @param list<int> $integers
 * @param array $unknown
 * @param Box<string,int> $box
 */
function run($strings, $integers, $unknown, $box) {
 t('string', keyOf($strings));
 t('integer', keyOf($integers));
 t('literalString', keyOf(['x'=>1]));
 t('literalInteger', keyOf([1]));
 t('both', keyOf(['x'=>1, 1]));
 t('named', keyOf(xs: $strings));
 t('mixed', mixedKey(['x'=>1]));
 t('combined', combinedKey(['x'=>1], [1]));
 t('empty', keyOf([]));
 t('unknown', keyOf($unknown));
 t('unpack', keyOf(...$unknown));
 t('compound', compoundKey($strings));
 t('compoundBoth', compoundKey(['x'=>1, 1]));
 t('compoundCovered', compoundKey($integers));
 t('compoundUnknown', compoundKey($unknown));
 t('compoundEmpty', compoundKey([]));
 t('compoundNamed', compoundKey(xs: $strings));
 t('stringCompound', stringCompoundKey($integers));
 t('stringCompoundBoth', stringCompoundKey(['x'=>1, 1]));
 t('stringCompoundCovered', stringCompoundKey($strings));
 t('covered', coveredKey(['x'=>1, 1]));
 t('ambiguous', ambiguousKey($strings));
 t('compoundCopy', compoundCopy($strings));
 t('compoundMethod', $box->key($strings));
 t('value', valueOf(['x'=>1]));
 t('unionValue', unionValueKey($strings));
 t('nonEmpty', nonEmptyKey(['x'=>1]));
 t('copy', copyKeys($strings));
 t('copyMixed', mixedCopy($strings));
 t('class', $box->values());
 foreach(copyKeys($strings) as $k => $v) { t('copyKey', $k); }
 foreach(mixedCopy($strings) as $k => $v) { t('mixedCopyKey', $k); }
 foreach($box->values() as $k => $v) { t('classKey', $k); }
 foreach(compoundCopy($strings) as $k => $v) { t('compoundCopyKey', $k); t('compoundCopyValue', $v); }
 foreach($box->compoundValues() as $k => $v) { t('compoundClassKey', $k); }
 $strings[3] = 1;
 t('written', keyOf($strings));
}
/** @param array<string,int> $strings */
function aliasing($strings) {
 $ref =& $strings;
 t('alias', keyOf($strings));
}
`, map[string]string{
		"string": "string", "integer": "int", "literalString": "string", "literalInteger": "int", "both": "int|string", "named": "string", "mixed": "string", "combined": "int|string", "empty": "mixed", "unknown": "mixed", "unpack": "mixed", "compound": "string", "value": "int", "unionValue": "string", "nonEmpty": "string", "copy": "int[]", "copyMixed": "mixed[]", "class": "int[]", "copyKey": "string", "mixedCopyKey": "string", "classKey": "string", "written": "int|string", "alias": "mixed",
		"compoundBoth": "string", "compoundCovered": "mixed", "compoundUnknown": "mixed", "compoundEmpty": "mixed", "compoundNamed": "string",
		"stringCompound": "int", "stringCompoundBoth": "int", "stringCompoundCovered": "mixed", "covered": "mixed", "ambiguous": "mixed",
		"compoundCopy": "int[]", "compoundMethod": "string", "compoundCopyKey": "int|string", "compoundCopyValue": "int", "compoundClassKey": "int|string",
	})
}

func BenchmarkArrayKeyTemplateBinding(b *testing.B) {
	benchmarkArrayKeyTemplateBinding(b, "copyKeys")
}

func BenchmarkCompoundArrayKeyTemplateBinding(b *testing.B) {
	benchmarkArrayKeyTemplateBinding(b, "compoundCopy")
}

func benchmarkArrayKeyTemplateBinding(b *testing.B, name string) {
	f := syntax.Parse("keys.php", []byte(arrayKeyTemplateLib+`function run() { `+name+`(['x'=>1]); }`), syntax.Options{Version: phpversion.PHP85})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	nm := names.New(f)
	var call *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if callee, ok := c.Name.(*syntax.Name); ok && callee.Value == name {
				call = c
			}
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		_ = infer.NewEnv(f, nm, ix, phpversion.PHP85).TypeOf(call)
	}
}
