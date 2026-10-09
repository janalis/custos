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

func TestGenericArrayKeys(t *testing.T) {
	check(t, `<?php
/** @param array<string,int> $xs */
function strings($xs) {
 foreach ($xs as $k => $v) { t('strings', $k); }
 foreach (array_filter($xs) as $k => $v) { t('filter', $k); }
 foreach (array_map('intval', $xs) as $k => $v) { t('map', $k); }
 foreach (array_map('intval', $xs, $xs) as $k => $v) { t('multi', $k); }
 foreach (array_values($xs) as $k => $v) { t('values', $k); }
 foreach (array_unique($xs) as $k => $v) { t('unique', $k); }
 foreach (array_slice($xs, 1) as $k => $v) { t('slice', $k); }
 foreach (array_reverse($xs) as $k => $v) { t('reverse', $k); }
 foreach (($xs + $xs) as $k => $v) { t('union', $k); }
 $xs[4] = 3;
 foreach ($xs as $k => $v) { t('integerWrite', $k); }
 $xs[] = 3;
 foreach ($xs as $k => $v) { t('append', $k); }
 $copy = $xs;
 foreach ($copy as $k => $v) { t('copy', $k); }
}
/** @param list<int> $xs */
function integers($xs) {
 $empty = []; foreach ($empty as $k => $v) { t('empty', $k); }
 foreach ($xs as $k => $v) { t('list', $k); }
 $xs['a'] = 2;
 foreach ($xs as $k => $v) { t('stringWrite', $k); }
}
/** @param array<string,int> $xs */
function dynamic($xs, $key) {
 $xs[$key] = 1;
 foreach ($xs as $k => $v) { t('dynamic', $k); }
}
/** @param array<string,array<int>> $xs */
function nested($xs) {
 $xs['a'][] = 2;
 foreach ($xs as $k => $v) { t('nested', $k); }
}
/** @param array<string,int> $xs */
function aliasing($xs) {
 $ref =& $xs;
 foreach ($xs as $k => $v) { t('alias', $k); }
 $capture = function() use ($xs) { foreach ($xs as $k => $v) { t('captureAlias', $k); } };
 $capture();
}
/** @param array<string,int> $xs */
function generator($xs) { yield from $xs; }
foreach (generator([]) as $k => $v) { t('yieldKey', $k); t('yieldValue', $v); }
/** @template K
 * @template V
 * @param iterable<K,V> $xs
 * @return K
 */
function firstKey($xs) { foreach ($xs as $k => $v) return $k; }
/** @param array<string,int> $xs */
function template($xs) { t('template', firstKey($xs)); }
`, map[string]string{
		"integerWrite": "int|string", "empty": "int|string", "strings": "string", "filter": "string", "map": "string", "multi": "int", "values": "int",
		"unique": "string", "slice": "int|string", "reverse": "int|string",
		"union": "string", "append": "int|string", "copy": "int|string", "list": "int", "stringWrite": "int|string",
		"dynamic": "int|string", "nested": "string", "alias": "int|string", "captureAlias": "int|string", "yieldKey": "string", "yieldValue": "int", "template": "string",
	})
}

func BenchmarkGenericArrayKeys(b *testing.B) {
	f := syntax.Parse("keys.php", []byte(`<?php /** @param array<string,int> $xs */ function f($xs) { $both = $xs + $xs; foreach ($xs as $k => $v) { strlen($k); } $xs[] = 1; foreach ($xs as $k => $v) { strlen($k); } }`), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	nm := names.New(f)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, nm, ix, phpver.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if expr, ok := n.(syntax.Expr); ok {
				_ = env.TypeOf(expr)
			}
			return true
		})
	}
}

func TestGenericArrayKeysCrossFile(t *testing.T) {
	f := syntax.Parse("key-source.php", []byte(`<?php /** @return array<string,int> */ function keyed() { return []; }`), syntax.Options{Version: phpver.PHP84})
	consumer := syntax.Parse("key-user.php", []byte(`<?php foreach (keyed() as $key => $value) { echo $key; }`), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	ix.Add(index.Extract(consumer))
	env := infer.NewEnv(consumer, names.New(consumer), ix, phpver.PHP84)
	syntax.InspectFile(consumer, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.Name == "key" {
			if got := env.TypeOf(v).String(); got != "string" {
				t.Errorf("cross-file key type: %s", got)
			}
		}
		return true
	})
}
