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
	f := syntax.Parse("keys.php", []byte(`<?php /** @param array<string,int> $xs */ function f($xs) { $both = $xs + $xs; foreach ($xs as $k => $v) { strlen($k); } $xs[] = 1; foreach ($xs as $k => $v) { strlen($k); } }`), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	nm := names.New(f)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, nm, ix, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if expr, ok := n.(syntax.Expr); ok {
				_ = env.TypeOf(expr)
			}
			return true
		})
	}
}

func TestGenericArrayKeysCrossFile(t *testing.T) {
	f := syntax.Parse("key-source.php", []byte(`<?php /** @return array<string,int> */ function keyed() { return []; }`), syntax.Options{Version: phpversion.PHP84})
	consumer := syntax.Parse("key-user.php", []byte(`<?php foreach (keyed() as $key => $value) { echo $key; }`), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	ix.Add(index.Extract(consumer))
	env := infer.NewEnv(consumer, names.New(consumer), ix, phpversion.PHP84)
	syntax.InspectFile(consumer, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.Name == "key" {
			if got := env.TypeOf(v).String(); got != "string" {
				t.Errorf("cross-file key type: %s", got)
			}
		}
		return true
	})
}

// Keys absent from a literal get no type from the other elements, also
// after writes to other keys or nested writes.
func TestAbsentKeysAfterWrites(t *testing.T) {
	checkAnywhere(t, `<?php
function f(int $x, string $k) {
    $rows = ['total' => $x];
    $rows['meta']['s'] = 1;
    t('nested', $rows['meta']);
    $w = ['a' => 1];
    $w['b'] = 'x';
    t('otherKey', $w['c']);
    t('written', $w['b']);
    $l = [1, 2];
    $l[] = 3;
    t('append', $l[7]);
    $l2 = ['a' => 1];
    $l2[] = 3;
    t('appendStr', $l2['z']);
    $c = ['a' => 1];
    $c[$k] = 'v';
    t('computed', $c['q']);
    $d = ['a' => 1];
    [$d['b']] = [2];
    t('destructured', $d['b']);
}
`, map[string]string{
		"nested": "?unknown", "otherKey": "?unknown", "written": "string", "append": "?unknown",
		"appendStr": "?unknown", "computed": "string", "destructured": "?unknown",
	})
}

func TestComputedArrayKeysAndCoalescingAssignments(t *testing.T) {
	checkAnywhere(t, `<?php
function f($id, array $rows) {
    $r = [];
    $r[$id] ??= [];
    t('coal', $r[$id]);
    $q = [];
    t('expr', $q[$id] ??= []);
    $w = ['a' => 1];
    t('expr2', $w[$id] ??= 'x');
    $by = [];
    foreach ($rows as $row) { $by[$row['k']] ??= []; $by[$row['k']][] = $row; t('nested', $by[$row['k']]); }
    $lit = ['a' => 1, 'b' => 'x'];
    t('computed', $lit[$id]);
    $u = [];
    t('emptyComputed', $u[$id]);
    $z = rand() ? ['a' => 1] : 'x';
    t('mixedBase', $z[$id]);
    t('literalBase', ['a' => 1, 'b' => 2][$id]);
    $x = [];
    [$x[$id]] = [1];
    t('unknownWrite', $x[$id]);
    $obj = (object) [];
    t('objDim', $obj->{'a'} ??= 1);
    t('propCoalesce', $obj->list[$id] ??= 1);
}
`, map[string]string{
		"coal":          "array{}",
		"expr":          "array{}",
		"expr2":         "int|string",
		"nested":        "array",
		"computed":      "?unknown",
		"emptyComputed": "?unknown",
		"mixedBase":     "?unknown",
		"literalBase":   "int",
		"unknownWrite":  "?unknown",
		"objDim":        "?unknown",
		"propCoalesce":  "?unknown",
	})
}
