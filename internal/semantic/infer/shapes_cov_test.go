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

// shapesAt types the second argument of every `t('label', expr)` call at
// PHP version ver (parse errors tolerated: permissive old syntax), with
// array facts.
func shapesAt(t *testing.T, ver phpversion.Version, src string) map[string]string {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: ver})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, ver)
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" && c.Args != nil && len(c.Args.Args) == 2 {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).ShapeString()
			}
		}
		return true
	})
	return got
}

func expectShapes(t *testing.T, got, want map[string]string) {
	t.Helper()
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %q want %q", k, got[k], w)
		}
	}
}

func TestShapesKeysAndWrites(t *testing.T) {
	big := "[" + strings.Repeat("1, ", 40) + "]"
	got := shapesAt(t, phpversion.PHP84, `<?php
function keys($k) {
    t('overflow', [9999999999999999999 => 1]);
    t('binary', [b'k' => 1]);
    t('nowdoc', [<<<'EOT'
k
EOT => 1]);
    t('dq', ["k" => 1]);
    t('big', `+big+`);
    $a = ['x' => 1];
    t('dyn', $a[$k]);
    $b = ['x' => 1];
    $v = 's';
    $b['x'] = &$v;
    t('byref', $b['x']);
    foreach ([] as $e) { t('emptyLit', $e); }
    $c = [];
    $c['x']['y'] = 1;
    foreach ($c as $e) { t('emptyNested', $e); }
    $d = ['x' => 1];
    $d['y']['z'] = 2;
    foreach ($d as $kk => $e) { t('keyNested', $kk); }
    $g = ['x' => 1];
    $g['y'] = 2;
    foreach ($g as $kk => $e) { t('keyLit', $kk); }
    ['q' => $miss] = ['x' => 1];
    t('target', $miss);
}
`)
	expectShapes(t, got, map[string]string{
		"overflow":    "non-empty int[]", // out of int64 range: no literal key, no shape
		"binary":      "int[]{k: int}",
		"nowdoc":      "non-empty int[]", // heredoc/nowdoc keys are not plain strings
		"dq":          "int[]{k: int}",
		"big":         "non-empty int[]", // more items than MaxShapeKeys
		"dyn":         "int",             // computed key: element type
		"byref":       "?unknown",        // a reference write into the key
		"emptyLit":    "?unknown",
		"emptyNested": "?unknown", // nested write into an empty literal
		"keyNested":   "string",
		"keyLit":      "string",
		"target":      "?unknown", // destructured key absent from the shape
	})
}

func TestShapesMutations(t *testing.T) {
	got := shapesAt(t, phpversion.PHP84, `<?php
class Box {
    function __construct(&$a) {}
    function m(&$a) {}
    static function s(&$a) {}
}
function byRef(&$a, $b = 0) {}
function byRefVariadic($a, &...$rest) {}
function plain($a) {}
function muts($o, $m, $other) {
    $l = ['k' => 1];
    list(&$l) = [[1]];
    t('listRef', $l);
    $l2 = ['k' => 1];
    [$z, &$l2] = [1, [2]];
    t('arrRef', $l2);
    $u = ['k' => 1];
    $f = function () use (&$u) {};
    t('useRef', $u);
    $gl = ['k' => 1];
    global $gl;
    t('global', $gl);
    $st = ['k' => 1];
    static $st;
    t('static', $st);
    $mc = ['k' => 1];
    $o->$m($mc);
    t('dynMethod', $mc);
    $sc = ['k' => 1];
    Box::$m($sc);
    t('dynStatic', $sc);
    $nw = ['k' => 1];
    new Box($nw);
    t('new', $nw);
    $nwu = ['k' => 1];
    new $other($nwu);
    t('newDyn', $nwu);
    $na = ['k' => 1];
    byRef(a: $na);
    t('named', $na);
    $nb = ['k' => 1];
    byRef(1, zz: $nb);
    t('namedMissing', $nb);
    $va = ['k' => 1];
    byRefVariadic(1, 2, $va);
    t('variadic', $va);
    $nv = ['k' => 1];
    plain(1, $nv);
    t('extra', $nv);
}
function branches($c, $d) {
    $x = [1];
    if ($c) { array_pop($x); } elseif ($d) { t('elseif', $x); } else { array_pop($x); }
}
$arrow = fn() => [$a = ['k' => 1], sort($a), t('arrow', $a)];
`)
	expectShapes(t, got, map[string]string{
		"listRef":      "?unknown",
		"arrRef":       "?unknown",
		"useRef":       "?unknown",
		"global":       "?unknown",
		"static":       "?unknown",
		"dynMethod":    "int[]{k: int}", // unresolvable callee: not clobbered
		"dynStatic":    "int[]{k: int}",
		"newDyn":       "int[]{k: int}",
		"new":          "int[]",           // by-reference constructor parameter
		"named":        "int[]",           // named by-reference parameter
		"namedMissing": "int[]{k: int}",   // no parameter of that name
		"variadic":     "int[]",           // by-reference variadic
		"extra":        "int[]{k: int}",   // beyond the parameters, not variadic
		"elseif":       "non-empty int[]", // mutations only in exclusive branches
		"arrow":        "int[]",           // sort() inside the arrow function
	})
}

func TestShapesVersionAndCaps(t *testing.T) {
	many := strings.Repeat("    $w['x'] = 2;\n", 600)
	got := shapesAt(t, phpversion.PHP82, `<?php
function plain($a) {}
function f() {
    t('neg82', [-5 => 'a', 'b']);
    $e = [];
    $e['x'] = throw new Exception();
    foreach ($e as $v) { t('never', $v); }
    $x = ['k' => 1];
    plain(&$x);
    t('callRef', $x);
    $w = ['x' => 1];
`+many+`    t('manyKey', $w['x']);
    foreach ($w as $k => $v) { t('manyForeachKey', $k); t('manyForeachVal', $v); }
}
`)
	got2 := shapesAt(t, phpversion.PHP84, `<?php t('neg84', [-5 => 'a', 'b']);`)
	expectShapes(t, got, map[string]string{
		"neg82":          "string[]{-5: string, 0: string}",
		"never":          "?unknown", // only never-typed writes into []
		"callRef":        "int[]",    // call-time pass-by-reference (PHP < 5.4)
		"manyKey":        "?unknown", // past maxVarDefs element writes
		"manyForeachKey": "int|string",
		"manyForeachVal": "?unknown",
	})
	expectShapes(t, got2, map[string]string{"neg84": "string[]{-5: string, -4: string}"})
}

// A destructuring element write (`[$w['y']] = …`) is an unknown write: the
// keys it may change and add are unknown.
func TestShapesDestructuringElemWrite(t *testing.T) {
	got := shapesAt(t, phpversion.PHP84, `<?php
function f() {
    $w = ['x' => 1];
    [$w['y']] = [2];
    t('key', $w['x']);
    foreach ($w as $k => $v) { t('foreachKey', $k); }
    t('badOctal', [09 => 1]);
}
`)
	expectShapes(t, got, map[string]string{
		"key":        "?unknown",
		"foreachKey": "int|string",
		"badOctal":   "non-empty int[]", // invalid integer literal: no key
	})
}
