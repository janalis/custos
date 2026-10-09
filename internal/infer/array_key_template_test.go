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
 t('value', valueOf(['x'=>1]));
 t('unionValue', unionValueKey($strings));
 t('nonEmpty', nonEmptyKey(['x'=>1]));
 t('copy', copyKeys($strings));
 t('copyMixed', mixedCopy($strings));
 t('class', $box->values());
 foreach(copyKeys($strings) as $k => $v) { t('copyKey', $k); }
 foreach(mixedCopy($strings) as $k => $v) { t('mixedCopyKey', $k); }
 foreach($box->values() as $k => $v) { t('classKey', $k); }
 $strings[3] = 1;
 t('written', keyOf($strings));
}
/** @param array<string,int> $strings */
function aliasing($strings) {
 $ref =& $strings;
 t('alias', keyOf($strings));
}
`, map[string]string{
		"string": "string", "integer": "int", "literalString": "string", "literalInteger": "int", "both": "int|string", "named": "string", "mixed": "string", "combined": "int|string", "empty": "mixed", "unknown": "mixed", "unpack": "mixed", "compound": "mixed", "value": "int", "unionValue": "string", "nonEmpty": "string", "copy": "int[]", "copyMixed": "mixed[]", "class": "int[]", "copyKey": "string", "mixedCopyKey": "string", "classKey": "string", "written": "int|string", "alias": "mixed",
	})
}

func BenchmarkArrayKeyTemplateBinding(b *testing.B) {
	f := syntax.Parse("keys.php", []byte(arrayKeyTemplateLib+`function run() { copyKeys(['x'=>1]); }`), syntax.Options{Version: phpver.PHP85})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	nm := names.New(f)
	var call *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name, ok := c.Name.(*syntax.Name); ok && name.Value == "copyKeys" {
				call = c
			}
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		_ = infer.NewEnv(f, nm, ix, phpver.PHP85).TypeOf(call)
	}
}
