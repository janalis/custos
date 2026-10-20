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
	"custos/internal/semantic/types"
)

// NarrowedFrom narrows a type assumed for a read by its guards, whatever the
// read's declared type: request data tested for its length, a member, a
// value carried around a loop, and locals the narrowing cannot follow. An
// untyped receiver's property may run code when read: no proof. A call of
// the project may write a superglobal; a copy carries the tests of its
// unchanged source.
func TestNarrowedFrom(t *testing.T) {
	src := `<?php
function f($o, $n) {
    $tag = $_POST['tag'];
    if (strlen($tag) !== 16) { return; }
    u($tag);
    if (16 === strlen($o->tag)) { u($o->tag); }
    if (mb_strlen($_GET['t'], '8bit') >= 4) { u($_GET['t']); }
    $s = $n; if (strlen($s) < 1) { return; } $s = $n; u($s);
    $r = $n; while (strlen($r) > 2) { u($r); $r = substr($r, 1); }
    $q = $n; $alias = &$q; if (strlen($q) === 3) { u($q); }
    if (strlen($$n) === 3) { u($$n); }
    if (strlen($s, 2) === 3) { u($s); }
    if (mb_strlen($s, 'UTF-8') === 3) { u($s); }
    if (mb_strlen($s, $n) === 3) { u($s); }
    if (strlen($s) === $n) { u($s); }
    if (u(1)) {}
    if (strlen($s) > 9223372036854775807) { u($s); }
    if (strlen($_POST['k']) === 2) { refresh(); u($_POST['k']); }
    if (strlen($_POST['k']) === 2) { strtolower('x'); u($_POST['k']); }
    $c = $_POST['c']; if (strlen($_POST['c']) === 5) { u($c); }
    $c = $_POST['c']; $_POST['c'] = $n; if (strlen($_POST['c']) === 5) { u($c); }
    $a = $n; $c = $a; $a = $o; if (strlen($a) === 5) { u($c); }
    for ($i = $n; strlen($i) > 1; $i = substr($i, 1)) { u($i); }
    if (strpos(haystack: $n, needle: 'a', offset: 1) !== false) { u($n); }
    $p = $n; extract($_GET); if (strlen($p) === 3) { u($p); }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && c.Name.(*syntax.Name).Value == "u" {
			got = append(got, env.NarrowedFrom(c.Args.Args[0].(*syntax.Arg).Value, types.String).ShapeString())
		}
		return true
	})
	want := []string{
		"string<len 16, 16>", "string", "string<len 4, max>", "string", "string<len 3, max>", "string<len 3, 3>",
		"string", "string", "string", "string", "string", "string", "?unknown",
		"string", "string<len 2, 2>", "string<len 5, 5>", "string", "string", "string<len 2, max>", "string", "?unknown",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// Once lengths are known, truthiness, comparisons with `”` and a byte read at
// an offset narrow them: a fread result seeded as false|string<len 0, 4>.
func TestKnownLengths(t *testing.T) {
	src := `<?php
function f($h) {
    $d = fread($h, 4); if ($d) { u($d); }
    $d = fread($h, 4); if (!$d) { u($d); }
    $d = fread($h, 4); if ($d !== '') { u($d); }
    $d = fread($h, 4); if ($d != '') { u($d); }
    $d = fread($h, 4); if ($d === '') { u($d); }
    $d = fread($h, 4); if ($d == '') { u($d); }
    $d = fread($h, 4); if (isset($d[3])) { u($d); }
    $d = fread($h, 4); if (isset($d[$h])) { u($d); }
    $d = fread($h, 4); if (strlen($d) >= 4) { u($d); }
    $d = fread($h, 4); if (strlen($d) !== 0) { u($d); }
    $d = fread($h, 4); if (strlen($d) != 2) { u($d); }
    $d = fread($h, 4); if (strlen($d) > 5) { u($d); }
    $d = fread($h, 4); if ($d === '') { if (strlen($d) === 4) { u($d); } }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	seed := types.Union(types.StringLengths(types.IntRange{Lo: 0, Hi: 4}), types.Of("false"))
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && c.Name.(*syntax.Name).Value == "u" {
			got = append(got, env.NarrowedFrom(c.Args.Args[0].(*syntax.Arg).Value, seed).ShapeString())
		}
		return true
	})
	want := []string{
		"string<len 1, 4>", "false|string<len 0, 1>", "false|string<len 1, 4>", "string<len 1, 4>",
		"string<len 0, 0>", "false|string<len 0, 0>", "string<len 4, 4>", "string<len 0, 4>",
		"string<len 4, 4>", "string<len 1, 4>", "false|string<len 0, 1>|string<len 3, 4>", "?unknown", "?unknown",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A needle found in a haystack proves the haystack was not null or false,
// which read as `”`.
func TestNeedleFound(t *testing.T) {
	src := `<?php
function f(?string $s, $n) {
    if (str_contains($s, 'a')) { u($s); }
    if (str_starts_with($s, '')) { u($s); }
    if (strpos($s, 'a') !== false) { u($s); }
    if (false !== stripos($s, 'a')) { u($s); }
    if (strpos($s, 'a') === false) { u($s); }
    if (strstr($s, 'a')) { u($s); }
    if (!str_ends_with($s, 'a')) { u($s); }
    if (str_contains($s, $n)) { u($s); }
    if (str_contains($n, 'a')) { u($s); }
    if (str_contains(haystack: $s, needle: 'a')) { u($s); }
    if (str_contains($s, 'a', 3)) { u($s); }
    if ($s[0] === 'a') { u($s); }
    if (str_contains($s, 'a', haystack: $s)) { u($s); }
    if (false === $i = strpos($s, 'a')) { return; } u($s);
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && c.Name.(*syntax.Name).Value == "u" {
			got = append(got, env.TypeOf(c.Args.Args[0].(*syntax.Arg).Value).String())
		}
		return true
	})
	want := []string{"string", "null|string", "string", "string", "null|string", "string", "null|string", "null|string", "null|string", "string", "string", "string", "null|string", "string"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A member copy whose source changed carries no test of it, and a local
// past the definition budget is unknown.
func TestNarrowedFromLimits(t *testing.T) {
	src := "<?php\nfinal class B { public $p; }\n" +
		"function f(B $o, $n) { $c = $o->p; $o->p = $n; if (strlen($o->p) === 5) { u($c); } }\n" +
		"function g($n, $k) { $v = $n;" + strings.Repeat(" if ($k > 1) { $v = $k; }", 513) + " u($v); }\n"
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && c.Name.(*syntax.Name).Value == "u" {
			got = append(got, env.NarrowedFrom(c.Args.Args[0].(*syntax.Arg).Value, types.String).ShapeString())
		}
		return true
	})
	if len(got) != 2 || got[0] != "string" || got[1] != "?unknown" {
		t.Fatalf("got %v", got)
	}
}
